package server

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/grpcreflect"
	"github.com/benjasper/releases.one/internal/config"
	"github.com/benjasper/releases.one/internal/gen/api/v1/apiv1connect"
	"github.com/benjasper/releases.one/internal/github"
	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/internal/server/services"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/gorilla/feeds"
	"github.com/olivere/vite"
	"github.com/rs/cors"
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
	config            *config.Config
	repository        *repository.Queries
	syncService       *services.SyncService
	githubOAuthConfig *oauth2.Config
	baseURL           *url.URL
	distFS            *fs.FS
	indexHTML         []byte
}

func NewServer(config *config.Config, repository *repository.Queries, githubOAuthConfig *oauth2.Config, baseURL *url.URL, distFS *fs.FS, indexHTML []byte) *Server {
	return &Server{
		config:            config,
		repository:        repository,
		githubOAuthConfig: githubOAuthConfig,
		syncService:       services.NewSyncService(repository, githubOAuthConfig),
		baseURL:           baseURL,
		distFS:            distFS,
		indexHTML:         indexHTML,
	}
}

// Start runs the server
func (s *Server) Start() {
	rpcServer := NewRpcServer(s.config, s.repository, s.syncService, s.baseURL)
	mux := http.NewServeMux()

	middleware := authn.NewMiddleware(func(ctx context.Context, req *http.Request) (any, error) {
		rawToken, _ := authn.BearerToken(req)

		if rawToken == "" {
			cookies := req.CookiesNamed("access_token")
			if len(cookies) == 0 {
				return nil, errors.New("missing authorization")
			}

			rawToken = cookies[0].Value
		}

		userID, err := validateAccessTokenClaims(rawToken, []byte(s.config.JWTSecret))
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

	indexHtml, err := s.CreateViteTemplate()
	if err != nil {
		panic(err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexHtml.Bytes())
	})
	mux.Handle("/assets/", http.FileServerFS(*s.distFS))

	s.ScheduleJobs()

	// TODO: Make origin configurable
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	corsHandledMux := corsConfig.Handler(mux)

	slog.Info("Starting server on port 80")
	http.ListenAndServe(":80", h2c.NewHandler(corsHandledMux, &http2.Server{}))
}

func (s *Server) CreateViteTemplate() (*bytes.Buffer, error) {
	viteConfig := vite.Config{
		FS:           os.DirFS("frontend/src"), // required: Vite build output
		IsDev:        true,                     // required: true or false
		ViteURL:      "http://localhost:3000",  // optional: defaults to this
		ViteEntry:    "src/index.tsx",          // optional: dependent on your frontend stack
		ViteTemplate: vite.SolidTs,
	}

	if s.config.IsProduction {
		viteConfig = vite.Config{
			FS:    *s.distFS, // required: Vite build output
			IsDev: false,     // required: true or false
		}
	}

	viteFragment, err := vite.HTMLFragment(viteConfig)
	if err != nil {
		panic(err)
	}

	indexHtml := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("name").Parse(string(s.indexHTML)))
	pageData := map[string]interface{}{
		"Vite": viteFragment,
	}

	err = t.Execute(indexHtml, pageData)
	if err != nil {
		panic(err)
	}

	return indexHtml, nil
}

func (s *Server) ScheduleJobs() {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}
	_, err = scheduler.NewJob(gocron.CronJob("*/5 * * * *", false), gocron.NewTask(func(s *Server) {
		interval := s.config.UserSyncInterval
		if interval == 0 {
			interval = 8
		}

		users, err := s.repository.GetUsersInNeedOfAnUpdate(context.Background(), time.Now().Add(time.Hour*-time.Duration(interval)))
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
		slog.Info("No user found, creating new user", "username", githubUser.Login)
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

	accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt, err := GenerateTokens(&user, []byte(s.config.JWTSecret))
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
		Domain:   s.baseURL.Hostname(),
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		Expires:  *accessTokenExpiresAt,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Domain:   s.baseURL.Hostname(),
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		Expires:  *refreshTokenExpiresAt,
	})

	loginSuccessRedirectURL := os.Getenv("LOGIN_SUCCESS_REDIRECT_URL")
	if loginSuccessRedirectURL == "" {
		loginSuccessRedirectURL = "/"
	}

	u, err := url.Parse(loginSuccessRedirectURL)
	if err != nil {
		http.Error(w, "Failed to parse login success redirect URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	query := u.Query()
	query.Add("access_token_expires_at", accessTokenExpiresAt.Format(time.RFC3339))
	u.RawQuery = query.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
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
