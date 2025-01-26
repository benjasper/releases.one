package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/benjasper/releases.one/internal/config"
	apiv1 "github.com/benjasper/releases.one/internal/gen/api/v1"
	"github.com/benjasper/releases.one/internal/repository"
	"github.com/benjasper/releases.one/internal/server/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RpcServer struct {
	config      *config.Config
	repository  *repository.Queries
	syncService *services.SyncService
	baseURL     *url.URL
}

func NewRpcServer(config *config.Config, repository *repository.Queries, syncService *services.SyncService, baseURL *url.URL) *RpcServer {
	return &RpcServer{
		config:      config,
		repository:  repository,
		syncService: syncService,
		baseURL:     baseURL,
	}
}

func (s *RpcServer) Sync(ctx context.Context, req *connect.Request[apiv1.SyncRequest]) (*connect.Response[apiv1.SyncResponse], error) {
	contextUserID := authn.GetInfo(ctx)
	if contextUserID == nil {
		return nil, errors.New("no user id in context")
	}

	userID, ok := contextUserID.(int)
	if !ok {
		return nil, errors.New("invalid user id in context")
	}

	if userID == 0 {
		return nil, errors.New("must provide user id")
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve user"))
	}

	// Only sync if it's been more than 24 hours since the last sync
	if user.LastSyncedAt.Before(time.Now().Add(-1 * time.Hour * 24)) {
		err = s.syncService.SyncUser(ctx, &user)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to sync user: %s", err.Error()))
			return nil, errors.Join(err, errors.New("failed to sync user"))
		}
	}

	releases, err := s.repository.GetReleasesForUserShortDescription(ctx, repository.GetReleasesForUserShortDescriptionParams{
		UserID:       user.ID,
		IsPrerelease: sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve releases"))
	}

	repositories, err := s.repository.FindRepositoriesByUser(ctx, user.ID)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve repositories"))
	}

	res := connect.NewResponse(&apiv1.SyncResponse{})

	res.Msg.RepositoryCount = int32(len(repositories))

	for _, release := range releases {
		res.Msg.Timeline = append(res.Msg.Timeline, &apiv1.TimelineEntry{
			Id:             release.ID,
			RepositoryId:   release.RepositoryID,
			Name:           release.Name,
			Url:            release.Url,
			TagName:        release.TagName,
			Description:    release.DescriptionShort,
			Author:         release.Author.String,
			IsPrerelease:   release.IsPrerelease,
			ReleasedAt:     timestamppb.New(release.ReleasedAt),
			RepositoryName: release.RepositoryName.String,
			ImageUrl:       release.ImageUrl.String,
			StarType:       apiv1.RepositoryStarType(release.RepositoryStarType),
		})
	}

	return res, nil
}

