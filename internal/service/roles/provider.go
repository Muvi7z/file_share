package roles

import (
	"context"
	"errors"
	"file_share/internal/entity"
)

type contextKey string

const (
	roleContextKey   contextKey = "user_role"
	userIDContextKey contextKey = "user_id"
)

var (
	ErrNoRoleInContext   = errors.New("no role found in context")
	ErrNoUserIDInContext = errors.New("no user id found in context")
)

type RolesProvider struct{}

func NewProvider() *RolesProvider {
	return &RolesProvider{}
}

func (p *RolesProvider) GetRole(ctx context.Context) (entity.Role, error) {
	role, ok := ctx.Value(roleContextKey).(entity.Role)
	if !ok {
		return "", ErrNoRoleInContext
	}

	return role, nil
}

func (p *RolesProvider) GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		return "", ErrNoUserIDInContext
	}

	return userID, nil
}

func SetRoleAndUserID(ctx context.Context, role entity.Role, userID string) context.Context {
	ctx = context.WithValue(ctx, roleContextKey, role)
	ctx = context.WithValue(ctx, userIDContextKey, userID)

	return ctx
}
