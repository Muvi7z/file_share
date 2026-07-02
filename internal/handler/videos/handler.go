package videos

import (
	"context"
	"file_share/internal/entity"
)

type VideoService interface {
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.Video, error)
}

type Handler struct {
	videoService VideoService
}

func NewHandler(videoService VideoService) *Handler {
	return &Handler{
		videoService: videoService,
	}
}
