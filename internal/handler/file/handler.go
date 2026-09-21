package file

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type fileService interface {
	CreateFile(ctx context.Context, file entity.File) (entity.File, error)
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.File, error)
	GetEntries(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.FileBrowserEntry, error)
	GetFileById(ctx context.Context, id string) (entity.File, error)
}

type Handler struct {
	fileService fileService
	logger      deps.Logger
}

func NewHandler(fileService fileService, logger deps.Logger) *Handler {
	return &Handler{
		fileService: fileService,
		logger:      logger,
	}
}
