package roles

import (
	"context"
	"errors"
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

func (p *RolesProvider) GetRole(ctx context.Context)