func (s *RpcServer) GetRepositories(ctx context.Context, req *connect.Request[apiv1.GetRepositoriesRequest]) (*connect.Response[apiv1.GetRepositoriesResponse], error) {
	res := connect.NewResponse(&apiv1.GetRepositoriesResponse{})

	contextUserID := authn.GetInfo(ctx)
	if contextUserID == nil {
		return nil, errors.New("no user id in context")
	}

	userID, ok := contextUserID.(int)
	if !ok {
		return nil, errors.New("invalid user id in context")
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve user"))
	}

	optionalPrerelease := sql.NullBool{Bool: req.Msg.Prerelease, Valid: !req.Msg.Prerelease}

	optionalStarType := sql.NullInt16{Int16: 0, Valid: false}
	if req.Msg.StarType != nil {
		optionalStarType = sql.NullInt16{Int16: int16(*req.Msg.StarType), Valid: true}
	}

	releases, err := s.repository.GetReleasesForUserShortDescription(ctx, repository.GetReleasesForUserShortDescriptionParams{
		UserID:       user.ID,
		IsPrerelease: optionalPrerelease,
		StarType:     optionalStarType,
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve releases"))
	}

	for _, release := range releases {
		res.Msg.Timeline = append(res.Msg.Timeline, &apiv1.TimelineEntry{
			Id:             release.ID,
			RepositoryId:   release.RepositoryID,
			Name:           release.Name,
			Url:            release.Url,
			TagName:        release.TagName,
			Description:    release.DescriptionShort,
			Author:         release.Author.String,
			IsPrerelease:   release.IsPrerelease,
			ReleasedAt:     timestamppb.New(release.ReleasedAt),
			RepositoryName: release.RepositoryName.String,
			RepositoryUrl:  release.RepositoryUrl.String,
			ImageUrl:       release.ImageUrl.String,
			StarType:       apiv1.RepositoryStarType(release.RepositoryStarType),
		})
	}

	return res, nil
}

func (s *RpcServer) ToogleUserPublicFeed(ctx context.Context, req *connect.Request[apiv1.ToogleUserPublicFeedRequest]) (*connect.Response[apiv1.ToogleUserPublicFeedResponse], error) {
	userIDAny := authn.GetInfo(ctx)
	if userIDAny == nil {
		return nil, errors.New("no user id in context")
	}

	userID, ok := userIDAny.(int)
	if !ok {
		return nil, errors.New("invalid user id in context")
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, errors.Join(err, errors.New("user not found"))
	}

	err = s.repository.UpdateUserIsPublic(ctx, repository.UpdateUserIsPublicParams{
		IsPublic: req.Msg.Enabled,
		ID:       user.ID,
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to update user"))
	}

	return connect.NewResponse(&apiv1.ToogleUserPublicFeedResponse{
		PublicId: user.PublicID,
	}), nil
}

func (s *RpcServer) RefreshToken(ctx context.Context, req *connect.Request[apiv1.RefreshTokenRequest]) (*connect.Response[apiv1.RefreshTokenResponse], error) {
	httpReq := http.Request{Header: req.Header()}
	cookies := httpReq.CookiesNamed("refresh_token")
	if len(cookies) == 0 {
		return nil, errors.New("missing refresh token cookie")
	}

	refreshToken := cookies[0].Value

	userID, err := validateRefreshTokenClaims(refreshToken, []byte(s.config.JWTSecret))
	if err != nil {
		return nil, errors.Join(err, errors.New("invalid refresh token"))
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve user"))
	}

	accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt, err := GenerateTokens(&user, []byte(s.config.JWTSecret))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to generate tokens"))
	}

	res := connect.NewResponse(&apiv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccessTokenExpiresAt: &timestamppb.Timestamp{
			Seconds: accessTokenExpiresAt.Unix(),
		},
		RefreshTokenExpiresAt: &timestamppb.Timestamp{
			Seconds: refreshTokenExpiresAt.Unix(),
		},
	})

	accessTokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		Domain:   s.baseURL.Hostname(),
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		Expires:  *accessTokenExpiresAt,
	}

	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Domain:   s.baseURL.Hostname(),
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  *refreshTokenExpiresAt,
	}

	res.Header().Add("Set-Cookie", accessTokenCookie.String())
	res.Header().Add("Set-Cookie", refreshTokenCookie.String())

	return res, nil
}

func (s *RpcServer) Logout(ctx context.Context, req *connect.Request[apiv1.LogoutRequest]) (*connect.Response[apiv1.LogoutResponse], error) {
	userID := authn.GetInfo(ctx)
	if userID == nil {
		return nil, errors.New("no user id in context")
	}

	res := connect.NewResponse(&apiv1.LogoutResponse{})

	accessTokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Domain:   s.baseURL.Hostname(),
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	}

	refreshTokenCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Domain:   s.baseURL.Hostname(),
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Unix(0, 0),
	}

	res.Header().Add("Set-Cookie", accessTokenCookie.String())
	res.Header().Add("Set-Cookie", refreshTokenCookie.String())

	return res, nil
}

func (s *RpcServer) GetMyUser(ctx context.Context, req *connect.Request[apiv1.GetMyUserRequest]) (*connect.Response[apiv1.GetMyUserResponse], error) {
	userID := authn.GetInfo(ctx)
	if userID == nil {
		return nil, errors.New("no user id in context")
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID.(int)))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve user"))
	}

	res := connect.NewResponse(&apiv1.GetMyUserResponse{
		Id:           user.ID,
		LastSyncedAt: timestamppb.New(user.LastSyncedAt),
		IsPublic:     user.IsPublic,
		PublicId:     user.PublicID,
		Name:         user.Username,
		IsOnboarded:  user.IsOnboarded,
	})

	return res, nil
}

func (s *RpcServer) ToggleUserOnboarded(ctx context.Context, req *connect.Request[apiv1.ToggleUserOnboardedRequest]) (*connect.Response[apiv1.ToggleUserOnboardedResponse], error) {
	userIDAny := authn.GetInfo(ctx)
	if userIDAny == nil {
		return nil, errors.New("no user id in context")
	}

	userID, ok := userIDAny.(int)
	if !ok {
		return nil, errors.New("invalid user id in context")
	}

	user, err := s.repository.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to retrieve user"))
	}

	if user.IsOnboarded {
		return nil, errors.New("user is already onboarded")
	}

	err = s.repository.UpdateUserOnboarded(ctx, repository.UpdateUserOnboardedParams{
		IsOnboarded: true,
		ID:          user.ID,
	})
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to update user"))
	}

	return connect.NewResponse(&apiv1.ToggleUserOnboardedResponse{}), nil
}
