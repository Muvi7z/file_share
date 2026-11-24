package auth

import (
	"context"
	"file_share/internal/entity"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, user entity.LoginUser) (entity.LoginResponse, error) {

	findUser, err := s.userRepository.GetUserByLogin(ctx, user.Login)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.LoginResponse{}, entity.ErrorInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.PasswordHash), []byte(user.Password))
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.LoginResponse{}, entity.ErrorInvalidCredentials
	}

	//_, err = s.tokenService.GenerateToken(findUser.Id, findUser.Role)
	//if err != nil {
	//	return entity.ErrorLoginUser
	//}

	id := uuid.New().String()
	secret := generateSecret()
	token := hashSecret(secret)

	session := entity.Session{
		Id:        id,
		Token:     token,
		Login:     user.Login,
		Role:      findUser.Role,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}

	err = s.sessionRepository.SetSession(ctx, id, session, s.cacheTTL)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.LoginResponse{}, entity.ErrorInvalidCredentials
	}

	generateToken, err := s.tokenService.GenerateToken(findUser.Id, findUser.Role, id)
	if err != nil {
		return entity.LoginResponse{}, entity.ErrorInvalidCredentials
	}

	meUser := entity.MeUser{
		Login: findUser.Login,
		Role:  findUser.Role,
	}

	return entity.LoginResponse{
		Token: entity.Token{
			AccessToken:  generateToken,
			RefreshToken: buildRefreshToken(id, secret),
		},
		User: meUser,
	}, nil
}
