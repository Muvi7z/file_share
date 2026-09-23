package videoFixer

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	video2 "file_share/pkg/utils/video"
	"time"
)

type repository interface {
	GetFolderById(ctx context.Context, id string) (entity.Folder, error)
	GetAllReportVideo(ctx context.Context, status string, videoId string) ([]entity.ReportVideo, error)
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
	UpdateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error)
}

type VideoFixer struct {
	repository repository
	logger     deps.Logger
}

func New(logger deps.Logger, repository repository) *VideoFixer {
	return &VideoFixer{
		logger:     logger,
		repository: repository,
	}
}

func (f *VideoFixer) StartProcessFix(ctx context.Context, handlePeriod time.Duration) {
	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				f.logger.Info(ctx, "Stop video fixer")
				return
			case <-ticker.C:
			}

			reports, err := f.repository.GetAllReportVideo(ctx, entity.RepostTypeOpen, "")
			if err != nil {
				f.logger.Error(ctx, err)
				continue
			}

			for _, report := range reports {
				video, err := f.repository.GetVideoById(ctx, report.VideoId)
				if err != nil {
					f.logger.Error(ctx, err)
					continue
				}

				err = video2.FixFastStart(ctx, video.Path)
				if err != nil {
					f.logger.Error(ctx, err)
					continue
				}

				report.Status = entity.RepostTypeAuto

				_, err = f.repository.UpdateReportVideo(ctx, report)
				if err != nil {
					f.logger.Error(ctx, err)
					continue
				}

			}
		}
	}()
}
