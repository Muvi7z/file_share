package videos

import (
	"context"
	"file_share/internal/entity"
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
