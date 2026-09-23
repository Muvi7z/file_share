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

const reportVideoTable = "report_video"

type ReportVideoRow struct {
	Id             string    `db:"id"`
	VideoId        string    `db:"video_id"`
	Title          string    `db:"title"`
	Reason         string    `db:"reason"`
	PositionSecond int64     `db:"position_second"`
	Comment        string    `db:"comment"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
}

func (r *Repository) CreateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error) {
	var res entity.ReportVideo
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createReportVideoTx(ctx, report, tx)
		if err != nil {
			return err
		}
		return err
	})

	if txErr != nil {
		return entity.ReportVideo{}, txErr
	}

	return res, nil
}

func (r *Repository) createReportVideoTx(ctx context.Context, report entity.ReportVideo, tx *sqlx.Tx) (entity.ReportVideo, error) {
	insertMap := map[string]any{
		"id":              report.Id,
		"video_id":        report.VideoId,
		"title":           report.Title,
		"reason":          report.Reason,
		"position_second": report.PositionSecond,
		"comment":         report.Comment,
		"status":          report.Status,
		"created_at":      report.CreatedAt,
	}

	sql, args, err := r.qb.Insert(reportVideoTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.ReportVideo{}, fmt.Errorf("error to building query: %w", err)
	}

	var row ReportVideoRow
	var result entity.ReportVideo

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.ReportVideo{}, fmt.Errorf("error to executing query: %w", err)
	}

	result = entity.ReportVideo{
		Id:             row.Id,
		VideoId:        row.VideoId,
		Title:          row.Title,
		Reason:         row.Reason,
		PositionSecond: row.PositionSecond,
		Comment:        row.Comment,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt,
	}

	return result, nil
}

func (r *Repository) UpdateReportVideo(ctx context.Context, report entity.ReportVideo) (entity.ReportVideo, error) {
	var res entity.ReportVideo
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateReportVideoTx(ctx, report, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return entity.ReportVideo{}, txErr
	}

	return res, nil
}

func (r *Repository) updateReportVideoTx(ctx context.Context, report entity.ReportVideo, tx *sqlx.Tx) (entity.ReportVideo, error) {
	updateMap := map[string]any{}

	if report.Status != "" {
		updateMap["status"] = report.Status
	}

	sql, args, err := r.qb.Update(reportVideoTable).
		SetMap(updateMap).
		Suffix("RETURNING *").
		Where(sq.Eq{"id": report.Id}).
		ToSql()
	if err != nil {
		return entity.ReportVideo{}, fmt.Errorf("error to building query: %w", err)
	}

	var row ReportVideoRow

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ReportVideo{}, err
		}

		return entity.ReportVideo{}, fmt.Errorf("error to update: %w", err)
	}

	return entity.ReportVideo{
		Id:             row.Id,
		VideoId:        row.VideoId,
		Title:          row.Title,
		Reason:         row.Reason,
		PositionSecond: row.PositionSecond,
		Comment:        row.Comment,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repository) GetAllReportVideo(ctx context.Context, status string, videoId string) ([]entity.ReportVideo, error) {
	var whereMap sq.Sqlizer

	if status != "" {
		whereMap = sq.Eq{"status": status}
	}

	if videoId != "" {
		if whereMap == nil {
			whereMap = sq.Eq{"video_id": videoId}
		} else {
			whereMap = sq.And{
				whereMap,
				sq.Eq{"video_id": videoId},
			}

		}
	}

	sql, args, err := r.qb.Select("id").
		Columns("video_id", "title", "reason", "position_second", "comment", "status", "created_at").
		From(reportVideoTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query: %w", err)
	}

	var rows []ReportVideoRow

	err = r.conn.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	var results []entity.ReportVideo

	for _, row := range rows {
		results = append(results, entity.ReportVideo{
			Id:             row.Id,
			VideoId:        row.VideoId,
			Title:          row.Title,
			Reason:         row.Reason,
			PositionSecond: row.PositionSecond,
			Comment:        row.Comment,
			Status:         row.Status,
			CreatedAt:      row.CreatedAt,
		})
	}

	return results, nil
}

func (r *Repository) GetReportVideoById(ctx context.Context, id string) (entity.ReportVideo, error) {
	whereMap := map[string]any{
		"id": id,
	}
	sql, args, err := r.qb.Select("id").
		Columns("video_id", "title", "reason", "position_second", "comment", "status", "created_at").
		From(reportVideoTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.ReportVideo{}, fmt.Errorf("error to building query: %w", err)
	}

	var row ReportVideoRow

	err = r.conn.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ReportVideo{}, entity.ErrorNotFoundVideo
		}
		return entity.ReportVideo{}, fmt.Errorf("error to executing query: %w", err)
	}

	return entity.ReportVideo{
		Id:             row.Id,
		VideoId:        row.VideoId,
		Title:          row.Title,
		Reason:         row.Reason,
		PositionSecond: row.PositionSecond,
		Comment:        row.Comment,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repository) DeleteReportVideo(ctx context.Context, id string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteReportVideoByIdTx(ctx, id, tx)
		if err != nil {
			return err
		}

		return err
	})

	if txErr != nil {
		return txErr
	}

	return nil
}

func (r *Repository) deleteReportVideoByIdTx(ctx context.Context, id string, tx *sqlx.Tx) error {
	sql, args, err := r.qb.Delete(reportVideoTable).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("error to building query %v", err)
	}

	row, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error to executing query %v", err)
	}

	rowsAffected, err := row.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to executing query %v", err)
	}

	if rowsAffected == 0 {

	}

	return nil
}
