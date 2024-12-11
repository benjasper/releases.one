package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log"
	"net/http"
)

type GitHubService struct {
	client *http.Client
}

func NewGitHubService(client *http.Client) *GitHubService {
	return &GitHubService{
		client: client,
	}
}

var pageSize = 50

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

			var starredReposResponse StarredReposResponse
			if err := json.NewDecoder(resp.Body).Decode(&starredReposResponse); err != nil {
				yield(nil, err)
				return
			}

			hasNextPage = starredReposResponse.Data.Viewer.StarredRepositories.PageInfo.HasNextPage
			after = starredReposResponse.Data.Viewer.StarredRepositories.PageInfo.EndCursor

			if len(starredReposResponse.Errors) > 0 {
				for _, err := range starredReposResponse.Errors {
					log.Printf("Error: %s", err.Message)
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

func (s *GitHubService) GetUserData(ctx context.Context) (*UserData, error) {
	url := "https://api.github.com/user"
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch user data"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected response from GitHub")
	}

	var userData UserData
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, errors.Join(err, errors.New("failed to decode response"))
	}

	return &userData, nil
}
