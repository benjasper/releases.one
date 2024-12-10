package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/benjasper/releases.one/pkg/github"
	"github.com/benjasper/releases.one/pkg/repository"
	"golang.org/x/oauth2"
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
	mux.HandleFunc("/github", s.GetLoginWithGithubCallback)

	log.Println("Starting server on port 80")
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

	client := s.githubOAuthConfig.Client(r.Context(), token)

	githubService := github.NewGitHubService(client)

	githubUser, err := githubService.GetUserData(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve user data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := s.repository.GetUserByUsername(r.Context(), githubUser.Login)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		log.Println("No user found, creating new user")
		_, err = s.repository.CreateUser(r.Context(), repository.CreateUserParams{
			Username:     githubUser.Login,
			RefreshToken: token.RefreshToken,
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

	log.Printf("Syncing user: %s", user.Username)
	log.Printf("Token: %s", token.AccessToken)

	err = s.syncUser(r.Context(), token, &user)
	if err != nil {
		http.Error(w, "Failed to sync user: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) syncUser(ctx context.Context, token *oauth2.Token, user *repository.User) error {
	client := s.githubOAuthConfig.Client(ctx, token)

	syncStartedAt := time.Now()

	githubService := github.NewGitHubService(client)

	response, err := githubService.GetStarredRepos(ctx)
	if err != nil {
		return err
	}

	for _, repo := range response.Data.Viewer.StarredRepositories.Nodes {
		githubRepo, err := s.repository.GetRepositoryByName(ctx, repo.NameWithOwner)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			log.Printf("No repository found, creating new repository: %s", repo.NameWithOwner)

			err = s.repository.CreateRepository(ctx, repository.CreateRepositoryParams{
				Name:         repo.NameWithOwner,
				Url:          repo.URL,
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
			log.Printf("No repository star found, creating new repository star: %s", repo.NameWithOwner)
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
				log.Printf("Release not found, creating new release: %s", ghRelease.TagName)
				err = s.repository.InsertRelease(ctx, repository.InsertReleaseParams{
					RepositoryID: githubRepo.ID,
					TagName:      ghRelease.TagName,
					Description:  ghRelease.DescriptionHTML,
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

		s.repository.DeleteLastXReleases(ctx, repository.DeleteLastXReleasesParams{
			Limit: 10,
			RepositoryID: githubRepo.ID,
		})
	}

	s.repository.DeleteRepositoryStarsUpdatedBefore(ctx, repository.DeleteRepositoryStarsUpdatedBeforeParams{
		UpdatedAt: syncStartedAt,
	})

	return nil
}
