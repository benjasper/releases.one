package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/benjasper/releases.one/internal/github"
	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/pkg/keyedmutex"
	"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type SyncService struct {
	repository        *repository.Queries
	githubOAuthConfig *oauth2.Config
	repositoryMutex   *keyedmutex.KeyedMutex
	userMutex         *keyedmutex.KeyedMutex
}

func NewSyncService(repository *repository.Queries, githubOAuthConfig *oauth2.Config) *SyncService {
	return &SyncService{
		repository:        repository,
		githubOAuthConfig: githubOAuthConfig,
		repositoryMutex:   keyedmutex.NewKeyedMutex(),
		userMutex:         keyedmutex.NewKeyedMutex(),
	}
}

func (s *SyncService) SyncUser(ctx context.Context, user *repository.User) error {
	s.userMutex.Lock(user.Username)
	defer s.userMutex.Unlock(user.Username)

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
		UserID:    user.ID,
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

func (s *SyncService) syncRepositoriesAndReleases(ctx context.Context, user *repository.User, githubService *github.GitHubService) error {
	reposGroup, releasesCtx := errgroup.WithContext(ctx)

	reposGroup.Go(func() error {
		releasesErrGroup, ctx := errgroup.WithContext(ctx)
		releasesErrGroup.SetLimit(10)

		for repo, err := range githubService.GetStarredRepos(releasesCtx) {
			if err != nil {
				if errors.Is(err, context.Canceled) {
					slog.Error(fmt.Sprintf("Error syncing repositories (context canceled): %s", context.Cause(ctx)))
				}
				return err
			}

			releasesErrGroup.Go(func() error {
				return s.syncRepository(ctx, repo, user, repository.RepositoryStarTypeStar)
			})
		}

		err := releasesErrGroup.Wait()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Error(fmt.Sprintf("Error syncing repositories (context canceled): %s", context.Cause(ctx)))
			}

			return err
		}

		return nil
	})

	reposGroup.Go(func() error {
		releasesErrGroup, ctx := errgroup.WithContext(ctx)
		releasesErrGroup.SetLimit(10)

		for repo, err := range githubService.GetWatchingRepos(releasesCtx) {
			if err != nil {
				if errors.Is(err, context.Canceled) {
					slog.Error(fmt.Sprintf("Error syncing repositories (context canceled): %s", context.Cause(ctx)))
				}
				return err
			}

			releasesErrGroup.Go(func() error {
				return s.syncRepository(ctx, repo, user, repository.RepositoryStarTypeWatch)
			})
		}

		err := releasesErrGroup.Wait()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Error(fmt.Sprintf("Error syncing repositories (context canceled): %s", context.Cause(ctx)))
			}

			return err
		}

		return nil
	})

	reposErr := reposGroup.Wait()
	if reposErr != nil {
		if errors.Is(reposErr, context.Canceled) {
			slog.Error(fmt.Sprintf("Error syncing repositories (context canceled): %s", context.Cause(ctx)))
		}

		return reposErr
	}

	return nil
}

