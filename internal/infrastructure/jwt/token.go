package jwt

import (
	"errors"
	"file_share/internal/entity"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UserId    string `json:"userId"`
	Role      string `json:"role"`
	SessionId string `json:"sessionId"`
}

type Token struct {
	Key      []byte
	Duration time.Duration
}

func NewToken(key []byte, duration time.Duration) *Token {
	return &Token{
		Key:      key,
		Duration: duration,
	}
}

func (t *Token) GenerateToken(userId string, role entity.Role, sessionId string) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.Duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		UserId:    userId,
		Role:      string(role),
		SessionId: sessionId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.Key)
}

func (t *Token) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, entity.ErrInvalidToken
		}
		return t.Key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, entity.ErrExpiredToken
		}
		return nil, entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, entity.ErrInvalidToken
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, entity.ErrExpiredToken
	}

	return claims, nil

}
