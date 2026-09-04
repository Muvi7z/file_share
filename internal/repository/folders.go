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

type FolderRow struct {
	Id               string     `db:"id"`
	Name             string     `db:"name"`
	Path             string     `db:"path"`
	ParentId         *string    `db:"parent_id"`
	RootFolderId     string     `db:"root_folder_id"`
	IsRoot           bool       `db:"is_root"`
	Enabled          bool       `db:"enabled"`
	FilesCount       int        `db:"files_count"`
	VideoCount       int        `db:"video_count"`
	ChildFolderCount int        `db:"child_folder_count"`
	LastScanAt       *time.Time `db:"last_scan_at"`
}

const folderTable = "folder"

func (r *Repository) GetFolderById(ctx context.Context, id string) (entity.Folder, error) {
	whereMap := map[string]any{
		"id": id,
	}
	//30

	sql, args, err := r.qb.Select("id").
		Columns("name", "path", "parent_id", "root_folder_id", "is_root", "enabled", "files_count", "video_count", "child_folder_count", "last_scan_at").
		From(folderTable).
		Where(whereMap).
		ToSql()
	if err != nil {
		return entity.Folder{}, fmt.Errorf("error building query: %w", err)
	}

	var row FolderRow
	err = r.conn.GetContext(ctx, &row, sql, args...)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Folder{}, entity.ErrorNotFoundFolder
		}
		return entity.Folder{}, fmt.Errorf("error executing query: %w", err)
	}

	var parentId string

	if row.ParentId != nil {
		parentId = *row.ParentId
	}

	var lastScanAt time.Time

	if row.LastScanAt != nil {
		lastScanAt = *row.LastScanAt
	}

	return entity.Folder{
		Id:               row.Id,
		Name:             row.Name,
		Path:             row.Path,
		ParentId:         parentId,
		RootFolderId:     row.RootFolderId,
		IsRoot:           row.IsRoot,
		Enabled:          row.Enabled,
		FilesCount:       row.FilesCount,
		VideosCount:      row.VideoCount,
		ChildFolderCount: row.ChildFolderCount,
		LastScanAt:       lastScanAt,
	}, err
}

func (r *Repository) GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot, enabled *bool) ([]entity.Folder, error) {
	var whereMap sq.Sqlizer

	if rootFolderId != "" {
		whereMap = sq.Eq{"root_folder_id": rootFolderId}

	}

	if isRoot != nil {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Eq{"is_root": isRoot}}
		} else {
			whereMap = sq.Eq{"is_root": isRoot}
		}
	}

	if enabled != nil {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Eq{"enabled": enabled}}
		} else {
			whereMap = sq.Eq{"enabled": enabled}
		}
	}

	if parentFolderId != "" {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Eq{"parent_id": parentFolderId}}
		} else {
			whereMap = sq.Eq{"parent_id": parentFolderId}
		}
	}

	if query != "" {
		if whereMap != nil {
			whereMap = sq.And{whereMap, sq.Like{"title": "%" + query + "%"}}
		} else {
			whereMap = sq.Like{"title": "%" + query + "%"}
		}

	}

	selectB := r.qb.Select("id").
		Columns("name", "path", "parent_id", "root_folder_id", "is_root", "enabled", "files_count", "video_count", "child_folder_count", "last_scan_at").
		From(folderTable)

	if whereMap != nil {
		selectB = selectB.Where(whereMap)
	}

	sql, args, err := selectB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query %v", err)
	}

	var rows []FolderRow

	err = r.conn.SelectContext(ctx, &rows, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error  to executing query %v", err)
	}

	var result []entity.Folder
	for _, row := range rows {

		var parentId string

		if row.ParentId != nil {
			parentId = *row.ParentId
		}

		var lastScanAt time.Time

		if row.LastScanAt != nil {
			lastScanAt = *row.LastScanAt
		}

		result = append(result, entity.Folder{
			Id:               row.Id,
			Name:             row.Name,
			Path:             row.Path,
			ParentId:         parentId,
			RootFolderId:     row.RootFolderId,
			IsRoot:           row.IsRoot,
			Enabled:          row.Enabled,
			FilesCount:       row.FilesCount,
			VideosCount:      row.VideoCount,
			ChildFolderCount: row.ChildFolderCount,
			LastScanAt:       lastScanAt,
		})
	}

	return result, nil
}

func (r *Repository) CreateFolder(ctx context.Context, folder entity.Folder) (entity.Folder, error) {
	var res entity.Folder
	var err, txErr error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.createFolderTx(ctx, folder, tx)
		if err != nil {
			return err
		}
		return err
	})

	if txErr != nil {
		return entity.Folder{}, txErr
	}

	return res, nil
}