func (s *SyncService) syncRepository(ctx context.Context, repo *github.Repository, user *repository.User, starType repository.RepositoryStarType) error {
	// Lock the syncing of this repository by name
	s.repositoryMutex.Lock(repo.NameWithOwner)
	defer s.repositoryMutex.Unlock(repo.NameWithOwner)

	// Ignore private repositories for now
	if repo.IsPrivate {
		return nil
	}

	hash, err := hashstructure.Hash(repo, hashstructure.FormatV2, nil)
	if err != nil {
		return err
	}

	githubRepo, err := s.repository.GetRepositoryByGithubID(ctx, repo.ID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		slog.Info(fmt.Sprintf("No repository found, creating new repository: %s", repo.NameWithOwner))

		// openGraphImageSize, err := githubService.GetImageSize(ctx, repo.OpenGraphImageURL)
		// if err != nil {
		// 	return err
		// }

		err = s.repository.CreateRepository(ctx, repository.CreateRepositoryParams{
			GithubID:     repo.ID,
			Name:         repo.NameWithOwner,
			Url:          repo.URL,
			ImageUrl:     fmt.Sprintf("https://opengraph.githubassets.com/1/%s", repo.NameWithOwner),
			ImageSize:    0,
			Private:      repo.IsPrivate,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastSyncedAt: time.Now(),
			Hash:         hash,
		})
		if err != nil {
			return err
		}

		githubRepo, err = s.repository.GetRepositoryByGithubID(ctx, repo.ID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// If the repository was updated in the last hour, don't sync it
		if githubRepo.UpdatedAt.After(time.Now().Add(1 * time.Hour)) {
			return nil
		}

		// In case the image changed, refetch the image size
		// if repo.OpenGraphImageURL != githubRepo.ImageUrl {
		// 	openGraphImageSize, err := githubService.GetImageSize(ctx, repo.OpenGraphImageURL)
		// 	if err != nil {
		// 		return err
		// 	}
		//
		// 	githubRepo.ImageUrl = repo.OpenGraphImageURL
		// 	githubRepo.ImageSize = int32(openGraphImageSize)
		// }

		// Check hash
		if hash != githubRepo.Hash {
			githubRepo.Hash = hash
			githubRepo.UpdatedAt = time.Now()

			slog.Info(fmt.Sprintf("Repository hash changed, updating repository: %s", repo.NameWithOwner))
			_, err = s.repository.UpdateRepository(ctx, repository.UpdateRepositoryParams{
				ID:           githubRepo.ID,
				Url:          githubRepo.Url,
				ImageUrl:     fmt.Sprintf("https://opengraph.githubassets.com/1/%s", githubRepo.Name),
				ImageSize:    githubRepo.ImageSize,
				Private:      githubRepo.Private,
				CreatedAt:    githubRepo.CreatedAt,
				UpdatedAt:    githubRepo.UpdatedAt,
				LastSyncedAt: githubRepo.LastSyncedAt,
				Hash:         githubRepo.Hash,
			})
			if err != nil {
				return err
			}
		}
	}

	// Now check if the repository has already been starred by the user
	result, err := s.repository.UpdateRepositoryStar(ctx, repository.UpdateRepositoryStarParams{
		UpdatedAt:    time.Now(),
		RepositoryID: githubRepo.ID,
		UserID:       user.ID,
	})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		slog.Info(fmt.Sprintf("No repository star found, creating new repository star: %s", repo.NameWithOwner))
		err = s.repository.InsertRepositoryStar(ctx, repository.InsertRepositoryStarParams{
			RepositoryID: githubRepo.ID,
			UserID:       user.ID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Type:         int8(starType),
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
		var existingRelease *repository.Release
		existingReleaseIdx := slices.IndexFunc(releases, func(release repository.Release) bool {
			return release.TagName == ghRelease.TagName
		})

		if existingReleaseIdx >= 0 {
			existingRelease = &releases[existingReleaseIdx]
		}

		hash, err := hashstructure.Hash(ghRelease, hashstructure.FormatV2, nil)
		if err != nil {
			return err
		}

		if existingRelease == nil {
			slog.Info(fmt.Sprintf("Release not found, creating new release for user %s and repository %s: %s", user.Username, githubRepo.Name, ghRelease.TagName))
			author := ghRelease.Author.Name
			if author == "" {
				author = ghRelease.Author.Login
			}

			err = s.repository.InsertRelease(ctx, repository.InsertReleaseParams{
				GithubID:         ghRelease.ID,
				RepositoryID:     githubRepo.ID,
				Name:             ghRelease.Name,
				TagName:          ghRelease.TagName,
				Url:              ghRelease.URL,
				Description:      string(mdToHTML([]byte(ghRelease.Description))),
				DescriptionShort: ghRelease.ShortDescriptionHTML,
				Author:           sql.NullString{String: author, Valid: author != ""},
				ReleasedAt:       ghRelease.PublishedAt,
				IsPrerelease:     ghRelease.IsPrerelease,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				Hash:             hash,
			})
			if err != nil {
				return err
			}
		} else if hash != existingRelease.Hash {
			author := ghRelease.Author.Name
			if author == "" {
				author = ghRelease.Author.Login
			}

			slog.Info(fmt.Sprintf("Release hash changed (old: %d, new: %d), updating release: %s for repository %s", existingRelease.Hash, hash, ghRelease.Name, githubRepo.Name))
			_, err = s.repository.UpdateRelease(ctx, repository.UpdateReleaseParams{
				ID:               existingRelease.ID,
				GithubID:         ghRelease.ID,
				Name:             ghRelease.Name,
				Url:              ghRelease.URL,
				Description:      string(mdToHTML([]byte(ghRelease.Description))),
				DescriptionShort: ghRelease.ShortDescriptionHTML,
				Author:           sql.NullString{String: author, Valid: author != ""},
				ReleasedAt:       ghRelease.PublishedAt,
				IsPrerelease:     ghRelease.IsPrerelease,
				UpdatedAt:        time.Now(),
				Hash:             hash,
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
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
