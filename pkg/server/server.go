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
	"slices"
	"strconv"
	"time"

	"github.com/benjasper/releases.one/pkg/github"
	"github.com/benjasper/releases.one/pkg/repository"
	"github.com/go-co-op/gocron/v2"
	"github.com/gorilla/feeds"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type FeedType string

var (
	AtomFeedType FeedType = "application/atom+xml"
	RssFeedType  FeedType = "application/rss+xml"
)

type Server struct {
	repository        *repository.Queries
	githubOAuthConfig *oauth2.Config
}

func NewServer(repository *repository.Queries, githubOAuthConfig *oauth2.Config) *Server {
	return &Server{
		repository:        repository,
		githubOAuthConfig: githubOAuthConfig,
	}
}

// Start runs the server
func (s *Server) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/login/github", s.GetLoginWithGithub)
	mux.HandleFunc("/trigger/{username}", s.PostTriggerSync)
	mux.HandleFunc("/github", s.GetLoginWithGithubCallback)
	mux.HandleFunc("/atom/{username}", func(w http.ResponseWriter, r *http.Request) {
		s.GetFeed(w, r, AtomFeedType)
	})
	mux.HandleFunc("/rss/{username}", func(w http.ResponseWriter, r *http.Request) {
		s.GetFeed(w, r, RssFeedType)
	})

	scheduler, err := gocron.NewScheduler()
	_, err = scheduler.NewJob(gocron.DurationJob(time.Minute*5), gocron.NewTask(func(s *Server) {
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
			ctx, _ := context.WithTimeoutCause(context.Background(), time.Minute*5, errors.New("syncing user took too long"))
			err = s.syncUser(ctx, &user)
			if err != nil {
				slog.Info(fmt.Sprintf("Failed to sync user: %s", err.Error()))
			}
		}
	}, s))
	if err != nil {
		log.Fatal(err)
	}
	scheduler.Start()

	slog.Info("Starting server on port 80")
	http.ListenAndServe(":80", mux)
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

	user, err := s.repository.GetUserByUsername(r.Context(), githubUser.Login)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		slog.Info("No user found, creating new user")
		_, err = s.repository.CreateUser(r.Context(), repository.CreateUserParams{
			Username:     githubUser.Login,
			GithubToken:  repository.GitHubToken(*token),
			LastSyncedAt: time.Now(),
		})
		if err != nil {
			http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user, err = s.repository.GetUserByUsername(r.Context(), githubUser.Login)
		if err != nil {
			http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err == nil {
		err = s.repository.UpdateUserToken(r.Context(), repository.UpdateUserTokenParams{
			ID:          user.ID,
			GithubToken: repository.GitHubToken(*token),
		})
	}

	err = s.syncUser(r.Context(), &user)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to sync user: %s", err.Error()))
		http.Error(w, "Failed to sync user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) PostTriggerSync(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Must provide username", http.StatusBadRequest)
		return
	}

	user, err := s.repository.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "Failed to retrieve user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.syncUser(context.Background(), &user)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to sync user: %s", err.Error()))
		http.Error(w, "Failed to sync user: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) syncUser(ctx context.Context, user *repository.User) error {
	slog.Info(fmt.Sprintf("Syncing user: %s", user.Username))

	syncStartedAt := time.Now()

	githubService, newToken, err := github.NewGitHubService(ctx, s.githubOAuthConfig, (*oauth2.Token)(&user.GithubToken))
	if err != nil {
		return err
	}

	if newToken != nil {
		slog.Info(fmt.Sprintf("Saving refreshed token for user: %s", user.Username))
		err = s.repository.UpdateUserToken(ctx, repository.UpdateUserTokenParams{
			GithubToken: repository.GitHubToken(*newToken),
			ID:          user.ID,
		})
		if err != nil {
			return err
		}
	}

	err = s.syncRepositoriesAndReleases(ctx, user, githubService)
	if err != nil {
		return err
	}

	result, err := s.repository.DeleteRepositoryStarsUpdatedBefore(ctx, repository.DeleteRepositoryStarsUpdatedBeforeParams{
		UpdatedAt: syncStartedAt,
	})
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("Deleted %d repository stars for user: %s", rowsAffected, user.Username))

	err = s.repository.UpdateUserSyncedAt(ctx, repository.UpdateUserSyncedAtParams{
		ID:           user.ID,
		LastSyncedAt: syncStartedAt,
	})
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Synced user: %s", user.Username))

	return nil
}

