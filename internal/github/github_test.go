package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestIsTokenInvalid(t *testing.T) {
	expiredWithoutRefreshToken := (&oauth2.Config{}).TokenSource(context.Background(), &oauth2.Token{
		AccessToken: "expired",
		Expiry:      time.Now().Add(-time.Hour),
	})
	_, noRefreshTokenErr := expiredWithoutRefreshToken.Token()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"unauthorized response", fmt.Errorf("fetch: %w", ErrUnauthorized), true},
		{"bad refresh token", &oauth2.RetrieveError{ErrorCode: "bad_refresh_token"}, true},
		{
			"bad refresh token from http client",
			errors.Join(&url.Error{Op: "Post", URL: "https://api.github.com/graphql", Err: &oauth2.RetrieveError{ErrorCode: "bad_refresh_token"}}),
			true,
		},
		{"expired without refresh token", noRefreshTokenErr, true},
		{"token endpoint down", &oauth2.RetrieveError{Response: &http.Response{StatusCode: http.StatusBadGateway}}, false},
		{"client misconfigured", &oauth2.RetrieveError{ErrorCode: "incorrect_client_credentials"}, false},
		{"timeout", context.DeadlineExceeded, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsTokenInvalid(test.err); got != test.want {
				t.Errorf("IsTokenInvalid(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
}
