package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/grpcreflect"
	"github.com/benjasper/releases.one/internal/gen/api/v1/apiv1connect"
	"github.com/benjasper/releases.one/internal/github"
	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/internal/server/services"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/gorilla/feeds"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/oauth2"
)

type FeedType string

var (
	AtomFeedType FeedType = "application/atom+xml"
	RssFeedType  FeedType = "application/rss+xml"
)

type Server struct {
	repository        *repository.Queries
	syncService       *services.SyncService
	githubOAuthConfig *oauth2.Config
	jwtSecret         []byte
}

func NewServer(repository *repository.Queries, githubOAuthConfig *oauth2.Config) *Server {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	return &Server{
		repository:        repository,
		githubOAuthConfig: githubOAuthConfig,
		syncService:       services.NewSyncService(repository, githubOAuthConfig),
		jwtSecret:         []byte(jwtSecret),
	}
}

// Start runs the server
func (s *Server) Start() {
	rpcServer := NewRpcServer(s.repository, s.syncService, s.jwtSecret)
	mux := http.NewServeMux()

	middleware := authn.NewMiddleware(func(ctx context.Context, req *http.Request) (any, error) {
		rawToken, ok := authn.BearerToken(req)
		if !ok {
			return nil, errors.New("missing authorization header")
		}

		userID, err := validateAccessTokenClaims(rawToken, s.jwtSecret)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid access token: %s", err.Error()))
			return nil, errors.New("invalid token")
		}

		return userID, nil
	})

	path, handler := apiv1connect.NewApiServiceHandler(rpcServer)
	mux.Handle(path, middleware.Wrap(handler))

	path, handler = apiv1connect.NewAuthServiceHandler(rpcServer)
	mux.Handle(path, handler)

	// TODO: Disable reflection when in production
	reflector := grpcreflect.NewStaticReflector(
		apiv1connect.ApiServiceName,
		apiv1connect.AuthServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	// Many tools still expect the older version of the server reflection API, so
	// most servers should mount both handlers.
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	mux.HandleFunc("/login/github", s.GetLoginWithGithub)
	mux.HandleFunc("/github", s.GetLoginWithGithubCallback)
	mux.HandleFunc("/atom/{userID}", func(w http.ResponseWriter, r *http.Request) {
		s.GetFeed(w, r, AtomFeedType)
	})
	mux.HandleFunc("/rss/{userID}", func(w http.ResponseWriter, r *http.Request) {
		s.GetFeed(w, r, RssFeedType)
	})

	s.ScheduleJobs()

	slog.Info("Starting server on port 80")
	http.ListenAndServe(":80", h2c.NewHandler(mux, &http2.Server{}))
}

func (s *Server) ScheduleJobs() {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}
	_, err = scheduler.NewJob(gocron.CronJob("*/5 * * * *", false), gocron.NewTask(func(s *Server) {
		interval := os.Getenv("USER_SYNC_INTERVAL")
		if interval == "" {
			interval = "8"
		}

		intervalInt, err := strconv.Atoi(interval)
		if err != nil {
			log.Fatal(err)
		}

		users, err := s.repository.GetUsersInNeedOfAnUpdate(context.Background(), time.Now().Add(time.Hour*-time.Duration(intervalInt)))
		if err != nil {
			log.Fatal(err)
		}
		slog.Info(fmt.Sprintf("Found %d user(s) in need of an update\n", len(users)))

		for _, user := range users {
			ctx, cancel := context.WithTimeoutCause(context.Background(), time.Minute*5, errors.New("syncing user took too long"))
			defer cancel()
			err = s.syncService.SyncUser(ctx, &user)
			if err != nil {
				slog.Info(fmt.Sprintf("Failed to sync user: %s", err.Error()))
				return
			}
		}
	}, s))
	if err != nil {
		log.Fatal(err)
	}
	scheduler.Start()
}

func (s *Server) GetLoginWithGithub(w http.ResponseWriter, r *http.Request) {
	url := s.githubOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Server) GetLoginWithGithubCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify state
	code := r.URL.Query().Get("code")

	token, err := s.githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	githubService, newToken, err := github.NewGitHubService(r.Context(), s.githubOAuthConfig, token)
	if err != nil {
		slog.Error(fmt.Sprintf("problem with token: %s", err))
		http.Error(w, "Problem with token", http.StatusInternalServerError)
	}

	if newToken != nil {
		log.Fatalf("this should not happen, because the token we receive should be new")
	}

	githubUser, err := githubService.GetUserData(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve user data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := s.repository.GetUserByGitHubID(r.Context(), githubUser.ID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		slog.Info("No user found, creating new user")
		_, err = s.repository.CreateUser(r.Context(), repository.CreateUserParams{
			GithubID:     githubUser.ID,
			Username:     githubUser.Login,
			GithubToken:  repository.GitHubToken(*token),
			LastSyncedAt: time.UnixMicro(0),
			IsPublic:     false,
			PublicID:     uuid.NewString(),
		})
		if err != nil {
			http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user, err = s.repository.GetUserByGitHubID(r.Context(), githubUser.ID)
		if err != nil {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.repository.UpdateUserToken(r.Context(), repository.UpdateUserTokenParams{
		ID:          user.ID,
		GithubToken: repository.GitHubToken(*token),
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to update user: %s", err.Error()))
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := GenerateTokens(&user, s.jwtSecret)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to generate tokens: %s", err.Error()))
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(time.Hour * 24 * 365),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(time.Hour * 24 * 365),
	})

	loginSuccessRedirectURL := os.Getenv("LOGIN_SUCCESS_REDIRECT_URL")
	if loginSuccessRedirectURL == "" {
		loginSuccessRedirectURL = "/"
	}

	http.Redirect(w, r, loginSuccessRedirectURL, http.StatusTemporaryRedirect)
}

func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request, feedType FeedType) {
	userID := r.PathValue("userID")

	user, err := s.repository.GetUserByPublicID(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

	if !user.IsPublic {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	releases, err := s.repository.GetReleasesForUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve releases: "+err.Error(), http.StatusInternalServerError)
		return
	}

	feed := &feeds.Feed{
		Title:       "GitHub Releases",
		Link:        &feeds.Link{Href: "https://releases.one"},
		Description: "A list of all the releases for all of your starred GitHub repositories",
		Updated:     user.LastSyncedAt,
	}

	for _, release := range releases {
		feedItem := &feeds.Item{
			Id:          fmt.Sprintf("releases.one-%s-%s", release.RepositoryGithubID.String, release.GithubID),
			Title:       fmt.Sprintf("%s: %s", release.RepositoryName.String, release.Name),
			Link:        &feeds.Link{Href: release.Url},
			Description: release.DescriptionShort,
			Content:     release.Description,
			Created:     release.ReleasedAt,
		}

		if release.ImageUrl.Valid {
			feedItem.Enclosure = &feeds.Enclosure{
				Url:  release.ImageUrl.String,
				Type: "image/png",
				// Length: strconv.Itoa(int(release.ImageSize.Int32)),
			}
		}

		if release.Author.Valid {
			feedItem.Author = &feeds.Author{Name: release.Author.String}
		}

		feed.Add(feedItem)
	}

	var responseBody string
	if feedType == RssFeedType {
		responseBody, err = feed.ToRss()
		if err != nil {
			http.Error(w, "Failed to convert feed to rss: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/rss+xml")
	} else {
		responseBody, err = feed.ToAtom()
		if err != nil {
			http.Error(w, "Failed to convert feed to atom: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/atom+xml")
	}

	w.Write([]byte(responseBody))
}
