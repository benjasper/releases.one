package server

import (
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/benjasper/releases.one/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

var (
	AudienceAccess  = "access.releases.one"
	AudienceRefresh = "refresh.releases.one"
)

var Issuer = "releases.one"

func GenerateTokens(user *repository.User, signingKey []byte) (accessToken, refreshToken string, err error) {
	// Create the access token
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 2)),
		Issuer:    Issuer,
		Subject:   strconv.Itoa(int(user.ID)),
		Audience:  jwt.ClaimStrings{AudienceAccess},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString(signingKey)
	if err != nil {
		return "", "", errors.Join(err, errors.New("could not sign access token"))
	}

	// Create the refresh token
	refreshTokenClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 365)),
		Issuer:    Issuer,
		Subject:   strconv.Itoa(int(user.ID)),
		Audience:  jwt.ClaimStrings{AudienceRefresh},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString(signingKey)
	if err != nil {
		return "", "", errors.Join(err, errors.New("could not sign refresh token"))
	}

	return accessToken, refreshToken, nil
}

func parseToken(tokenString string, signingKey []byte) (*jwt.Token, *jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, nil, errors.Join(err, ErrInvalidToken)
	} else if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		return token, claims, nil
	} else {
		return nil, nil, errors.New("could not parse claims unknown claims type")
	}
}

func validateAccessTokenClaims(tokenString string, signingKey []byte) (int, error) {
	_, claims, err := parseToken(tokenString, signingKey)
	if err != nil {
		return 0, err
	}

	if claims.Issuer != Issuer {
		return 0, errors.New("invalid issuer")
	}

	if !slices.Contains(claims.Audience, AudienceAccess) {
		return 0, errors.New("invalid audience")
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, errors.Join(err, errors.New("invalid subject"))
	}

	return userID, nil
}

func validateRefreshTokenClaims(tokenString string, signingKey []byte) error {
	_, claims, err := parseToken(tokenString, signingKey)
	if err != nil {
		return err
	}

	if claims.Issuer != Issuer {
		return errors.New("invalid issuer")
	}

	if !slices.Contains(claims.Audience, AudienceRefresh) {
		return errors.New("invalid audience")
	}

	return nil
}
