package auth

import (
	"context"
	"file_share/internal/entity"
)

func (s *Service) Me(ctx context.Context, token string) (entity.MeUser, error) {
	//session, err := s.sessionRepository.GetSession(ctx, token)
	//if err != nil {
	//	return entity.MeUser{}, entity.ErrorAuthMe
	//}

	claims, err := s.tokenService.ValidateToken(token)
	if err != nil {
		return entity.MeUser{}, err
	}

	user, err := s.userRepository.GetUserById(ctx, claims.UserId)
	if err != nil {
		return entity.MeUser{}, entity.ErrorAuthMe
	}

	res := entity.MeUser{
		Login: user.Login,
		Role:  user.Role,
	}

	return res, nil
}
