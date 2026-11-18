package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"slices"
	"strconv"
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

// SyncUser syncs the user unless they have been synced since staleBefore, re-checking after acquiring the user lock
// so that concurrent callers don't sync twice or use an outdated token
func (s *SyncService) SyncUser(ctx context.Context, userID int32, staleBefore time.Time) error {
	lockKey := strconv.Itoa(int(userID))
	s.userMutex.Lock(lockKey)
	defer s.userMutex.Unlock(lockKey)

	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.LastSyncedAt.Before(staleBefore) {
		return nil
	}

	slog.Info(fmt.Sprintf("Syncing user: %s", user.Username))

	syncStartedAt := time.Now()

	githubService, newToken, err := github.NewGitHubService(ctx, s.githubOAuthConfig, (*oauth2.Token)(&user.GithubToken))
	if err != nil {
		return err
	}

	savedToken := (*oauth2.Token)(&user.GithubToken)
	if newToken != nil {
		slog.Info(fmt.Sprintf("Saving refreshed token for user: %s", user.Username))
		err = s.repository.UpdateUserToken(ctx, repository.UpdateUserTokenParams{
			GithubToken: repository.GitHubToken(*newToken),
			ID:          user.ID,
		})
		if err != nil {
			return err
		}
		savedToken = newToken
	}

	// GitHub refresh tokens are single-use, so a refresh during the sync must be persisted even if the sync fails
	defer s.saveTokenIfRefreshed(context.WithoutCancel(ctx), &user, githubService, savedToken)

	err = s.syncRepositoriesAndReleases(ctx, &user, githubService)
	if err != nil {
		return err
	}

	// datetime columns have second precision, so rows written in the same second as syncStartedAt must survive
	cutoff := syncStartedAt.Truncate(time.Second)

	// Star lists are optional, a failure must not block the regular sync or wipe existing memberships
	if err := s.syncStarLists(ctx, &user, githubService, cutoff); err != nil {
		slog.Error(fmt.Sprintf("Failed to sync star lists for user %s: %s", user.Username, err.Error()))
	}

	result, err := s.repository.DeleteRepositoryStarsUpdatedBefore(ctx, repository.DeleteRepositoryStarsUpdatedBeforeParams{
		UpdatedAt: cutoff,
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

func (s *SyncService) saveTokenIfRefreshed(ctx context.Context, user *repository.User, githubService *github.GitHubService, savedToken *oauth2.Token) {
	token, err := githubService.Token()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get current token for user %s: %s", user.Username, err.Error()))
		return
	}

	if token.AccessToken == savedToken.AccessToken && token.RefreshToken == savedToken.RefreshToken {
		return
	}

	slog.Info(fmt.Sprintf("Saving token refreshed during sync for user: %s", user.Username))
	err = s.repository.UpdateUserToken(ctx, repository.UpdateUserTokenParams{
		GithubToken: repository.GitHubToken(*token),
		ID:          user.ID,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to save refreshed token for user %s: %s", user.Username, err.Error()))
	}
}

func (s *SyncService) syncRepositoriesAndReleases(ctx context.Context, user *repository.User, githubService *github.GitHubService) error {
	sources := []struct {
		repos    func(context.Context) iter.Seq2[*github.Repository, error]
		starType repository.RepositoryStarType
	}{
		{githubService.GetStarredRepos, repository.RepositoryStarTypeStar},
		{githubService.GetWatchingRepos, repository.RepositoryStarTypeWatch},
	}

	sourcesGroup, sourcesCtx := errgroup.WithContext(ctx)

	for _, source := range sources {
		sourcesGroup.Go(func() error {
			workers, workersCtx := errgroup.WithContext(sourcesCtx)
			workers.SetLimit(10)

			for repo, err := range source.repos(workersCtx) {
				if err != nil {
					// A failing worker cancels the fetch, so its error is the root cause
					if workersErr := workers.Wait(); workersErr != nil {
						return workersErr
					}
					return err
				}

				workers.Go(func() error {
					return s.syncRepository(workersCtx, repo, user, source.starType)
				})
			}

			return workers.Wait()
		})
	}

	err := sourcesGroup.Wait()
	if err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return errors.Join(err, cause)
		}
		return err
	}

	return nil
}

func (s *SyncService) syncStarLists(ctx context.Context, user *repository.User, githubService *github.GitHubService, cutoff time.Time) error {
	lists, err := githubService.GetStarLists(ctx)
	if err != nil {
		return err
	}

	for _, list := range lists {
		err = s.repository.UpsertStarList(ctx, repository.UpsertStarListParams{
			GithubID:  list.ID,
			UserID:    user.ID,
			Name:      list.Name,
			Slug:      list.Slug,
			IsPrivate: list.IsPrivate,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return err
		}

		starList, err := s.repository.GetStarListByGithubID(ctx, list.ID)
		if err != nil {
			return err
		}

		for _, repositoryGithubID := range list.RepositoryIDs {
			err = s.repository.UpsertStarListRepository(ctx, repository.UpsertStarListRepositoryParams{
				StarListID:         starList.ID,
				UpdatedAt:          time.Now(),
				RepositoryGithubID: repositoryGithubID,
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = s.repository.DeleteStarListRepositoriesUpdatedBefore(ctx, repository.DeleteStarListRepositoriesUpdatedBeforeParams{
		UpdatedAt: cutoff,
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	_, err = s.repository.DeleteStarListsUpdatedBefore(ctx, repository.DeleteStarListsUpdatedBeforeParams{
		UpdatedAt: cutoff,
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Synced %d star lists for user: %s", len(lists), user.Username))

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
		slog.Debug(fmt.Sprintf("No repository found, creating new repository: %s", repo.NameWithOwner))

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
			githubRepo.Name = repo.NameWithOwner
			githubRepo.Url = repo.URL
			githubRepo.ImageUrl = fmt.Sprintf("https://opengraph.githubassets.com/1/%s", repo.NameWithOwner)
			githubRepo.Private = repo.IsPrivate
			githubRepo.Hash = hash
			githubRepo.UpdatedAt = time.Now()

			slog.Debug(fmt.Sprintf("Repository hash changed, updating repository: %s", repo.NameWithOwner))
			_, err = s.repository.UpdateRepository(ctx, repository.UpdateRepositoryParams{
				ID:           githubRepo.ID,
				Name:         githubRepo.Name,
				Url:          githubRepo.Url,
				ImageUrl:     githubRepo.ImageUrl,
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

	err = s.repository.UpsertRepositoryStar(ctx, repository.UpsertRepositoryStarParams{
		RepositoryID: githubRepo.ID,
		UserID:       user.ID,
		Type:         int8(starType),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		return err
	}

	releases, err := s.repository.GetReleases(ctx, githubRepo.ID)
	if err != nil {
		return err
	}

	for _, ghRelease := range repo.Releases.Nodes {
		if ghRelease.IsDraft {
			continue
		}

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
			slog.Debug(fmt.Sprintf("Release not found, creating new release for user %s and repository %s: %s", user.Username, githubRepo.Name, ghRelease.TagName))
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
				if ctx.Err() != nil {
					return err
				}
				// A broken release is retried next sync, it must not block the rest of the user's sync
				slog.Error(fmt.Sprintf("Failed to insert release %s for repository %s: %s", ghRelease.TagName, githubRepo.Name, err.Error()))
				continue
			}
		} else if hash != existingRelease.Hash {
			author := ghRelease.Author.Name
			if author == "" {
				author = ghRelease.Author.Login
			}

			slog.Debug(fmt.Sprintf("Release hash changed (old: %d, new: %d), updating release: %s for repository %s", existingRelease.Hash, hash, ghRelease.Name, githubRepo.Name))
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
				if ctx.Err() != nil {
					return err
				}
				slog.Error(fmt.Sprintf("Failed to update release %s for repository %s: %s", ghRelease.TagName, githubRepo.Name, err.Error()))
				continue
			}
		}
	}

	// Re-fetch releases after syncing to get the updated count
	releases, err = s.repository.GetReleases(ctx, githubRepo.ID)
	if err != nil {
		return err
	}

	// Delete releases beyond the 10 most recent (releases are sorted newest-first)
	if len(releases) > 10 {
		// The 10th release is at index 9 (0-indexed), delete everything older than it
		tenthNewestRelease := &releases[9]

		result, err := s.repository.DeleteReleasesOlderThan(ctx, repository.DeleteReleasesOlderThanParams{
			ReleasedAt:   tenthNewestRelease.ReleasedAt,
			RepositoryID: githubRepo.ID,
		})
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		slog.Debug(fmt.Sprintf("Deleted %d releases older than %s for repository: %s", rowsAffected, tenthNewestRelease.ReleasedAt.String(), repo.NameWithOwner))
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
