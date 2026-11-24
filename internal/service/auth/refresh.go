package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"file_share/internal/entity"
	"fmt"
	"time"
)

func generateSecret() string {
	return rand.Text()

}

func hashSecret(secret string) string {

	token := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(token[:])
}

func buildRefreshToken(sessionId, secret string) string {
	return fmt.Sprintf("%s.%s", sessionId, secret)
}

func verifySecret(secret, expectedHash string) bool {
	actual := hashSecret(secret)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expectedHash)) == 1
}

func (s *Service) RefreshTokens(ctx context.Context, sessionId, secret string) (entity.Token, error) {
	session, err := s.sessionRepository.GetSession(ctx, sessionId)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Token{}, err
	}

	ok := verifySecret(secret, session.Token)
	if !ok {
		_ = s.sessionRepository.DeleteSession(ctx, sessionId)
		return entity.Token{}, entity.ErrInvalidToken
	}

	user, err := s.userRepository.GetUserByLogin(ctx, session.Login)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Token{}, err
	}

	secretNew := generateSecret()
	token := hashSecret(secretNew)

	newSession := entity.Session{
		Id:        sessionId,
		Token:     token,
		Login:     user.Login,
		Role:      user.Role,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}

	err = s.sessionRepository.SetSession(ctx, newSession.Id, newSession, s.cacheTTL)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Token{}, err
	}

	accessToken, err := s.tokenService.GenerateToken(user.Id, user.Role, newSession.Id)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.Token{}, err
	}

	return entity.Token{AccessToken: accessToken, RefreshToken: buildRefreshToken(newSession.Id, secretNew)}, nil
}
