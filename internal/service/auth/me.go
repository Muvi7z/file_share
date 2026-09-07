package auth

import (
	"context"
	"file_share/internal/entity"
)

func (s *Service) Me(ctx context.Context, token string) (entity.MeUser, error) {
	session, err := s.sessionRepository.GetSession(ctx, token)
	if err != nil {
		return entity.MeUser{}, entity.ErrorAuthMe
	}

	user, err := s.userRepository.GetUserByLogin(ctx, session.Login)
	if err != nil {
		return entity.MeUser{}, entity.ErrorAuthMe
	}

	res := entity.MeUser{
		Token:     token,
		Login:     user.Login,
		Role:      user.Role,
		ExpiresAt: session.ExpiresAt,
	}

	return res, nil
}
