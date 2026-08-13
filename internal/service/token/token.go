package token

import (
	"errors"
	"file_share/internal/entity"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserId   string      `json:"userId"`
	UserType entity.Role `json:"userType"`
	jwt.RegisteredClaims
}

type Service struct {
	Key           []byte
	tokenDuration time.Duration
}

func New(key []byte) *Service {
	return &Service{
		Key:           key,
		tokenDuration: 24 * time.Hour,
	}
}

func (s *Service) GenerateToken(userId string, userType entity.Role) (string, error) {
	claims := Claims{
		UserId:   userId,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.Key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return s.Key, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
