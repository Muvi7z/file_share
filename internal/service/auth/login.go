package auth

import (
	"context"
	"file_share/internal/entity"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) Login(ctx context.Context, user entity.LoginUser) error {
	//Поиск по логину
	//проверка пароля
	//Создать токен
	//создать и вернуть сессия
	findUser, err := s.userRepository.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return entity.ErrorGetUser
	}

	err = bcrypt.CompareHashAndPassword([]byte(findUser.PasswordHash), []byte(user.Password))
	if err != nil {
		return entity.ErrorInvalidCredentials
	}

	_, err = s.tokenService.GenerateToken(findUser.Id, findUser.Role)
	if err != nil {
		return entity.ErrorLoginUser
	}

	return nil
}
