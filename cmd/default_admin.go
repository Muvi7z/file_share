package cmd

import (
	"context"
	"database/sql"
	"errors"
	"file_share/internal/entity"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (c *Container) EnsureDefaultAdmin(ctx context.Context) error {
	login := c.configuration.GetDefaultAdminLogin()
	password := c.configuration.GetDefaultAdminPassword()

	_, err := c.GetRepository().GetUserByLogin(ctx, login)
	if err == nil {
		c.GetLogger().Info(ctx, fmt.Sprintf("default admin %q already exists", login))
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking default admin: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing default admin password: %w", err)
	}

	now := time.Now()
	_, err = c.GetRepository().CreateUser(ctx, entity.User{
		Id:           uuid.New().String(),
		Login:        login,
		PasswordHash: string(passwordHash),
		Role:         entity.RoleAdmin,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return fmt.Errorf("error creating default admin: %w", err)
	}

	c.GetLogger().Info(ctx, fmt.Sprintf("default admin %q created", login))
	return nil
}