func (r *Repository) createFolderTx(ctx context.Context, folder entity.Folder, tx *sqlx.Tx) (entity.Folder, error) {
	insertMap := map[string]any{
		"id":   folder.Id,
		"name": folder.Name,
		"path": folder.Path,

		"is_root":            folder.IsRoot,
		"enabled":            folder.Enabled,
		"files_count":        folder.FilesCount,
		"video_count":        folder.VideosCount,
		"child_folder_count": folder.ChildFolderCount,
		"last_scan_at":       nil,
	}

	if folder.RootFolderId != "" {
		insertMap["root_folder_id"] = folder.RootFolderId

	}

	if folder.ParentId != "" {
		insertMap["parent_id"] = folder.ParentId
	}

	sql, args, err := r.qb.Insert(folderTable).
		SetMap(insertMap).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return entity.Folder{}, fmt.Errorf("error to building query: %w", err)
	}

	var row FolderRow
	var result entity.Folder

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		return entity.Folder{}, fmt.Errorf("error to executing query: %w", err)
	}

	var parentId string

	if row.ParentId != nil {
		parentId = *row.ParentId
	}

	var lastScanAt time.Time

	if row.LastScanAt != nil {
		lastScanAt = *row.LastScanAt
	}

	result = entity.Folder{
		Id:               row.Id,
		Name:             row.Name,
		Path:             row.Path,
		ParentId:         parentId,
		RootFolderId:     row.RootFolderId,
		IsRoot:           row.IsRoot,
		Enabled:          row.Enabled,
		FilesCount:       row.FilesCount,
		VideosCount:      row.VideoCount,
		ChildFolderCount: row.ChildFolderCount,
		LastScanAt:       lastScanAt,
	}

	return result, nil
}
func (r *Repository) UpdateFolder(ctx context.Context, folder entity.UpdateFolderRequest, id string) (entity.Folder, error) {
	var res entity.Folder
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		res, err = r.updateFolderTx(ctx, folder, id, tx)
		if err != nil {
			return err
		}

		return err
	})

	if txErr != nil {
		return entity.Folder{}, txErr
	}

	return res, nil
}

func (r *Repository) updateFolderTx(ctx context.Context, folder entity.UpdateFolderRequest, id string, tx *sqlx.Tx) (entity.Folder, error) {
	updateMap := map[string]any{}

	if folder.Name != "" {
		updateMap["name"] = folder.Name
	}

	if folder.Enabled != nil {
		updateMap["enabled"] = *folder.Enabled
	}

	if folder.FilesCount != nil {
		updateMap["files_count"] = *folder.FilesCount
	}

	if folder.VideosCount != nil {
		updateMap["video_count"] = *folder.VideosCount
	}

	if folder.ChildFolderCount != nil {
		updateMap["child_folder_count"] = *folder.ChildFolderCount
	}

	sql, args, err := r.qb.Update(folderTable).
		SetMap(updateMap).
		Suffix("RETURNING *").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return entity.Folder{}, fmt.Errorf("error to building query: %w", err)
	}

	var row FolderRow

	err = tx.GetContext(ctx, &row, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Folder{}, entity.ErrorNoRowsFound
		}
		return entity.Folder{}, fmt.Errorf("error to executing query: %w", err)
	}
	var parentId string

	if row.ParentId != nil {
		parentId = *row.ParentId
	}

	var lastScanAt time.Time

	if row.LastScanAt != nil {
		lastScanAt = *row.LastScanAt
	}

	return entity.Folder{
		Id:               row.Id,
		Name:             row.Name,
		Path:             row.Path,
		ParentId:         parentId,
		RootFolderId:     row.RootFolderId,
		Enabled:          row.Enabled,
		FilesCount:       row.FilesCount,
		VideosCount:      row.VideoCount,
		ChildFolderCount: row.ChildFolderCount,
		LastScanAt:       lastScanAt,
		IsRoot:           row.IsRoot,
	}, nil

}

func (r *Repository) DeleteFolder(ctx context.Context, id string) error {
	var txErr, err error

	txErr = sqlxTransaction(ctx, r.conn, func(tx *sqlx.Tx) error {
		err = r.deleteFolderByIdTx(ctx, id, tx)
		if err != nil {
			return err
		}

		err = r.deleteVideoByFolderTx(ctx, id, "", tx)
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

func (r *Repository) deleteFolderByIdTx(ctx context.Context, id string, tx *sqlx.Tx) error {
	sql, args, err := r.qb.Delete(folderTable).Where(sq.Eq{"id": id}).ToSql()
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
