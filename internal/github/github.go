package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
)

type GitHubService struct {
	client            *http.Client
	tokenSource       oauth2.TokenSource
	githubOAuthConfig *oauth2.Config
}

// NewGitHubService creates a new GitHubService, uses an existing token and refreshes it if necessary and returns the new token in case it was refreshed
func NewGitHubService(ctx context.Context, oauthConfig *oauth2.Config, token *oauth2.Token) (*GitHubService, *oauth2.Token, error) {
	tokenSource := oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, nil, err
	}

	var refreshedToken *oauth2.Token
	if token.AccessToken != newToken.AccessToken || token.RefreshToken != newToken.RefreshToken {
		refreshedToken = newToken
	}

	return &GitHubService{
		client:            oauth2.NewClient(ctx, tokenSource),
		tokenSource:       tokenSource,
		githubOAuthConfig: oauthConfig,
	}, refreshedToken, nil
}

// Token returns the current token, which differs from the initial one if the client refreshed it while in use
func (s *GitHubService) Token() (*oauth2.Token, error) {
	return s.tokenSource.Token()
}

var ErrUnauthorized = errors.New("GitHub rejected the token")

// IsTokenInvalid reports whether err means the user's token can no longer be used and they have to log in again
func IsTokenInvalid(err error) bool {
	if errors.Is(err, ErrUnauthorized) {
		return true
	}

	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		return retrieveErr.ErrorCode == "bad_refresh_token" || retrieveErr.ErrorCode == "invalid_grant"
	}

	// oauth2 doesn't export this error, it happens for expired tokens that were issued without a refresh token
	return strings.Contains(err.Error(), "token expired and refresh token is not set")
}

var pageSize = 25

func (s *GitHubService) GetStarredRepos(ctx context.Context) iter.Seq2[*Repository, error] {
	return func(yield func(*Repository, error) bool) {
		hasNextPage := true
		after := ""
		for hasNextPage {
			var starredReposResponse StarredReposResponse
			err := s.postGraphQL(ctx, map[string]any{"query": StarredReposQuery(pageSize, after)}, &starredReposResponse)
			if err != nil {
				yield(nil, errors.Join(err, fmt.Errorf("failed to fetch starred repositories")))
				return
			}

			hasNextPage = starredReposResponse.Data.Viewer.StarredRepositories.PageInfo.HasNextPage
			after = starredReposResponse.Data.Viewer.StarredRepositories.PageInfo.EndCursor

			if len(starredReposResponse.Errors) > 0 {
				for _, err := range starredReposResponse.Errors {
					slog.Info(fmt.Sprintf("Error: %s", err.Message))
				}
				yield(nil, errors.Join(errors.New("failed to fetch starred repos (graphql error)"), errors.New(starredReposResponse.Errors[0].Message)))
				return
			}

			if starredReposResponse.Message != "" {
				yield(nil, fmt.Errorf("failed to fetch starred repos(api error): %s", starredReposResponse.Message))
				return
			}

			for _, repo := range starredReposResponse.Data.Viewer.StarredRepositories.Nodes {
				if !yield(&repo, nil) {
					return
				}
			}
		}
	}
}

func (s *GitHubService) GetWatchingRepos(ctx context.Context) iter.Seq2[*Repository, error] {
	return func(yield func(*Repository, error) bool) {
		hasNextPage := true
		after := ""
		for hasNextPage {
			var watchingReposResponse WatchingReposResponse
			err := s.postGraphQL(ctx, map[string]any{"query": WatchingReposQuery(pageSize, after)}, &watchingReposResponse)
			if err != nil {
				yield(nil, errors.Join(err, fmt.Errorf("failed to fetch watching repositories")))
				return
			}

			hasNextPage = watchingReposResponse.Data.Viewer.Watching.PageInfo.HasNextPage
			after = watchingReposResponse.Data.Viewer.Watching.PageInfo.EndCursor

			if len(watchingReposResponse.Errors) > 0 {
				for _, err := range watchingReposResponse.Errors {
					slog.Info(fmt.Sprintf("Error: %s", err.Message))
				}
				yield(nil, errors.Join(errors.New("failed to fetch watching repos (graphql error)"), errors.New(watchingReposResponse.Errors[0].Message)))
				return
			}

			if watchingReposResponse.Message != "" {
				yield(nil, fmt.Errorf("failed to fetch watching repos(api error): %s", watchingReposResponse.Message))
				return
			}

			for _, repo := range watchingReposResponse.Data.Viewer.Watching.Nodes {
				if !yield(&repo, nil) {
					return
				}
			}
		}
	}
}

func (s *GitHubService) postGraphQL(ctx context.Context, payload any, out any) error {
	requestJson, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(requestJson))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "releases.one")

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Join(err, errors.New("failed to make GraphQL request to GitHub"))
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from GitHub GraphQL API, status: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (s *GitHubService) graphql(ctx context.Context, query string, variables map[string]any, out any) error {
	var response struct {
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Data json.RawMessage `json:"data"`
	}
	if err := s.postGraphQL(ctx, map[string]any{"query": query, "variables": variables}, &response); err != nil {
		return err
	}

	if len(response.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Message != "" {
		return fmt.Errorf("GitHub API error: %s", response.Message)
	}

	return json.Unmarshal(response.Data, out)
}

func (s *GitHubService) GetStarLists(ctx context.Context) ([]StarList, error) {
	var lists []StarList

	after := ""
	for {
		variables := map[string]any{"first": 10, "itemsFirst": 100, "after": nil}
		if after != "" {
			variables["after"] = after
		}

		var response StarListsResponse
		if err := s.graphql(ctx, StarListsQuery, variables, &response); err != nil {
			return nil, errors.Join(err, errors.New("failed to fetch star lists"))
		}

		for _, node := range response.Viewer.Lists.Nodes {
			list := StarList{
				ID:        node.ID,
				Name:      node.Name,
				Slug:      node.Slug,
				IsPrivate: node.IsPrivate,
			}

			items := node.Items
			for {
				for _, item := range items.Nodes {
					if item.ID != "" {
						list.RepositoryIDs = append(list.RepositoryIDs, item.ID)
					}
				}

				if !items.PageInfo.HasNextPage {
					break
				}

				var itemsResponse StarListItemsResponse
				err := s.graphql(ctx, StarListItemsQuery, map[string]any{
					"id":    node.ID,
					"first": 100,
					"after": items.PageInfo.EndCursor,
				}, &itemsResponse)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed to fetch items of star list %s", node.Slug))
				}
				items = itemsResponse.Node.Items
			}

			lists = append(lists, list)
		}

		if !response.Viewer.Lists.PageInfo.HasNextPage {
			return lists, nil
		}
		after = response.Viewer.Lists.PageInfo.EndCursor
	}
}

func (s *GitHubService) GetUserData(ctx context.Context) (*UserData, error) {
	url := "https://api.github.com/user"
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch user data"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from GitHub, while fetching user data: %s", resp.Status)
	}

	var userData UserData
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, errors.Join(err, errors.New("failed to decode response"))
	}

	return &userData, nil
}

func (s *GitHubService) GetImageSize(ctx context.Context, url string) (int, error) {
	resp, err := s.client.Head(url)
	if err != nil {
		return 0, errors.Join(err, errors.New("failed to fetch image"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected response from GitHub, while fetching image size: %s", resp.Status)
	}

	contentLengthHeader := resp.Header.Get("Content-Length")
	if contentLengthHeader == "" {
		return 0, errors.New("failed to fetch image size")
	}

	size, err := strconv.Atoi(contentLengthHeader)
	if err != nil {
		return 0, err
	}

	return size, nil
}
