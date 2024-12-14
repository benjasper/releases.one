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
        nameWithOwner
	    url
		openGraphImageUrl
	    isPrivate
        releases(first: 3, orderBy: {field: CREATED_AT, direction: DESC}) {
          nodes {
            name
            tagName
            isDraft
            isPrerelease
            publishedAt
            url
			descriptionHTML
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
		RateLimit struct {
			Limit     int       `json:"limit"`
			Cost      int       `json:"cost"`
			Remaining int       `json:"remaining"`
			ResetAt   time.Time `json:"resetAt"`
		} `json:"rateLimit"`
		Viewer struct {
			StarredRepositories struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []Repository `json:"nodes"`
			} `json:"starredRepositories"`
		} `json:"viewer"`
	} `json:"data"`
}

type Repository struct {
	NameWithOwner     string `json:"nameWithOwner"`
	URL               string `json:"url"`
	OpenGraphImageURL string `json:"openGraphImageUrl"`
	IsPrivate         bool   `json:"isPrivate"`
	Releases          struct {
		Nodes []struct {
			Name            string    `json:"name"`
			URL             string    `json:"url"`
			IsDraft         bool      `json:"isDraft"`
			IsPrerelease    bool      `json:"isPrerelease"`
			PublishedAt     time.Time `json:"publishedAt"`
			TagName         string    `json:"tagName"`
			DescriptionHTML string    `json:"descriptionHTML"`
			Author          struct {
				Name string `json:"name"`
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"releases"`
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
