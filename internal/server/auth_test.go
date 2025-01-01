package server

import (
	"testing"

	"github.com/benjasper/releases.one/internal/repository"
)

func TestGenerateTokens(t *testing.T) {
	user := &repository.User{
		ID: 1,
	}

	accessToken, refreshToken, _, _, err := GenerateTokens(user, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	if accessToken == "" {
		t.Fatal("access token is empty")
	}

	if refreshToken == "" {
		t.Fatal("refresh token is empty")
	}

	userID, err := validateAccessTokenClaims(accessToken, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	if userID != 1 {
		t.Fatal("user id does not match")
	}

	err = validateRefreshTokenClaims(refreshToken, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseTokenSignatureInvalid(t *testing.T) {
	invalidSignatureToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJyZWxlYXNlcy5vbmUiLCJzdWIiOiIxIiwiYXVkIjpbImFjY2Vzcy5yZWxlYXNlcy5vbmUiXSwiZXhwIjoxNzM1NTQ3ODM5fQ.cIn1voK3wc2KHF9i33TcN-3J"

	_, _, err := parseToken(invalidSignatureToken, []byte("secret"))
	if err == nil {
		t.Fatal("expected error when signature is invalid")
	}
}

func TestParseTokenExpired(t *testing.T) {
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJyZWxlYXNlcy5vbmUiLCJzdWIiOiIxIiwiYXVkIjpbImFjY2Vzcy5yZWxlYXNlcy5vbmUiXSwiZXhwIjowfQ.uh325BJTvUnIvZ1X1F7NjiET13OM4tYBhA9X0jq20U8"

	_, _, err := parseToken(expiredToken, []byte("secret"))
	if err == nil {
		t.Fatal("expected error when token is expired")
	}
}

