package repository

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type FolderRow struct {
	Id               string    `db:"id"`
	Name             string    `db:"name"`
	Path             string    `db:"path"`
	ParentId         string    `db:"parent_id"`
	RootFolderId     string    `db:"root_folder_id"`
	IsRoot           bool      `db:"is_root"`
	Enabled          bool      `db:"enabled"`
	FilesCount       int       `db:"files_count"`
	VideoCount       int       `db:"video_count"`
	ChildFolderCount int       `db:"child_folder_count"`
	LastScanAt       time.Time `db:"last_scan_at"`
}

const folderTable = "folders"

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
		return entity.Folder{}, fmt.Errorf("error executing query: %w", err)
	}

	return entity.Folder{
		Id:               row.Id,
		Name:             row.Name,
		Path:             row.Path,
		ParentId:         row.ParentId,
		RootFolderId:     row.RootFolderId,
		IsRoot:           row.IsRoot,
		Enabled:          row.Enabled,
		FilesCount:       row.FilesCount,
		VideosCount:      row.VideoCount,
		ChildFolderCount: row.ChildFolderCount,
		LastScanAt:       row.LastScanAt,
	}, err
}

func (r *Repository) GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string) ([]entity.Folder, error) {
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
		Columns("name", "path", "parent_id", "root_folder_id", "is_root", "enabled", "files_count", "video_count", "child_folder_count", "last_scan_at").
		Where(whereMap).
		From(folderTable).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to building query %v", err)
	}

	var rows []FolderRow

	err = r.conn.SelectContext(ctx, &args, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error  to executing query %v", err)
	}

	var result []entity.Folder
	for _, row := range rows {
		result = append(result, entity.Folder{
			Id:               row.Id,
			Name:             row.Name,
			Path:             row.Path,
			ParentId:         row.ParentId,
			RootFolderId:     row.RootFolderId,
			IsRoot:           row.IsRoot,
			Enabled:          row.Enabled,
			FilesCount:       row.FilesCount,
			VideosCount:      row.VideoCount,
			ChildFolderCount: row.ChildFolderCount,
			LastScanAt:       row.LastScanAt,
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
		"id":                 folder.Id,
		"name":               folder.Name,
		"path":               folder.Path,
		"parent_id":          folder.ParentId,
		"root_folder_id":     folder.RootFolderId,
		"is_root":            folder.IsRoot,
		"enabled":            folder.Enabled,
		"files_count":        folder.FilesCount,
		"video_count":        folder.VideosCount,
		"child_folder_count": folder.ChildFolderCount,
		"last_scan_at":       nil,
	}

	sql, args, err := r.qb.Insert(videosTable).
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

	result = entity.Folder{
		Id:               row.Id,
		Name:             row.Name,
		Path:             row.Path,
		ParentId:         row.ParentId,
		RootFolderId:     row.RootFolderId,
		IsRoot:           row.IsRoot,
		Enabled:          row.Enabled,
		FilesCount:       row.FilesCount,
		VideosCount:      row.VideoCount,
		ChildFolderCount: row.ChildFolderCount,
		LastScanAt:       row.LastScanAt,
	}

	return result, nil
}
