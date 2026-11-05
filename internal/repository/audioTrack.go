package repository

import (
	"context"
	"database/sql"
	"errors"
	"file_share/internal/entity"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const audioTrackTable = "audio_track"

type AudioTrackRow struct {
	Id          string         `db:"id"`
	Title       string         `db:"title"`
	VideoId     sql.NullString `db:"video_id"`
	StreamIndex sql.NullInt64  `db:"stream_index"`
	Language    sql.NullString `db:"language"`
	Codec       sql.NullString `db:"codec"`
	Channels    sql.NullInt64  `db:"channels"`
	IsDefault   sql.NullBool   `db:"is_default"`
	Status      sql.NullString `db:"status"`
}

func (row AudioTrackRow) toEntity() entity.AudioTrack {
	return entity.AudioTrack{
		Id:          row.Id,
		Title:       row.Title,
		VideoId:     row.VideoId.String,
		StreamIndex: int(row.StreamIndex.Int64),
		Language:    row.Language.String,
		Codec:       row.Codec.String,
		Channels:    int(row.Channels.Int64),
		IsDefault:   row.IsDefault.Bool,
		Status:      row.Status.String,
	}
}

func (r *Repository) GetAllAudioTrack(ctx context.Context, videoId string, limit uint64, offset uint64) ([]entity.AudioTrack, error) {
	var whereMap sq.Sqlizer
	if videoId != "" {
		whereMap = sq.Eq{"video_id": videoId}
	}

	querySql := r.qb.Select("*").
		From(audioTrackTable).
		Where(whereMap).
		OrderBy("stream_index", "id")
	if limit > 0 {
		querySql = querySql.Limit(limit)
	}
	if offset > 0 {
		querySql = querySql.Offset(offset)
	}

	query, args, err := querySql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query: %w", err)
	}

	var rows []AudioTrackRow
	if err = r.conn.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	var results []entity.AudioTrack
	for _, row := range rows {
		results = append(results, row.toEntity())
	}
	return results, nil
}

func (r *Repository) GetAudioTrackById(ctx context.Context, id string) (entity.AudioTrack, error) {
	query, args, err := r.qb.Select("*").
		From(audioTrackTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return entity.AudioTrack{}, fmt.Errorf("error to building query: %w", err)
	}

	var row AudioTrackRow
	if err = r.conn.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AudioTrack{}, entity.ErrorNotFoundAudioTrack
		}
		return entity.AudioTrack{}, fmt.Errorf("error to executing query: %w", err)
	}
	return row.toEntity(), nil
}

func (r *Repository) CreateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack) (entity.AudioTrack, error) {
	var res entity.AudioTrack
	var err error
	txErr := sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createAudioTrackTx(ctx, audioTrack, tx)
		return err
	})
	if txErr != nil {
		return entity.AudioTrack{}, txErr
	}
	return res, nil
}

func (r *Repository) CreateAudioTracks(ctx context.Context, audioTracks []entity.AudioTrack) ([]entity.AudioTrack, error) {
	if len(audioTracks) == 0 {
		return []entity.AudioTrack{}, nil
	}

	var res []entity.AudioTrack
	var err error
	txErr := sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createAudioTracksTx(ctx, audioTracks, tx)
		return err
	})
	if txErr != nil {
		return nil, txErr
	}
	return res, nil
}

func (r *Repository) createAudioTracksTx(ctx context.Context, audioTracks []entity.AudioTrack, tx *sqlx.Tx) ([]entity.AudioTrack, error) {
	querySql := r.qb.Insert(audioTrackTable).
		Columns("id", "title", "video_id", "stream_index", "language", "codec", "channels", "is_default", "status")
	for _, audioTrack := range audioTracks {
		querySql = querySql.Values(
			audioTrack.Id,
			audioTrack.Title,
			audioTrack.VideoId,
			audioTrack.StreamIndex,
			audioTrack.Language,
			audioTrack.Codec,
			audioTrack.Channels,
			audioTrack.IsDefault,
			audioTrack.Status,
		)
	}

	query, args, err := querySql.Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query: %w", err)
	}

	var rows []AudioTrackRow
	if err = tx.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	results := make([]entity.AudioTrack, 0, len(rows))
	for _, row := range rows {
		results = append(results, row.toEntity())
	}
	return results, nil
}

func (r *Repository) createAudioTrackTx(ctx context.Context, audioTrack entity.AudioTrack, tx *sqlx.Tx) (entity.AudioTrack, error) {
	insertMap := map[string]any{
		"id":           audioTrack.Id,
		"title":        audioTrack.Title,
		"video_id":     audioTrack.VideoId,
		"stream_index": audioTrack.StreamIndex,
		"language":     audioTrack.Language,
		"codec":        audioTrack.Codec,
		"channels":     audioTrack.Channels,
		"is_default":   audioTrack.IsDefault,
		"status":       audioTrack.Status,
	}
	query, args, err := r.qb.Insert(audioTrackTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.AudioTrack{}, fmt.Errorf("error to building query: %w", err)
	}

	var row AudioTrackRow
	if err = tx.GetContext(ctx, &row, query, args...); err != nil {
		return entity.AudioTrack{}, fmt.Errorf("error to executing query: %w", err)
	}
	return row.toEntity(), nil
}

// UpdateAudioTrack replaces all fields except the track ID.
func (r *Repository) UpdateAudioTrack(ctx context.Context, audioTrack entity.AudioTrack, id string) (entity.AudioTrack, error) {
	var res entity.AudioTrack
	var err error
	txErr := sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateAudioTrackTx(ctx, audioTrack, id, tx)
		return err
	})
	if txErr != nil {
		return entity.AudioTrack{}, txErr
	}
	return res, nil
}

func (r *Repository) updateAudioTrackTx(ctx context.Context, audioTrack entity.AudioTrack, id string, tx *sqlx.Tx) (entity.AudioTrack, error) {
	updateMap := map[string]any{
		"title":        audioTrack.Title,
		"video_id":     audioTrack.VideoId,
		"stream_index": audioTrack.StreamIndex,
		"language":     audioTrack.Language,
		"codec":        audioTrack.Codec,
		"channels":     audioTrack.Channels,
		"is_default":   audioTrack.IsDefault,
		"status":       audioTrack.Status,
	}
	query, args, err := r.qb.Update(audioTrackTable).
		SetMap(updateMap).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.AudioTrack{}, fmt.Errorf("error to building query: %w", err)
	}

	var row AudioTrackRow
	if err = tx.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AudioTrack{}, entity.ErrorNotFoundAudioTrack
		}
		return entity.AudioTrack{}, fmt.Errorf("error to update: %w", err)
	}
	return row.toEntity(), nil
}

func (r *Repository) DeleteAudioTrack(ctx context.Context, id string) error {
	return sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		return r.deleteAudioTrackByIdTx(ctx, id, tx)
	})
}

func (r *Repository) deleteAudioTrackByIdTx(ctx context.Context, id string, tx *sqlx.Tx) error {
	query, args, err := r.qb.Delete(audioTrackTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error to building query: %w", err)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error to executing query: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error to executing query: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrorNotFoundAudioTrack
	}
	return nil
}
