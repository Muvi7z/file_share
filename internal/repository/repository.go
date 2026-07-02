package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	conn *sqlx.DB
	qb   sq.StatementBuilderType
}

func NewRepository(conn *sqlx.DB) *Repository {
	return &Repository{
		conn: conn,
		qb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
