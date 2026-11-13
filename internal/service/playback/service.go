package playback

import (
	"context"
	"errors"
	"file_share/internal/entity"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type mediaProcessor interface {
	CreateHLS(ctx context.Context, video entity.Video, audioTracks []entity.AudioTrack) error
	GetAudioTrack(inputPath string) ([]entity.AudioTrackStream, error)
}

type videoService interface {
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
}

type audioService interface {
	CreateAudioTracks(ctx context.Context, audioTracks []entity.AudioTrack) ([]entity.AudioTrack, error)
	UpdateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack, id string) (entity.AudioTrack, error)
	GetAudioTrackById(ctx context.Context, id string) (entity.AudioTrack, error)
}

type fileStorage interface {
	Open(ctx context.Context, path string) (entity.FileStream, error)
}

type Service struct {
	mediaProcessor mediaProcessor
	audioService   audioService
	videoService   videoService
	fileStorage    fileStorage
	WorkDir        string
}

func NewService(mediaProcessor mediaProcessor, audioService audioService, videoService videoService, fileStorage fileStorage, workDir string) *Service {
	return &Service{
		mediaProcessor: mediaProcessor,
		audioService:   audioService,
		videoService:   videoService,
		fileStorage:    fileStorage,
		WorkDir:        workDir,
	}
}

func (s *Service) StreamHLS(ctx context.Context, fileName, id, typeFile string) (entity.FileStream, error) {
	switch typeFile {
	case entity.FileTypeVideo:
		video, err := s.videoService.GetVideoById(ctx, id)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideo
		}

		path := filepath.Join(s.WorkDir, video.Id, "video", fileName)

		stream, err := s.fileStorage.Open(ctx, path)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideoStream
		}
		return stream, nil
	case entity.FileTypeAudio:
		audio, err := s.audioService.GetAudioTrackById(ctx, id)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideo
		}

		path := filepath.Join(s.WorkDir, audio.VideoId, audio.Id, fileName)

		stream, err := s.fileStorage.Open(ctx, path)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideoStream
		}
		return stream, nil
	case entity.FileTypePlaylistHLS:
		video, err := s.videoService.GetVideoById(ctx, id)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideo
		}

		path := filepath.Join(s.WorkDir, video.Id, fileName)

		stream, err := s.fileStorage.Open(ctx, path)
		if err != nil {
			return entity.FileStream{}, entity.ErrorGetVideoStream
		}
		return stream, nil
	}

	return entity.FileStream{}, errors.New("invalid file type")
}

func (s *Service) PrepareHls(ctx context.Context, video entity.Video) ([]entity.AudioTrack, func() error, error) {
	//получить инфу о потоках аудио

	trackStreams, err := s.mediaProcessor.GetAudioTrack(video.Path)
	if err != nil {
		return nil, nil, err
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

	//разделить и подготовить потоки аудио и видео
	//_, err = s.audioService.CreateAudioTracks(ctx, audioTracks)
	//if err != nil {
	//	return err
	//}

	err = s.mediaProcessor.CreateHLS(ctx, video, audioTracks)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(filepath.Join(s.WorkDir, video.Id))
	}

	//изменить аудиодорожки в бд
	//for _, audioTrack := range audioTracks {
	//	audioTrack.Status = entity.AudioTrackReady
	//	_, err = s.audioService.UpdateAudioTrack(ctx, audioTrack, video.Id)
	//	if err != nil {
	//		return err
	//	}
	//}

	return audioTracks, cleanup, nil
}
