package repository

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var userTable = "user"

type userRow struct {
	Id           string    `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r *Repository) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	var res entity.User
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createUserTx(ctx, tx, user)
		if err != nil {
			return err
		}

		return nil

	})

	if txErr != nil {
		return entity.User{}, txErr
	}

	return res, nil
}

func (r *Repository) createUserTx(ctx context.Context, tx *sqlx.Tx, user entity.User) (entity.User, error) {
	insertMap := map[string]any{
		"id":            user.Id,
		"login":         user.Login,
		"password_hash": user.PasswordHash,
		"role":          user.Role,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	sql, args, err := r.qb.Insert(userTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("error building query: %w", err)
	}

	var row userRow
	var res entity.User

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.User{}, fmt.Errorf("error executing query: %w", err)
	}

	res = entity.User{
		Id:           row.Id,
		Login:        row.Login,
		PasswordHash: row.PasswordHash,
		Role:         entity.Role(row.Role),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}

	return res, err
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (entity.User, error) {
	var res entity.User
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.getUserByLoginTx(ctx, login, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return entity.User{}, txErr
	}

	return res, nil
}

func (r *Repository) getUserByLoginTx(ctx context.Context, login string, tx *sqlx.Tx) (entity.User, error) {
	whereMap := map[string]any{
		"login": login,
	}

	sql, args, err := r.qb.Select("id").
		Columns("login", "password_hash", "role", "created_at", "updated_at").
		From(userTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("error building query: %w", err)
	}

	var row userRow
	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.User{}, fmt.Errorf("error executing query: %w", err)
	}

	return entity.User{
		Id:           row.Id,
		Login:        row.Login,
		PasswordHash: row.PasswordHash,
		Role:         entity.Role(row.Role),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, err
}
