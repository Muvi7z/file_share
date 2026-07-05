package folders

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type folderService interface {
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string) ([]entity.Folder, error)
}

type Handler struct {
	logger        deps.Logger
	folderService folderService
}

func NewHandler(folderService folderService, logger deps.Logger) *Handler {
	return &Handler{
		logger:        logger,
		folderService: folderService,
	}
}
