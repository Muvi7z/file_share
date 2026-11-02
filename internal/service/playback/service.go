package playback

import "file_share/internal/infrastructure/ffmpeg"

type mediaProcessor interface {
}

type videoRepository interface {
}

type Service struct {
	ffmpeg ffmpeg.FFMPEG
}

func NewService(ffmpeg ffmpeg.FFMPEG) *Service {
	return &Service{ffmpeg: ffmpeg}
}

func (s *Service) PrepareHls() {
	//получить инфу о потоках аудио
	//разделить подготовить потоки аудио и видео
	//сохранить аудиодорожки в бд

}
