package audioTrack

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type audioTrackRepository interface {
	CreateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack) (entity.AudioTrack, error)
	CreateAudioTracks(ctx context.Context, audioTracks []entity.AudioTrack) ([]entity.AudioTrack, error)
	GetAudioTrackById(ctx context.Context, id string) (entity.AudioTrack, error)
	GetAllAudioTrack(ctx context.Context, videoId string, limit uint64, offset uint64) ([]entity.AudioTrack, error)
	UpdateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack, id string) (entity.AudioTrack, error)
	DeleteAudioTrack(ctx context.Context, id string) error
}

type Service struct {
	logger               deps.Logger
	audioTrackRepository audioTrackRepository
}

func NewService(logger deps.Logger, audioTrackRepository audioTrackRepository) *Service {
	return &Service{
		logger:               logger,
		audioTrackRepository: audioTrackRepository,
	}
}

func (s *Service) CreateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack) (entity.AudioTrack, error) {
	res, err := s.audioTrackRepository.CreateAudioTrack(ctx, audioTrack)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.AudioTrack{}, entity.ErrorCreateAudioTrack
	}
	return res, nil
}

func (s *Service) CreateAudioTracks(ctx context.Context, audioTracks []entity.AudioTrack) ([]entity.AudioTrack, error) {
	res, err := s.audioTrackRepository.CreateAudioTracks(ctx, audioTracks)
	if err != nil {
		s.logger.Error(ctx, err)
		return nil, entity.ErrorCreateAudioTrack
	}
	return res, nil
}

func (s *Service) GetAudioTrackById(ctx context.Context, id string) (entity.AudioTrack, error) {
	audioTrack, err := s.audioTrackRepository.GetAudioTrackById(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundAudioTrack) {
			return entity.AudioTrack{}, entity.ErrorNotFoundAudioTrack
		}
		s.logger.Error(ctx, err)
		return entity.AudioTrack{}, entity.ErrorGetAudioTrack
	}
	return audioTrack, nil
}

func (s *Service) GetAll(ctx context.Context, videoId string, limit int, offset int) ([]entity.AudioTrack, error) {
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	result, err := s.audioTrackRepository.GetAllAudioTrack(ctx, videoId, uint64(limit), uint64(offset))
	if err != nil {
		s.logger.Error(ctx, err)
		return nil, entity.ErrorGetAudioTracks
	}
	return result, nil
}

func (s *Service) UpdateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack, id string) (entity.AudioTrack, error) {
	res, err := s.audioTrackRepository.UpdateAudioTrack(ctx, audioTrack, id)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundAudioTrack) {
			return entity.AudioTrack{}, entity.ErrorNotFoundAudioTrack
		}
		s.logger.Error(ctx, err)
		return entity.AudioTrack{}, entity.ErrorUpdateAudioTrack
	}
	return res, nil
}

func (s *Service) DeleteAudioTrack(ctx context.Context, id string) error {
	err := s.audioTrackRepository.DeleteAudioTrack(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundAudioTrack) {
			return entity.ErrorNotFoundAudioTrack
		}
		s.logger.Error(ctx, err)
		return entity.ErrorDeleteAudioTrack
	}
	return nil
}
