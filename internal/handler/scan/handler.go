package scan

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
)

type scanService interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
	GetScanJob(ctx context.Context, id string) (entity.ScanJob, error)
}

type Handler struct {
	logger      deps.Logger
	scanService scanService
}

func NewHandler(scanService scanService, logger deps.Logger) *Handler {
	return &Handler{
		logger:      logger,
		scanService: scanService,
	}
}
