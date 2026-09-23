package report

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type reportVideoService interface {
	CreateReportVideo(ctx context.Context, report entity.CreateReportVideo) (entity.ReportVideo, error)
	UpdateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error)
}

type Handler struct {
	reportVideoService reportVideoService
	logger             deps.Logger
}

func NewHandler(reportVideoService reportVideoService, logger deps.Logger) *Handler {
	return &Handler{
		reportVideoService: reportVideoService,
		logger:             logger,
	}
}
