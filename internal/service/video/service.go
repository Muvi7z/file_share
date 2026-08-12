package video

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	video2 "file_share/pkg/utils/video"
	"fmt"
	"path/filepath"
	"strconv"
)

type videoRepository interface {
	GetAllVideo(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.Video, error)
	CreateVideo(ctx context.Context, video entity.Video) (entity.Video, error)
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
}

type fileStorage interface {
	Open(ctx context.Context, path string) (entity.VideoStream, error)
	OpenPoster(ctx context.Context, path string) (entity.PosterFile, error)
}

type posterGenerator interface {
	GeneratePosterFFmpeg(ctx context.Context, videoPath, videoId, duration string) (entity.PosterFile, error)
}

type Service struct {
	logger          deps.Logger
	videoRepository videoRepository
	fileStorage     fileStorage
	posterGenerator posterGenerator
	PosterDir       string
}

func NewService(videoRepository videoRepository, fileStorage fileStorage, posterGenerator posterGenerator, logger deps.Logger, PosterDir string) *Service {
	return &Service{
		logger:          logger,
		videoRepository: videoRepository,
		fileStorage:     fileStorage,
		posterGenerator: posterGenerator,
		PosterDir:       PosterDir,
	}
}

func (s *Service) CreateVideo(ctx context.Context, videoReq entity.Video) (entity.Video, error) {
	video, err := s.videoRepository.CreateVideo(ctx, videoReq)
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
		return entity.Video{}, entity.ErrorCreateVideo
	}

	duration, _ := strconv.Atoi(video.Duration)

	halfTime := video2.GetHalfTimeVideo(int64(duration))

	_, err = s.posterGenerator.GeneratePosterFFmpeg(ctx, video.Path, video.Id, halfTime)
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed generate poster: %v", err))
		return entity.Video{}, entity.ErrorCreateVideo
	}

	return video, nil
}

func (s *Service) GetPoster(ctx context.Context, id string) (entity.PosterFile, error) {
	video, err := s.videoRepository.GetVideoById(ctx, id)
	if err != nil {
		return entity.PosterFile{}, entity.ErrorGetVideo
	}

	posterPath := filepath.Join(s.PosterDir, "poster-"+video.Id+".jpg")

	poster, err := s.fileStorage.OpenPoster(ctx, posterPath)
	if err != nil {
		return entity.PosterFile{}, err
	}

	return poster, nil
}

func (s *Service) Stream(ctx context.Context, videoId string) (entity.VideoStream, error) {
	video, err := s.videoRepository.GetVideoById(ctx, videoId)
	if err != nil {
		return entity.VideoStream{}, entity.ErrorGetVideo
	}

	stream, err := s.fileStorage.Open(ctx, video.Path)
	if err != nil {
		return entity.VideoStream{}, entity.ErrorGetVideoStream
	}

	return stream, nil

}

func (s *Service) GetAll(ctx context.Context, query, rootFolderId, parentFolderId string, limit int, offset int) ([]entity.Video, error) {

	uLimit := uint64(limit)
	uOffset := uint64(offset)

	result, err := s.videoRepository.GetAllVideo(ctx, query, rootFolderId, parentFolderId, uLimit, uOffset)
	if err != nil {
		return nil, entity.ErrorGetVideos
	}

	return result, nil
}

func (s *Service) GetEntries(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.FileBrowserEntry, error) {
	videos, err := s.videoRepository.GetAllVideo(ctx, query, rootFolderId, parentFolderId, limit, offset)
	if err != nil {
		return nil, entity.ErrorGetVideosEntries
	}

	var result []entity.FileBrowserEntry

	for _, video := range videos {
		browserEntry := entity.FileBrowserEntry{
			Type:  "video",
			Video: &video,
		}
		result = append(result, browserEntry)
	}

	return result, nil
}

func (s *Service) GetVideoById(ctx context.Context, id string) (entity.Video, error) {
	video, err := s.videoRepository.GetVideoById(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundVideo) {
			return entity.Video{}, entity.ErrorNotFoundVideo
		}
		return entity.Video{}, entity.ErrorGetVideo
	}

	return video, nil
}
