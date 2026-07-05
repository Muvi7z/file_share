package folder

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type folderRepository interface {
	GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string) ([]entity.Folder, error)
	CreateFolder(ctx context.Context, folder entity.Folder) (entity.Folder, error)
}

type scanJobRepository interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type Service struct {
	logger            deps.Logger
	folderRepository  folderRepository
	scanJobRepository scanJobRepository
}

func NewService(folderRepository folderRepository, logger deps.Logger, scanJobRepository scanJobRepository) *Service {
	return &Service{
		logger:            logger,
		folderRepository:  folderRepository,
		scanJobRepository: scanJobRepository,
	}
}

func (s *Service) GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string) ([]entity.Folder, error) {
	return s.folderRepository.GetFolders(ctx, query, rootFolderId, parentFolderId)
}

func (s *Service) CreateFolderRootFolder(ctx context.Context, path string) (entity.CreateRootFolderResponse, error) {
	if err := s.checkPath(path); err != nil {
		return entity.CreateRootFolderResponse{}, fmt.Errorf("invalid path")
	}

	folderUuid := uuid.New().String()

	lastElement := filepath.Base(path)

	folder := entity.Folder{
		Id:               folderUuid,
		Name:             lastElement,
		Path:             path,
		RootFolderId:     folderUuid,
		IsRoot:           true,
		Enabled:          false,
		FilesCount:       0,
		VideosCount:      0,
		ChildFolderCount: 0,
	}

	_, err := s.folderRepository.CreateFolder(ctx, folder)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.CreateRootFolderResponse{}, fmt.Errorf("create folder")
	}

	scanJobID := uuid.New().String()

	scanJob := entity.ScanJob{
		Id:               scanJobID,
		FolderId:         folderUuid,
		Status:           "queued",
		ProcessedVideos:  0,
		ProcessedFolders: 0,
		StartedAt:        time.Now(),
		Error:            "",
	}

	job, err := s.scanJobRepository.CreateScanJob(ctx, scanJob)
	if err != nil {
		return entity.CreateRootFolderResponse{}, err
	}

	result := entity.CreateRootFolderResponse{
		Folder:  folder,
		ScanJob: job,
	}

	return result, nil
}

func (s *Service) checkPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("path is not absolute")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	}

	return nil
}
