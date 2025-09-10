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

	"golang.org/x/oauth2"
)

type GitHubService struct {
	client            *http.Client
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

	client := oauthConfig.Client(ctx, newToken)
	return &GitHubService{
		client:            client,
		githubOAuthConfig: oauthConfig,
	}, refreshedToken, nil
}

var pageSize = 25

func (s *GitHubService) GetStarredRepos(ctx context.Context) iter.Seq2[*Repository, error] {
	return func(yield func(*Repository, error) bool) {
		hasNextPage := true
		after := ""
		for hasNextPage {
			requestBody := make(map[string]string)
			requestBody["query"] = StarredReposQuery(pageSize, after)
			requestJson, err := json.Marshal(requestBody)
			if err != nil {
				yield(nil, err)
				return
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(requestJson))
			if err != nil {
				yield(nil, err)
				return
			}

			req.Header.Set("User-Agent", "releases.one")

			resp, err := s.client.Do(req)
			if err != nil {
				yield(nil, errors.Join(err, fmt.Errorf("failed to make starred repositories request to GitHub")))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				yield(nil, fmt.Errorf("unexpected response from GitHub, while fetching starred repos, status: %s", resp.Status))
				return
			}

			var starredReposResponse StarredReposResponse
			if err := json.NewDecoder(resp.Body).Decode(&starredReposResponse); err != nil {
				yield(nil, err)
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
			requestBody := make(map[string]string)
			requestBody["query"] = WatchingReposQuery(pageSize, after)
			requestJson, err := json.Marshal(requestBody)
			if err != nil {
				yield(nil, err)
				return
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(requestJson))
			if err != nil {
				yield(nil, err)
				return
			}

			req.Header.Set("User-Agent", "releases.one")

			resp, err := s.client.Do(req)
			if err != nil {
				yield(nil, errors.Join(err, fmt.Errorf("failed to make watching repositories request to GitHub")))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				yield(nil, fmt.Errorf("unexpected response from GitHub, while fetching watching repos, status: %s", resp.Status))
				return
			}

			var watchingReposResponse WatchingReposResponse
			if err := json.NewDecoder(resp.Body).Decode(&watchingReposResponse); err != nil {
				yield(nil, err)
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
