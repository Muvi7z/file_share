package cmd

import (
	"context"
	"file_share/internal/generator"
	folder2 "file_share/internal/handler/folder"
	scan2 "file_share/internal/handler/scan"
	"file_share/internal/handler/videos"
	"file_share/internal/repository"
	"file_share/internal/server"
	"file_share/internal/service/folder"
	"file_share/internal/service/scan"
	"file_share/internal/service/video"
	"file_share/internal/storage"
	"file_share/migrations"
	"file_share/pkg/logger"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type Container struct {
	configuration *configuration

	repository *repository.Repository

	folderService *folder.Service
	videoService  *video.Service
	scanService   *scan.Scan

	folderHandler *folder2.Handler
	videoHandler  *videos.Handler
	scanHandler   *scan2.Handler

	fileStorage *storage.Storage

	posterGenerator *generator.PosterGenerator

	server *server.Server

	tempDir string

	ctx               context.Context
	db                *sqlx.DB
	postgresContainer *postgres.PostgresContainer
	migrator          *migrations.Migrator
	logger            *logger.Logger
}

func Init() (*Container, func(), error) {
	c := &Container{}

	config, err := newFromEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("error to load configuration: %w", err)
	}

	c.configuration = config
	c.ctx = context.Background()
	c.logger = logger.New()

	c.db, err = NewSqlxConn(c.configuration.GetPostgresConfiguration())
	if err != nil {
		return nil, nil, fmt.Errorf("error to connect to database: %w", err)
	}

	c.tempDir = c.configuration.GetPosterDir()
	closer := func() {
		if err = c.db.Close(); err != nil {
			c.logger.Error(c.ctx, fmt.Errorf("failed to close database: %w", err))
		}
	}

	return c, closer, nil
}

func (c *Container) GetServerAddress() string {
	return c.configuration.GetServerConfiguration().GetAddress()
}

func (c *Container) GetPostgresConnectionString() string {
	return c.configuration.GetPostgresConfiguration().GetConnectionString()
}

func (c *Container) GetContext() context.Context {
	return c.ctx
}

func (c *Container) GetDB() *sqlx.DB {
	return c.db
}

func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

func (c *Container) GetMigrator() *migrations.Migrator {
	if c.migrator == nil {
		c.migrator = migrations.NewMigrator(c.configuration.GetPostgresConfiguration().GetConnectionString())
	}

	return c.migrator
}

func (c *Container) GetRepository() *repository.Repository {
	if c.repository == nil {
		c.repository = repository.NewRepository(c.db)
	}

	return c.repository
}

func (c *Container) GetFileStorage() *storage.Storage {
	if c.fileStorage == nil {
		c.fileStorage = storage.New()
	}

	return c.fileStorage
}

func (c *Container) GetFolderService() *folder.Service {
	if c.folderService == nil {
		c.folderService = folder.NewService(
			c.GetRepository(),
			c.GetLogger(),
			c.GetRepository(),
		)
	}

	return c.folderService
}

func (c *Container) GetPosterGenerator() *generator.PosterGenerator {
	if c.posterGenerator == nil {
		c.posterGenerator = generator.NewPosterGenerator(c.tempDir)
	}

	return c.posterGenerator
}

func (c *Container) GetVideoService() *video.Service {
	if c.videoService == nil {
		c.videoService = video.NewService(
			c.GetRepository(),
			c.GetFileStorage(),
			c.GetPosterGenerator(),
			c.GetLogger(),
			c.tempDir,
		)
	}

	return c.videoService
}

func (c *Container) GetScanService() *scan.Scan {
	if c.scanService == nil {
		c.scanService = scan.New(
			c.GetLogger(),
			c.GetRepository(),
			c.GetRepository(),
			c.GetPosterGenerator(),
		)
	}

	return c.scanService
}

func (c *Container) GetVideoHandler() *videos.Handler {
	if c.videoHandler == nil {
		c.videoHandler = videos.NewHandler(
			c.GetVideoService(),
			c.GetLogger(),
		)
	}

	return c.videoHandler
}

func (c *Container) GetScanHandler() *scan2.Handler {
	if c.scanHandler == nil {
		c.scanHandler = scan2.NewHandler(
			c.GetScanService(),
			c.GetLogger(),
		)
	}

	return c.scanHandler
}

func (c *Container) GetFolderHandler() *folder2.Handler {
	if c.folderHandler == nil {
		c.folderHandler = folder2.NewHandler(
			c.GetFolderService(),
			c.GetVideoService(),
			c.GetScanService(),
			c.GetLogger(),
		)
	}

	return c.folderHandler
}

func (c *Container) GetServer() *server.Server {
	if c.server == nil {

		c.server = server.NewServer(
			c.GetVideoHandler(),
			c.GetFolderHandler(),
			c.GetScanHandler(),
			c.GetLogger(),
			c.GetServerAddress(),
		)
	}

	return c.server
}