func (s *Server) syncRepositoriesAndReleases(ctx context.Context, user *repository.User, githubService *github.GitHubService) error {
	group, ctx := errgroup.WithContext(ctx)

	for repo, err := range githubService.GetStarredRepos(ctx) {
		if err != nil {
			return err
		}
		group.Go(func() error {
			// Ignore private repositories for now
			if repo.IsPrivate {
				return nil
			}

			githubRepo, err := s.repository.GetRepositoryByName(ctx, repo.NameWithOwner)
			if err != nil && errors.Is(err, sql.ErrNoRows) {
				slog.Info(fmt.Sprintf("No repository found, creating new repository: %s", repo.NameWithOwner))

				openGraphImageSize, err := githubService.GetImageSize(ctx, repo.OpenGraphImageURL)
				if err != nil {
					return err
				}

				err = s.repository.CreateRepository(ctx, repository.CreateRepositoryParams{
					Name:         repo.NameWithOwner,
					Url:          repo.URL,
					ImageUrl:     repo.OpenGraphImageURL,
					ImageSize:    int32(openGraphImageSize),
					Private:      repo.IsPrivate,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
					LastSyncedAt: time.Now(),
				})
				if err != nil {
					return err
				}

				githubRepo, err = s.repository.GetRepositoryByName(ctx, repo.NameWithOwner)
				if err != nil {
					return err
				}
			}

			// Now check if the repository has already been starred by the user
			result, err := s.repository.UpdateRepositoryStar(ctx, repository.UpdateRepositoryStarParams{
				UpdatedAt:    time.Now(),
				RepositoryID: githubRepo.ID,
				UserID:       user.ID,
			})
			starRowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}

			if starRowsAffected == 0 {
				slog.Info(fmt.Sprintf("No repository star found, creating new repository star: %s", repo.NameWithOwner))
				err = s.repository.InsertRepositoryStar(ctx, repository.InsertRepositoryStarParams{
					RepositoryID: githubRepo.ID,
					UserID:       user.ID,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				})
				if err != nil {
					return err
				}
			}

			releases, err := s.repository.GetReleases(ctx, githubRepo.ID)
			if err != nil {
				return err
			}

			for _, ghRelease := range repo.Releases.Nodes {
				releaseExists := slices.ContainsFunc(releases, func(release repository.Release) bool {
					return release.TagName == ghRelease.TagName
				})

				if !releaseExists {
					slog.Info(fmt.Sprintf("Release not found, creating new release for user %s and repository %s: %s", user.Username, githubRepo.Name, ghRelease.TagName))
					author := ghRelease.Author.Name
					if author == "" {
						author = ghRelease.Author.Login
					}

					err = s.repository.InsertRelease(ctx, repository.InsertReleaseParams{
						RepositoryID: githubRepo.ID,
						Name:         ghRelease.Name,
						TagName:      ghRelease.TagName,
						Url:          ghRelease.URL,
						Description:  ghRelease.DescriptionHTML,
						Author:       sql.NullString{String: author },
						ReleasedAt:   ghRelease.PublishedAt,
						IsPrerelease: ghRelease.IsPrerelease,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					})
					if err != nil {
						return err
					}
				}
			}

			// Find the date of the 10th most recent release
			var oldestRelease *repository.Release
			if len(releases) > 10 {
				oldestRelease = &releases[len(releases)-10]

				result, err = s.repository.DeleteReleasesOlderThan(ctx, repository.DeleteReleasesOlderThanParams{
					ReleasedAt:   oldestRelease.ReleasedAt,
					RepositoryID: githubRepo.ID,
				})
				if err != nil {
					return err
				}

				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				slog.Info(fmt.Sprintf("Deleted %d releases older than %s for repository: %s", rowsAffected, oldestRelease.ReleasedAt.String(), repo.NameWithOwner))
			}

			return nil
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GetFeed(w http.ResponseWriter, r *http.Request, feedType FeedType) {
	username := r.PathValue("username")

	user, err := s.repository.GetUserByUsername(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
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
			Title:   fmt.Sprintf("%s: %s", release.RepositoryName.String, release.Name),
			Link:    &feeds.Link{Href: release.Url},
			Content: release.Description,
			Created: release.ReleasedAt,
		}

		if release.ImageUrl.Valid {
			feedItem.Enclosure = &feeds.Enclosure{
				Url:    release.ImageUrl.String,
				Type:   "image/png",
				Length: strconv.Itoa(int(release.ImageSize.Int32)),
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
