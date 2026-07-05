package cmd

import (
	"context"
	"file_share/migrations"
	"file_share/pkg/logger"

	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type container struct {
	configuration *configuration

	ctx               context.Context
	db                *sqlx.DB
	postgresContainer *postgres.PostgresContainer
	migrator          *migrations.Migrator
	logger            *logger.Logger
}
