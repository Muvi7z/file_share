package auth

import (
	"context"
	"file_share/internal/entity"
	"github.com/google/uuid"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, user entity.LoginUser) (entity.Session, error) {

	findUser, err := s.userRepository.GetUserByLogin(ctx, user.Login)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Session{}, entity.ErrorGetUser
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.PasswordHash), []byte(user.Password))
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Session{}, entity.ErrorInvalidCredentials
	}

	//_, err = s.tokenService.GenerateToken(findUser.Id, findUser.Role)
	//if err != nil {
	//	return entity.ErrorLoginUser
	//}

	token := uuid.New().String()

	session := entity.Session{
		Token:     token,
		Login:     user.Login,
		Role:      findUser.Role,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}

	err = s.sessionRepository.SetSession(ctx, token, session, s.cacheTTL)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Session{}, entity.ErrorLoginUser
	}

	return session, nil
}
