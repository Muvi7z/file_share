package repository

import (
	"context"
	"errors"
	"file_share/internal/entity"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type scanJobRow struct {
	Id               string    `db:"id"`
	FolderId         string    `db:"folder_id"`
	Status           string    `db:"status"`
	ProcessedVideos  int       `db:"processed_videos"`
	ProcessedFolders int       `db:"processed_folders"`
	StartedAt        time.Time `db:"started_at"`
	FinishedAt       time.Time `db:"finished_at"`
	Error            string    `db:"error"`
}

var scanJobTable = "scan_jobs"

func (r *Repository) GetScanJob(ctx context.Context, id string) (entity.ScanJob, error) {
	whereMap := map[string]any{
		"id": id,
	}
	//30

	sql, args, err := r.qb.Select("id").
		Columns("folder_id", "status", "processed_videos", "processed_folders", "started_at", "finished_at", "error").
		From(scanJobTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.ScanJob{}, fmt.Errorf("error building query: %w", err)
	}

	var row scanJobRow
	err = r.conn.GetContext(ctx, &row, sql, args...)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ScanJob{}, entity.ErrorNotFoundScan
		}
		return entity.ScanJob{}, fmt.Errorf("error executing query: %w", err)
	}

	return entity.ScanJob{
		Id:               row.Id,
		FolderId:         row.FolderId,
		Status:           row.Status,
		ProcessedVideos:  row.ProcessedVideos,
		ProcessedFolders: row.ProcessedFolders,
		StartedAt:        row.StartedAt,
		FinishedAt:       row.FinishedAt,
		Error:            row.Error,
	}, nil
}

func (r *Repository) GetAllScanJob(ctx context.Context, status string) ([]entity.ScanJob, error) {
	var whereMap sq.Sqlizer

	if status != "" {
		whereMap = sq.Eq{"status": status}

	}
	sql, args, err := r.qb.Select("id").
		Columns("folder_id", "status", "processed_videos", "processed_folders", "started_at", "finished_at", "error").
		From(scanJobTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %w", err)
	}

	var rows []scanJobRow

	err = r.conn.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	var result []entity.ScanJob
	for _, row := range rows {
		result = append(result, entity.ScanJob{
			Id:               row.Id,
			FolderId:         row.FolderId,
			Status:           row.Status,
			ProcessedVideos:  row.ProcessedVideos,
			ProcessedFolders: row.ProcessedFolders,
			StartedAt:        row.StartedAt,
			FinishedAt:       row.FinishedAt,
			Error:            row.Error,
		})
	}

	return result, err
}

func (r *Repository) CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error) {
	var res entity.ScanJob
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createScanJobTx(ctx, job, tx)
		if err != nil {
			return err
		}
		return err
	})

	if txErr != nil {
		return entity.ScanJob{}, txErr
	}

	return res, nil
}

func (r *Repository) createScanJobTx(ctx context.Context, job entity.ScanJob, tx *sqlx.Tx) (entity.ScanJob, error) {
	insertMap := map[string]any{
		"id":                job.Id,
		"folder_id":         job.FolderId,
		"status":            job.Status,
		"processed_videos":  job.ProcessedVideos,
		"processed_folders": job.ProcessedFolders,
		"started_at":        job.StartedAt,
		"finished_at":       job.FinishedAt,
		"error":             job.Error,
	}

	sql, args, err := r.qb.Insert(scanJobTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.ScanJob{}, fmt.Errorf("error to building query: %w", err)
	}

	var row scanJobRow
	var result entity.ScanJob

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.ScanJob{}, fmt.Errorf("error to executing query: %w", err)
	}

	result = entity.ScanJob{
		Id:               row.Id,
		FolderId:         row.FolderId,
		Status:           row.Status,
		ProcessedVideos:  row.ProcessedVideos,
		ProcessedFolders: row.ProcessedFolders,
		StartedAt:        row.StartedAt,
		FinishedAt:       row.FinishedAt,
		Error:            row.Error,
	}

	return result, nil
}

func (r *Repository) UpdateJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error) {
	var res entity.ScanJob
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateJobTx(ctx, job, tx)
		if err != nil {
			return err
		}

		return err
	})

	if txErr != nil {
		return entity.ScanJob{}, txErr
	}

	return res, nil
}

func (r *Repository) updateJobTx(ctx context.Context, job entity.ScanJob, tx *sqlx.Tx) (entity.ScanJob, error) {
	updateMap := map[string]any{
		"finished_at": job.FinishedAt,
		"started_at":  job.StartedAt,
	}

	if job.Status != "" {
		updateMap["status"] = job.Status
	}

	if job.ProcessedVideos > 0 {
		updateMap["processed_videos"] = job.ProcessedVideos
	}

	if job.ProcessedFolders > 0 {
		updateMap["processed_folders"] = job.ProcessedFolders
	}

	sql, args, err := r.qb.Update(scanJobTable).
		SetMap(updateMap).
		Suffix("RETURNING *").
		Where(sq.Eq{"id": job.Id}).
		ToSql()
	if err != nil {
		return entity.ScanJob{}, fmt.Errorf("error to building query: %w", err)
	}

	var row scanJobRow

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.ScanJob{}, fmt.Errorf("error to executing query: %w", err)
	}

	return entity.ScanJob{
		Id:               row.Id,
		FolderId:         row.FolderId,
		Status:           row.Status,
		ProcessedVideos:  row.ProcessedVideos,
		ProcessedFolders: row.ProcessedFolders,
		StartedAt:        row.StartedAt,
		FinishedAt:       row.FinishedAt,
		Error:            row.Error,
	}, nil

}
