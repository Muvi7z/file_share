package file

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type fileRepository interface {
	GetFileById(ctx context.Context, id string) (entity.File, error)
	DeleteFile(ctx context.Context, id string) error
	UpdateFile(ctx context.Context, file entity.File, id string) (entity.File, error)
	CreateFile(ctx context.Context, file entity.File) (entity.File, error)
	GetAllFile(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.File, error)
}

type Service struct {
	logger         deps.Logger
	fileRepository fileRepository
}

func NewService(logger deps.Logger, fileRepository fileRepository) *Service {
	return &Service{
		logger:         logger,
		fileRepository: fileRepository,
	}
}

func (s *Service) DeleteFile(ctx context.Context, id string) error {
	err := s.fileRepository.DeleteFile(ctx, id)
	if err != nil {
		return entity.ErrorDeleteFile
	}

	return nil
}

func (s *Service) CreateFile(ctx context.Context, file entity.File) (entity.File, error) {
	res, err := s.fileRepository.CreateFile(ctx, file)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.File{}, entity.ErrorCreateFile
	}

	return res, nil
}

func (s *Service) GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.File, error) {
	uLimit := uint64(limit)
	uOffset := uint64(offset)

	result, err := s.fileRepository.GetAllFile(ctx, query, rootFolderId, parentFolderId, uLimit, uOffset)
	if err != nil {
		return nil, entity.ErrorGetFiles
	}

	return result, nil
}

func (s *Service) GetFileById(ctx context.Context, id string) (entity.File, error) {
	file, err := s.fileRepository.GetFileById(ctx, id)
	if err != nil {
		return entity.File{}, entity.ErrorGetFile
	}

	return file, nil
}

func (s *Service) GetEntries(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.FileBrowserEntry, error) {
	files, err := s.fileRepository.GetAllFile(ctx, query, rootFolderId, parentFolderId, limit, offset)
	if err != nil {
		return nil, entity.ErrorGetVideosEntries
	}

	var result []entity.FileBrowserEntry

	for _, file := range files {
		browserEntry := entity.FileBrowserEntry{
			Type: entity.FileTypeOther,
			File: &file,
		}
		result = append(result, browserEntry)
	}

	return result, nil
}
