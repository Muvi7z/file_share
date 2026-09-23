package report

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"

	"github.com/google/uuid"
)

type reportVideoRepository interface {
	CreateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error)
	UpdateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error)
	GetAllReportVideo(ctx context.Context, status string, videoId string) ([]entity.ReportVideo, error)
	GetReportVideoById(ctx context.Context, id string) (entity.ReportVideo, error)
	DeleteReportVideo(ctx context.Context, id string) error
}

type videoRepository interface {
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
}

type Service struct {
	logger                deps.Logger
	reportVideoRepository reportVideoRepository
	videoRepository       videoRepository
}

func NewService(logger deps.Logger, reportVideoRepository reportVideoRepository, videoRepository videoRepository) *Service {
	return &Service{
		logger:                logger,
		reportVideoRepository: reportVideoRepository,
		videoRepository:       videoRepository,
	}
}

func (s *Service) CreateReportVideo(ctx context.Context, report entity.CreateReportVideo) (entity.ReportVideo, error) {
	_, err := s.videoRepository.GetVideoById(ctx, report.VideoId)
	if err != nil {
		return entity.ReportVideo{}, err
	}

	reportCreate := entity.ReportVideo{
		Id:             uuid.New().String(),
		VideoId:        report.VideoId,
		Title:          report.Title,
		Reason:         report.Reason,
		PositionSecond: report.PositionSecond,
		Comment:        report.Comment,
		Status:         report.Status,
		CreatedAt:      report.CreatedAt,
	}

	res, err := s.reportVideoRepository.CreateReportVideo(ctx, reportCreate)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.ReportVideo{}, entity.ErrorReportVideoCreate
	}

	return res, nil
}

func (s *Service) UpdateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error) {
	report, err := s.reportVideoRepository.UpdateReportVideo(ctx, report)
	if err != nil {
		s.logger.Error(ctx, err)
		return entity.ReportVideo{}, entity.ErrorReportVideoUpdate
	}

	return report, nil

}

func (s *Service) GetAllReportVideo(ctx context.Context, status string, videoId string) ([]entity.ReportVideo, error) {
	reports, err := s.reportVideoRepository.GetAllReportVideo(ctx, status, videoId)
	if err != nil {
		s.logger.Error(ctx, err)
		return nil, entity.ErrorReportVideoGet
	}

	return reports, nil
}
