package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

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
	githubUser, err := s.retrieveUserData(client)
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

	fmt.Fprintf(w, "User: %+v", user)

	retrieveStarredRepos(client, w)
}

type UserData struct {
	Login                   string    `json:"login"`
	ID                      int       `json:"id"`
	NodeID                  string    `json:"node_id"`
	AvatarURL               string    `json:"avatar_url"`
	GravatarID              string    `json:"gravatar_id"`
	URL                     string    `json:"url"`
	HTMLURL                 string    `json:"html_url"`
	FollowersURL            string    `json:"followers_url"`
	FollowingURL            string    `json:"following_url"`
	GistsURL                string    `json:"gists_url"`
	StarredURL              string    `json:"starred_url"`
	SubscriptionsURL        string    `json:"subscriptions_url"`
	OrganizationsURL        string    `json:"organizations_url"`
	ReposURL                string    `json:"repos_url"`
	EventsURL               string    `json:"events_url"`
	ReceivedEventsURL       string    `json:"received_events_url"`
	Type                    string    `json:"type"`
	SiteAdmin               bool      `json:"site_admin"`
	Name                    string    `json:"name"`
	Company                 string    `json:"company"`
	Blog                    string    `json:"blog"`
	Location                string    `json:"location"`
	Email                   string    `json:"email"`
	Hireable                bool      `json:"hireable"`
	Bio                     string    `json:"bio"`
	TwitterUsername         string    `json:"twitter_username"`
	PublicRepos             int       `json:"public_repos"`
	PublicGists             int       `json:"public_gists"`
	Followers               int       `json:"followers"`
	Following               int       `json:"following"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	PrivateGists            int       `json:"private_gists"`
	TotalPrivateRepos       int       `json:"total_private_repos"`
	OwnedPrivateRepos       int       `json:"owned_private_repos"`
	DiskUsage               int       `json:"disk_usage"`
	Collaborators           int       `json:"collaborators"`
	TwoFactorAuthentication bool      `json:"two_factor_authentication"`
	Plan                    struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		PrivateRepos  int    `json:"private_repos"`
		Collaborators int    `json:"collaborators"`
	} `json:"plan"`
}

func (s *Server) retrieveUserData(client *http.Client) (*UserData, error) {
	url := "https://api.github.com/user"
	resp, err := client.Get(url)
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

func retrieveStarredRepos(client *http.Client, w http.ResponseWriter) {
	url := "https://api.github.com/user/starred"
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch starred repos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Unexpected response from GitHub: "+resp.Status, resp.StatusCode)
		return
	}

	var starredRepos []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&starredRepos); err != nil {
		http.Error(w, "Failed to decode response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Display the starred repos
	fmt.Fprintf(w, "Starred Repositories:\n")
	for _, repo := range starredRepos {
		fmt.Fprintf(w, "- %s (URL: %s)\n", repo["name"], repo["html_url"])
	}
}
