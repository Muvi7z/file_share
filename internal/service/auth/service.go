package auth

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/infrastructure/jwt"
	"time"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (entity.User, error)
	GetUserById(ctx context.Context, id string) (entity.User, error)
}

type tokenService interface {
	GenerateToken(userId string, role entity.Role, sessionId string) (string, error)
	ValidateToken(tokenString string) (*jwt.Claims, error)
}

type sessionRepository interface {
	GetSession(ctx context.Context, id string) (entity.Session, error)
	SetSession(ctx context.Context, key string, session entity.Session, ttl time.Duration) error
	DeleteSession(ctx context.Context, token string) error
}

type Service struct {
	userRepository    UserRepository
	tokenService      tokenService
	sessionRepository sessionRepository
	cacheTTL          time.Duration
	logger            deps.Logger
}

func NewService(userRepository UserRepository, tokenService tokenService, sessionRepository sessionRepository, cacheTTL time.Duration, logger deps.Logger) *Service {
	return &Service{
		cacheTTL:          cacheTTL,
		userRepository:    userRepository,
		tokenService:      tokenService,
		sessionRepository: sessionRepository,
		logger:            logger,
	}
}
