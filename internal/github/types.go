package github

import (
	"fmt"
	"time"
)

var StarredReposQueryTemplate = `
query {
  rateLimit {
    limit
    cost
    remaining
    resetAt
  }
  viewer {
    starredRepositories(first: %d, after: "%s") {
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
		id
        nameWithOwner
	    url
		openGraphImageUrl
	    isPrivate
        releases(first: 3, orderBy: {field: CREATED_AT, direction: DESC}) {
          nodes {
			id
            name
            tagName
            isDraft
            isPrerelease
            publishedAt
            url
			description
			shortDescriptionHTML
			author {
				name
				login
			}
          }
        }
      }
    }
  }
}
`

func StarredReposQuery(first int, after string) string {
	return fmt.Sprintf(StarredReposQueryTemplate, first, after)
}

type StarredReposResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
	Data struct {
		Viewer struct {
			StarredRepositories struct {
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
				Nodes []Repository `json:"nodes"`
			} `json:"starredRepositories"`
		} `json:"viewer"`
		RateLimit struct {
			ResetAt   time.Time `json:"resetAt"`
			Limit     int       `json:"limit"`
			Cost      int       `json:"cost"`
			Remaining int       `json:"remaining"`
		} `json:"rateLimit"`
	} `json:"data"`
}

type Repository struct {
	ID                string `json:"id"`
	NameWithOwner     string `json:"nameWithOwner"`
	URL               string `json:"url"`
	OpenGraphImageURL string `json:"openGraphImageUrl"`
	Releases          struct {
		Nodes []struct {
			ID          string    `json:"id"`
			PublishedAt time.Time `json:"publishedAt"`
			Author      struct {
				Name  string `json:"name"`
				Login string `json:"login"`
			} `json:"author"`
			Name                 string `json:"name"`
			URL                  string `json:"url"`
			TagName              string `json:"tagName"`
			Description          string `json:"description"`
			ShortDescriptionHTML string `json:"shortDescriptionHTML"`
			IsDraft              bool   `json:"isDraft"`
			IsPrerelease         bool   `json:"isPrerelease"`
		} `json:"nodes"`
	} `json:"releases" hash:"ignore"`
	IsPrivate bool `json:"isPrivate"`
}

type UserData struct {
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Name              string    `json:"name"`
	NodeID            string    `json:"node_id"`
	GravatarID        string    `json:"gravatar_id"`
	Blog              string    `json:"blog"`
	HTMLURL           string    `json:"html_url"`
	FollowersURL      string    `json:"followers_url"`
	FollowingURL      string    `json:"following_url"`
	GistsURL          string    `json:"gists_url"`
	StarredURL        string    `json:"starred_url"`
	SubscriptionsURL  string    `json:"subscriptions_url"`
	OrganizationsURL  string    `json:"organizations_url"`
	ReposURL          string    `json:"repos_url"`
	EventsURL         string    `json:"events_url"`
	ReceivedEventsURL string    `json:"received_events_url"`
	Type              string    `json:"type"`
	Location          string    `json:"location"`
	Login             string    `json:"login"`
	Company           string    `json:"company"`
	URL               string    `json:"url"`
	AvatarURL         string    `json:"avatar_url"`
	TwitterUsername   string    `json:"twitter_username"`
	Email             string    `json:"email"`
	Bio               string    `json:"bio"`
	Plan              struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		PrivateRepos  int    `json:"private_repos"`
		Collaborators int    `json:"collaborators"`
	} `json:"plan"`
	PrivateGists            int    `json:"private_gists"`
	PublicGists             int    `json:"public_gists"`
	Followers               int    `json:"followers"`
	Following               int    `json:"following"`
	ID                      uint64 `json:"id"`
	TotalPrivateRepos       int    `json:"total_private_repos"`
	OwnedPrivateRepos       int    `json:"owned_private_repos"`
	DiskUsage               int    `json:"disk_usage"`
	Collaborators           int    `json:"collaborators"`
	PublicRepos             int    `json:"public_repos"`
	Hireable                bool   `json:"hireable"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	SiteAdmin               bool   `json:"site_admin"`
}
