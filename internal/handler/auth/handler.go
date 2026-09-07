package auth

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type authService interface {
	Login(ctx context.Context, user entity.LoginUser) (entity.Session, error)
	Me(ctx context.Context, token string) (entity.MeUser, error)
}

type Handler struct {
	authService authService
	logger      deps.Logger
}

func NewHandler(authService authService, logger deps.Logger) *Handler {
	return &Handler{authService: authService, logger: logger}
}
