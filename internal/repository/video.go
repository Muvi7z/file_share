package repository

import (
	"context"
	"database/sql"
	"errors"
	"file_share/internal/entity"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

type VideoRow struct {
	Id             string         `db:"id"`
	Title          string         `db:"title"`
	FolderId       string         `db:"folder_id"`
	FolderName     string         `db:"folder_name"`
	ParentFolderId string         `db:"parent_folder_id"`
	Size           string         `db:"size"`
	SizeBytes      int64          `db:"size_bytes"`
	Duration       string         `db:"duration"`
	ModifiedAt     time.Time      `db:"modified_at"`
	Codec          string         `db:"codec"`
	Resolution     string         `db:"resolution"`
	PosterUrl      string         `db:"poster_url"`
	StreamUrl      string         `db:"stream_url"`
	Extension      sql.NullString `db:"extension"`
	Path           string         `db:"path"`
}

const videosTable = "video"

func (r *Repository) GetAllVideo(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64, enabled *bool) ([]entity.Video, error) {
	var whereMap sq.Sqlizer

	if rootFolderId != "" {
		whereMap = sq.Eq{"folder_id": rootFolderId}
	}

	if parentFolderId != "" {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Eq{"parent_folder_id": parentFolderId}}
		} else {
			whereMap = sq.Eq{"parent_folder_id": parentFolderId}
		}
	}

	if query != "" {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Expr("lower(title) LIKE lower(?)", "%"+query+"%")}
		} else {
			whereMap = sq.Expr("lower(title) LIKE lower(?)", "%"+query+"%")
		}

	}

	querySql := r.qb.Select("v.id").
		Columns("v.title", "v.folder_id", "v.folder_name", "v.parent_folder_id", "v.size", "v.size_bytes", "v.duration", "v.modified_at", "v.codec", "v.resolution", "v.poster_url", "v.stream_url", "v.path").
		From(fmt.Sprintf("%s v", videosTable)).
		Where(whereMap).
		Offset(offset)

	if enabled != nil {
		querySql = querySql.Join("folder f on f.id = v.folder_id").Where(sq.Eq{"f.enabled": *enabled})
	}

	sql, args, err := querySql.ToSql()
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
			Extension:      row.Extension.String,
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

func (r *Repository) UpdateVideo(ctx context.Context, video entity.Video, id string) (entity.Video, error) {
	var res entity.Video
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateVideoTx(ctx, video, id, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return entity.Video{}, txErr
	}

	return res, nil
}

func (r *Repository) updateVideoTx(ctx context.Context, video entity.Video, id string, tx *sqlx.Tx) (entity.Video, error) {
	updateMap := map[string]any{}

	if video.Path != "" {
		updateMap["path"] = video.Path
	}

	if video.Title != "" {
		updateMap["title"] = video.Title
	}

	if video.PosterUrl != "" {
		updateMap["poster_url"] = video.PosterUrl
	}

	if video.StreamUrl != "" {
		updateMap["stream_url"] = video.StreamUrl
	}

	sql, args, err := r.qb.Update(videosTable).
		SetMap(updateMap).
		Suffix("RETURNING *").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to building query: %w", err)
	}

	var row VideoRow

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Video{}, err
		}

		return entity.Video{}, fmt.Errorf("error to update: %w", err)
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
		Extension:      row.Extension.String,
		Path:           row.Path,
	}, nil
}

func (r *Repository) DeleteVideo(ctx context.Context, id string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteVideoByIdTx(ctx, id, tx)
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

func (r *Repository) DeleteVideoByFolder(ctx context.Context, idRootFolder, parentFolderId string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteVideoByFolderTx(ctx, idRootFolder, parentFolderId, tx)
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

func (r *Repository) deleteVideoByFolderTx(ctx context.Context, idRootFolder, parentFolderId string, tx *sqlx.Tx) error {
	var whereMap sq.Sqlizer

	if idRootFolder != "" {
		whereMap = sq.Eq{"folder_id": idRootFolder}
	}

	if parentFolderId != "" {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Eq{"parent_folder_id": parentFolderId}}
		} else {
			whereMap = sq.Eq{"parent_folder_id": parentFolderId}
		}
	}

	sql, args, err := r.qb.Delete(videosTable).Where(whereMap).ToSql()
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

func (r *Repository) deleteVideoByIdTx(ctx context.Context, id string, tx *sqlx.Tx) error {
	sql, args, err := r.qb.Delete(videosTable).Where(sq.Eq{"id": id}).ToSql()
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
		"extension":        video.Extension,
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
		Extension:      row.Extension.String,
	}

	return result, nil
}

func (r *Repository) GetVideoById(ctx context.Context, id string) (entity.Video, error) {
	whereMap := map[string]any{
		"id": id,
	}
	sql, args, err := r.qb.Select("*").
		From(videosTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.Video{}, fmt.Errorf("error to building query: %w", err)
	}

	var row VideoRow

	err = r.conn.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Video{}, entity.ErrorNotFoundVideo
		}
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
		Extension:      row.Extension.String,
	}, nil
}
