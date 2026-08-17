package auth

import (
	"context"
	"file_share/internal/entity"
)

func (s *Service) Me(ctx context.Context, token string) (entity.User, error) {
	session, err := s.sessionRepository.GetSession(ctx, token)
	if err != nil {
		return entity.User{}, entity.ErrorAuthMe
	}

	user, err := s.userRepository.GetUserByLogin(ctx, session.Login)
	if err != nil {
		return entity.User{}, entity.ErrorAuthMe
	}

	return user, nil
}
