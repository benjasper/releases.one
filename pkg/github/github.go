package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (s *GitHubService) GetStarredRepos(ctx context.Context) (*StarredReposResponse, error) {
	requestBody := make(map[string]string)
	requestBody["query"] = StarredReposQuery
	requestJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "releases.one")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var starredReposResponse StarredReposResponse
	if err := json.NewDecoder(resp.Body).Decode(&starredReposResponse); err != nil {
		return nil, err
	}

	if len(starredReposResponse.Errors) > 0 {
		for _, err := range starredReposResponse.Errors {
			log.Printf("Error: %s", err.Message)
		}
		return nil, errors.Join(errors.New("failed to fetch starred repos (graphql error)"), errors.New(starredReposResponse.Errors[0].Message))
	}

	if starredReposResponse.Message != "" {
		return nil, fmt.Errorf("failed to fetch starred repos(api error): %s", starredReposResponse.Message)
	}

	return &starredReposResponse, nil
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
