package cmd

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewSqlxConn(configuration *postgresConfiguration) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", configuration.GetConnectionString())
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("cant connect to db"))
	}

	db.SetMaxIdleConns(configuration.GetMaxIdleConns())
	db.SetMaxOpenConns(configuration.GetMaxOpenConns())

	if err = db.Ping(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("cant ping db"))
	}

	return db, nil
}
