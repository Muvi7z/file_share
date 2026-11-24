package auth

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"time"
)

type authService interface {
	Login(ctx context.Context, user entity.LoginUser) (entity.LoginResponse, error)
	Register(ctx context.Context, user entity.RegisterUser) (string, error)
	Me(ctx context.Context, token string) (entity.MeUser, error)
	RefreshTokens(ctx context.Context, sessionId, secret string) (entity.Token, error)
	Logout(ctx context.Context, sessionId string) error
}

type Handler struct {
	authService     authService
	logger          deps.Logger
	cookieSecure    bool
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewHandler(
	authService authService,
	logger deps.Logger,
	cookieSecure bool,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Handler {
	return &Handler{
		authService:     authService,
		logger:          logger,
		cookieSecure:    cookieSecure,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}
