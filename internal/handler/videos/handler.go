package videos

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type VideoService interface {
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.Video, error)
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
	Stream(ctx context.Context, videoId string) (entity.VideoStream, error)
	GetPoster(ctx context.Context, id string) (entity.PosterFile, error)
}

type Handler struct {
	logger       deps.Logger
	videoService VideoService
}

func NewHandler(videoService VideoService, logger deps.Logger) *Handler {
	return &Handler{
		videoService: videoService,
		logger:       logger,
	}
}
