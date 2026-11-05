package videos

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type VideoService interface {
	GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int, enabled *bool) ([]entity.Video, error)
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
	Stream(ctx context.Context, videoId string) (entity.FileStream, error)
	GetPoster(ctx context.Context, id string) (entity.PosterFile, error)
	UpdatePoster(ctx context.Context, posterFile entity.PosterFile) (string, error)
	DeleteVideo(ctx context.Context, id string) error
}

type audioTrackService interface {
	GetAll(ctx context.Context, videoId string, limit int, offset int) ([]entity.AudioTrack, error)
}

type PlaybackService interface {
	StreamHLS(ctx context.Context, fileName, id, typeFile string) (entity.FileStream, error)
}

type Handler struct {
	logger            deps.Logger
	videoService      VideoService
	audioTrackService audioTrackService
	playbackService   PlaybackService
}

func NewHandler(videoService VideoService, playbackService PlaybackService, audioTrackService audioTrackService, logger deps.Logger) *Handler {
	return &Handler{
		videoService:      videoService,
		playbackService:   playbackService,
		logger:            logger,
		audioTrackService: audioTrackService,
	}
}
