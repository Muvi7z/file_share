package videos

import (
	"context"
	"file_share/internal/entity"
	"fmt"
)

type videoRepository interface {
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.Video, error)
	CreateVideo(ctx context.Context, video entity.Video) (entity.Video, error)
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
}

type Service struct {
	videoRepository videoRepository
}

func NewService(videoRepository videoRepository) *Service {
	return &Service{
		videoRepository: videoRepository,
	}
}

func (s *Service) GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.Video, error) {

	uLimit := uint64(limit)
	uOffset := uint64(offset)

	result, err := s.videoRepository.GetAll(ctx, query, rootFolderId, parentFolderId, uLimit, uOffset)
	if err != nil {
		return nil, fmt.Errorf("error getting videos")
	}

	return result, nil
}
