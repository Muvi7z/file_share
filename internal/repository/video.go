package repository

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type VideoRow struct {
	Id             string    `db:"id"`
	Title          string    `db:"title"`
	FolderId       string    `db:"folder_id"`
	FolderName     string    `db:"folder_name"`
	ParentFolderId string    `db:"parent_folder_id"`
	Size           string    `db:"size"`
	SizeBytes      int64     `db:"size_bytes"`
	Duration       string    `db:"duration"`
	ModifiedAt     time.Time `db:"modified_at"`
	Codec          string    `db:"codec"`
	Resolution     string    `db:"resolution"`
	PosterUrl      string    `db:"poster_url"`
	StreamUrl      string    `db:"stream_url"`
	Path           string    `db:"path"`
}

const videosTable = "videos"

func (r *Repository) GetAllVideo(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.Video, error) {
	var whereMap sq.Sqlizer

	if rootFolderId != "" {
		whereMap = sq.Eq{"folder_id": rootFolderId}

	}

	if parentFolderId != "" {
		whereMap = sq.And{whereMap, sq.Eq{"folder_id": parentFolderId}}
	}

	if query != "" {
		whereMap = sq.And{whereMap, sq.Like{"title": "%" + query + "%"}}

	}

	sql, args, err := r.qb.Select("id").
		Columns("title", "folder_id", "folder_name", "parent_folder_id", "size", "size_bytes", "duration", "modified_at", "codec", "resolution", "poster_url", "stream_url", "path").
		From(videosTable).
		Where(whereMap).
		Limit(limit).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query: %w", err)
	}

	var rows []VideoRow

	err = r.conn.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	var results []entity.Video
	for _, row := range rows {
		results = append(results, entity.Video{
			Id:             row.Id,
			Title:          row.Title,
			FolderId:       row.FolderId,
			FolderName:     row.FolderName,
			ParentFolderId: row.ParentFolderId,
			Size:           row.Size,
			SizeBytes:      row.SizeBytes,
			Duration:       row.Duration,
			ModifiedAt:     row.ModifiedAt,
			Codec:          row.Codec,
			Resolution:     row.Resolution,
			PosterUrl:      row.PosterUrl,
			StreamUrl:      row.StreamUrl,
			Path:           row.Path,
		})
	}

	return results, nil
}

func (r *Repository) CreateVideo(ctx context.Context, video entity.Video) (entity.Video, error) {
	var res entity.Video
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createVideoTx(ctx, video, tx)
		if err != nil {
			return err
		}
		return err
	})

	if txErr != nil {
		return entity.Video{}, txErr
	}

	return res, nil
}

func (r *Repository) createVideoTx(ctx context.Context, video entity.Video, tx *sqlx.Tx) (entity.Video, error) {
	insertMap := map[string]any{
		"id":               video.Id,
		"title":            video.Title,
		"folder_id":        video.FolderId,
		"folder_name":      video.FolderName,
		"parent_folder_id": video.ParentFolderId,
		"size":             video.Size,
		"size_bytes":       video.SizeBytes,
		"duration":         video.Duration,
		"codec":            video.Codec,
		"modified_at":      video.ModifiedAt,
		"resolution":       video.Resolution,
		"poster_url":       video.PosterUrl,
		"stream_url":       video.StreamUrl,
		"path":             video.Path,
	}

	sql, args, err := r.qb.Insert(videosTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to building query: %w", err)
	}

	var row VideoRow
	var result entity.Video

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to executing query: %w", err)
	}

	result = entity.Video{
		Id:             row.Id,
		Title:          row.Title,
		FolderId:       row.FolderId,
		FolderName:     row.FolderName,
		ParentFolderId: row.ParentFolderId,
		Size:           row.Size,
		SizeBytes:      row.SizeBytes,
		Duration:       row.Duration,
		Codec:          row.Codec,
		Resolution:     row.Resolution,
		PosterUrl:      row.PosterUrl,
		StreamUrl:      row.StreamUrl,
		Path:           row.Path,
		ModifiedAt:     row.ModifiedAt,
	}

	return result, nil
}

func (r *Repository) GetVideoById(ctx context.Context, id string) (entity.Video, error) {
	whereMap := map[string]any{
		"id": id,
	}
	sql, args, err := r.qb.Select("id").
		Columns("title", "folder_id", "folder_name", "parent_folder_id", "size", "size_bytes", "duration", "modified_at", "codec", "resolution", "poster_url", "stream_url", "path").
		From(videosTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to building query: %w", err)
	}

	var row VideoRow

	err = r.conn.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to executing query: %w", err)
	}

	return entity.Video{
		Id:             row.Id,
		Title:          row.Title,
		FolderId:       row.FolderId,
		FolderName:     row.FolderName,
		ParentFolderId: row.ParentFolderId,
		Size:           row.Size,
		SizeBytes:      row.SizeBytes,
		Duration:       row.Duration,
		ModifiedAt:     row.ModifiedAt,
		Codec:          row.Codec,
		Resolution:     row.Resolution,
		PosterUrl:      row.PosterUrl,
		StreamUrl:      row.StreamUrl,
		Path:           row.Path,
	}, nil
}
