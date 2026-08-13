package repository

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"time"
)

var sessionTable = "session"

type sessionRow struct {
	Token     string    `db:"token"`
	Login     string    `db:"login"`
	Role      string    `db:"role"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (r *Repository) CreateSession(ctx context.Context, session entity.Session) (entity.Session, error) {
	var res entity.Session
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createSessionTx(ctx, tx, session)
		if err != nil {
			return err
		}

		return nil

	})

	if txErr != nil {
		return entity.Session{}, txErr
	}

	return res, nil
}

func (r *Repository) createSessionTx(ctx context.Context, tx *sqlx.Tx, session entity.Session) (entity.Session, error) {
	insertMap := map[string]any{
		"token":      session.Token,
		"login":      session.Login,
		"role":       session.Role,
		"expires_at": session.ExpiresAt,
	}

	sql, args, err := r.qb.Insert(sessionTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.Session{}, fmt.Errorf("error building query: %w", err)
	}

	var row sessionRow
	var res entity.Session

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.Session{}, fmt.Errorf("error executing query: %w", err)
	}

	res = entity.Session{
		Token:     row.Token,
		Login:     row.Login,
		Role:      entity.Role(row.Role),
		ExpiresAt: row.ExpiresAt,
	}

	return res, err
}

func (r *Repository) GetSession(ctx context.Context, token string) (entity.Session, error) {
	var res entity.Session
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.getSessionTx(ctx, token, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return entity.Session{}, txErr
	}

	return res, nil
}

func (r *Repository) getSessionTx(ctx context.Context, token string, tx *sqlx.Tx) (entity.Session, error) {
	whereMap := map[string]any{
		"token": token,
	}

	sql, args, err := r.qb.Select("token").
		Columns("login", "role", "expires_at").
		From(sessionTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.Session{}, fmt.Errorf("error building query: %w", err)
	}

	var row sessionRow
	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.Session{}, fmt.Errorf("error executing query: %w", err)
	}

	return entity.Session{
		Token:     row.Token,
		Login:     row.Login,
		Role:      entity.Role(row.Role),
		ExpiresAt: row.ExpiresAt,
	}, err
}

func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteSessionTx(ctx, token, tx)
		if err != nil {
			return err
		}

		return err
	})

	if txErr != nil {
		return txErr
	}

	return nil
}

func (r *Repository) deleteSessionTx(ctx context.Context, token string, tx *sqlx.Tx) error {
	sql, args, err := r.qb.Delete(sessionTable).Where(sq.Eq{"token": token}).ToSql()
	if err != nil {
		return fmt.Errorf("error to building query %v", err)
	}

	row, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error to executing query %v", err)
	}

	rowsAffected, err := row.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to executing query %v", err)
	}

	if rowsAffected == 0 {

	}

	return nil
}
