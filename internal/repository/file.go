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

type FileRow struct {
	Id             string    `db:"id"`
	Name           string    `db:"name"`
	Path           string    `db:"path"`
	Extension      string    `db:"extension"`
	FolderId       string    `db:"folder_id"`   //корневая папка
	FolderName     string    `db:"folder_name"` // название корневой папки
	ParentFolderId string    `db:"parent_folder_id"`
	Size           string    `db:"size"`
	SizeBytes      int64     `db:"size_bytes"`
	ModifiedAt     time.Time `db:"modified_at"`
}

const fileTable = "file"

func (r *Repository) GetAllFile(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.File, error) {
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

	sql, args, err := r.qb.Select("id").
		Columns("name", "path", "extension", "folder_id", "parent_folder_id", "size", "size_bytes", "modified_at").
		From(fileTable).
		Where(whereMap).
		Offset(offset).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query: %w", err)
	}

	var rows []FileRow

	err = r.conn.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error to executing query: %w", err)
	}

	var results []entity.File
	for _, row := range rows {
		results = append(results, entity.File{
			Id:             row.Id,
			Name:           row.Name,
			FolderId:       row.FolderId,
			FolderName:     row.FolderName,
			ParentFolderId: row.ParentFolderId,
			Size:           row.Size,
			SizeBytes:      row.SizeBytes,
			ModifiedAt:     row.ModifiedAt,
			Path:           row.Path,
			Extension:      row.Extension,
		})
	}

	return results, nil
}

func (r *Repository) CreateFile(ctx context.Context, file entity.File) (entity.File, error) {
	var res entity.File
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createFileTx(ctx, file, tx)
		if err != nil {
			return err
		}
		return err
	})

	if txErr != nil {
		return entity.File{}, txErr
	}

	return res, nil
}

func (r *Repository) createFileTx(ctx context.Context, file entity.File, tx *sqlx.Tx) (entity.File, error) {
	insertMap := map[string]any{
		"id":               file.Id,
		"name":             file.Name,
		"path":             file.Path,
		"extension":        file.Extension,
		"folder_id":        file.FolderId,
		"folder_name":      file.FolderName,
		"parent_folder_id": file.ParentFolderId,
		"size":             file.Size,
		"size_bytes":       file.SizeBytes,
		"modified_at":      file.ModifiedAt,
	}

	sql, args, err := r.qb.Insert(fileTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.File{}, fmt.Errorf("error to building query: %w", err)
	}

	var row FileRow
	var result entity.File

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.File{}, fmt.Errorf("error to executing query: %w", err)
	}

	result = entity.File{
		Id:             row.Id,
		Name:           row.Name,
		FolderId:       row.FolderId,
		FolderName:     row.FolderName,
		ParentFolderId: row.ParentFolderId,
		Size:           row.Size,
		SizeBytes:      row.SizeBytes,
		ModifiedAt:     row.ModifiedAt,
		Path:           row.Path,
		Extension:      row.Extension,
	}

	return result, nil
}

func (r *Repository) UpdateFile(ctx context.Context, file entity.File, id string) (entity.File, error) {
	var res entity.File
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateFileTx(ctx, file, id, tx)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return entity.File{}, txErr
	}

	return res, nil
}

func (r *Repository) updateFileTx(ctx context.Context, file entity.File, id string, tx *sqlx.Tx) (entity.File, error) {
	updateMap := map[string]any{}

	if file.Path != "" {
		updateMap["path"] = file.Path
	}

	if file.Name != "" {
		updateMap["name"] = file.Name
	}

	sql, args, err := r.qb.Update(fileTable).
		SetMap(updateMap).
		Suffix("RETURNING *").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return entity.File{}, fmt.Errorf("error to building query: %w", err)
	}

	var row FileRow

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.File{}, err
		}

		return entity.File{}, fmt.Errorf("error to update: %w", err)
	}

	return entity.File{
		Id:             row.Id,
		Name:           row.Name,
		FolderId:       row.FolderId,
		FolderName:     row.FolderName,
		ParentFolderId: row.ParentFolderId,
		Size:           row.Size,
		SizeBytes:      row.SizeBytes,
		ModifiedAt:     row.ModifiedAt,
		Path:           row.Path,
		Extension:      row.Extension,
	}, nil
}

func (r *Repository) DeleteFile(ctx context.Context, id string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteFileByIdTx(ctx, id, tx)
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

func (r *Repository) deleteFileByIdTx(ctx context.Context, id string, tx *sqlx.Tx) error {
	sql, args, err := r.qb.Delete(fileTable).Where(sq.Eq{"id": id}).ToSql()
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

func (r *Repository) deleteFileByFolderTx(ctx context.Context, idRootFolder, parentFolderId string, tx *sqlx.Tx) error {
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

	sql, args, err := r.qb.Delete(fileTable).Where(whereMap).ToSql()
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

func (r *Repository) GetFileById(ctx context.Context, id string) (entity.File, error) {
	whereMap := map[string]any{
		"id": id,
	}
	sql, args, err := r.qb.Select("id").
		Columns("name", "path", "extension", "folder_id", "parent_folder_id", "size", "size_bytes", "modified_at").
		From(fileTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.File{}, fmt.Errorf("error to building query: %w", err)
	}

	var row FileRow

	err = r.conn.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.File{}, entity.ErrorNotFoundVideo
		}
		return entity.File{}, fmt.Errorf("error to executing query: %w", err)
	}

	return entity.File{
		Id:             row.Id,
		Name:           row.Name,
		FolderId:       row.FolderId,
		FolderName:     row.FolderName,
		ParentFolderId: row.ParentFolderId,
		Size:           row.Size,
		SizeBytes:      row.SizeBytes,
		ModifiedAt:     row.ModifiedAt,
		Path:           row.Path,
		Extension:      row.Extension,
	}, nil
}
