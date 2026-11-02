package playback

import (
	"context"
	"file_share/internal/entity"

	"github.com/google/uuid"
)

type mediaProcessor interface {
	CreateHLS(ctx context.Context, video entity.Video, audioTracks []entity.AudioTrack) error
	GetAudioTrack(inputPath string) ([]entity.AudioTrackStream, error)
}

type videoRepository interface {
}

type Service struct {
	mediaProcessor mediaProcessor
}

func NewService(mediaProcessor mediaProcessor) *Service {
	return &Service{mediaProcessor: mediaProcessor}
}

func (s *Service) PrepareHls(ctx context.Context, video entity.Video) error {
	//получить инфу о потоках аудио

	trackStreams, err := s.mediaProcessor.GetAudioTrack(video.Path)
	if err != nil {
		return err
	}

	var audioTracks []entity.AudioTrack
	var defaultFlag = false
	for _, trackStream := range trackStreams {
		if trackStream.CodecType == "audio" {
			audio := entity.AudioTrack{
				Id:          uuid.New().String(),
				Title:       trackStream.Tags.Title,
				VideoId:     video.Id,
				StreamIndex: trackStream.Index,
				Language:    trackStream.Tags.Language,
				Codec:       trackStream.CodecName,
				Channels:    trackStream.Disposition.Default,
				IsDefault:   false,
				Status:      entity.AudioTrackProcessing,
			}

			if !defaultFlag {
				audio.IsDefault = true
				defaultFlag = true
			}

			audioTracks = append(audioTracks, audio)
		}

	}

	err = s.mediaProcessor.CreateHLS(ctx, video, audioTracks)
	if err != nil {
		return err
	}

	//разделить подготовить потоки аудио и видео
	//сохранить аудиодорожки в бд

	return nil
}
