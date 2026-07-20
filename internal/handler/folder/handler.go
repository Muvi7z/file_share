package folder

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type folderService interface {
	GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot *bool) ([]entity.Folder, error)
	GetFoldersEntries(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot *bool) ([]entity.FileBrowserEntry, error)
	CreateFolderRootFolder(ctx context.Context, path string) (entity.CreateRootFolderResponse, error)
	GetFolder(ctx context.Context, folderId string) (entity.Folder, error)
	UpdateFolder(ctx context.Context, update entity.UpdateFolderRequest, id string) (entity.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
}

type videoService interface {
	GetEntries(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.FileBrowserEntry, error)
}

type scanService interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type Handler struct {
	logger        deps.Logger
	folderService folderService
	videoService  videoService
	scanService   scanService
}

func NewHandler(folderService folderService, videoService videoService, scanService scanService, logger deps.Logger) *Handler {
	return &Handler{
		logger:        logger,
		folderService: folderService,
		videoService:  videoService,
		scanService:   scanService,
	}
}
