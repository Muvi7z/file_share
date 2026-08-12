package folder

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type folderRepository interface {
	GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot, enabled *bool) ([]entity.Folder, error)
	CreateFolder(ctx context.Context, folder entity.Folder) (entity.Folder, error)
	GetFolderById(ctx context.Context, id string) (entity.Folder, error)
	UpdateFolder(ctx context.Context, folder entity.UpdateFolderRequest, id string) (entity.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
}

type scanJobRepository interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type videoRepository interface {
	DeleteVideoByFolder(ctx context.Context, idRootFolder, parentFolderId string) error
}

type Service struct {
	logger            deps.Logger
	folderRepository  folderRepository
	scanJobRepository scanJobRepository
	videoRepository   videoRepository
}

func NewService(folderRepository folderRepository, logger deps.Logger, scanJobRepository scanJobRepository, videoRepository videoRepository) *Service {
	return &Service{
		logger:            logger,
		folderRepository:  folderRepository,
		scanJobRepository: scanJobRepository,
		videoRepository:   videoRepository,
	}
}

func (s *Service) DeleteFolder(ctx context.Context, id string) error {
	err := s.folderRepository.DeleteFolder(ctx, id)
	if err != nil {
		return entity.ErrorDeleteFolder
	}

	return nil
}

func (s *Service) UpdateFolder(ctx context.Context, update entity.UpdateFolderRequest, id string) (entity.Folder, error) {
	getFolder, err := s.folderRepository.GetFolderById(ctx, id)
	if err != nil {
		return entity.Folder{}, entity.ErrorGetFolder
	}

	if !getFolder.IsRoot {
		return entity.Folder{}, entity.ErrorUpdateFolder
	}

	folder, err := s.folderRepository.UpdateFolder(ctx, update, id)
	if err != nil {
		return entity.Folder{}, entity.ErrorUpdateFolder
	}

	return folder, nil
}

func (s *Service) GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot *bool) ([]entity.Folder, error) {
	folders, err := s.folderRepository.GetFolders(ctx, query, rootFolderId, parentFolderId, isRoot, nil)
	if err != nil {
		return nil, entity.ErrorGetFolders
	}

	return folders, nil
}

func (s *Service) GetFoldersEntries(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot *bool) ([]entity.FileBrowserEntry, error) {
	if parentFolderId != "" {
		_, err := s.folderRepository.GetFolderById(ctx, parentFolderId)
		if err != nil {
			if errors.Is(err, entity.ErrorNotFoundFolder) {
				return nil, entity.ErrorNotFoundFolder
			}
			return nil, entity.ErrorGetFolder
		}
	}

	folders, err := s.folderRepository.GetFolders(ctx, query, rootFolderId, parentFolderId, isRoot, nil)
	if err != nil {
		return nil, entity.ErrorGetFolders
	}

	browserEntityList := make([]entity.FileBrowserEntry, 0)

	for _, folder := range folders {
		browserEntity := entity.FileBrowserEntry{
			Type:   "folder",
			Folder: &folder,
		}

		browserEntityList = append(browserEntityList, browserEntity)
	}

	return browserEntityList, nil
}

func (s *Service) GetFolder(ctx context.Context, folderId string) (entity.Folder, error) {
	folder, err := s.folderRepository.GetFolderById(ctx, folderId)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundFolder) {
			return entity.Folder{}, entity.ErrorNotFoundFolder
		}
		return entity.Folder{}, entity.ErrorGetFolder
	}

	rootFolder, err := s.folderRepository.GetFolderById(ctx, folder.RootFolderId)
	if err != nil {
		return entity.Folder{}, fmt.Errorf("error getting root folder")
	}

	_ = rootFolder

	return folder, nil
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
