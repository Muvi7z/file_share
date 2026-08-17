package auth

import (
	"context"
	"file_share/internal/entity"
	"file_share/internal/service/token"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
}

type tokenService interface {
	GenerateToken(userId string, userType entity.Role) (string, error)
	ValidateToken(tokenString string) (*token.Claims, error)
}

type sessionRepository interface {
	GetSession(ctx context.Context, token string) (entity.Session, error)
	CreateSession(ctx context.Context, key string, session entity.Session) error
	DeleteSession(ctx context.Context, token string) error
}

type Service struct {
	userRepository    UserRepository
	tokenService      tokenService
	sessionRepository sessionRepository
}

func NewService(userRepository UserRepository, tokenService tokenService, sessionRepository sessionRepository) *Service {
	return &Service{
		userRepository:    userRepository,
		tokenService:      tokenService,
		sessionRepository: sessionRepository,
	}
}
