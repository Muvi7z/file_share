package auth

import (
	"context"
	"errors"
	"file_share/internal/entity"
)

func (s *Service) Register(ctx context.Context, user entity.RegisterUser) (string, error) {
	return "", errors.New("not implemented")
}
