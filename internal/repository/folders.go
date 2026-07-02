package repository

import (
	"context"
	"file_share/internal/entity"
	"fmt"
)

type FolderRow struct {
	Id               string `db:"id"`
	Name             string `db:"name"`
	Path             string `db:"path"`
	ParentId         string `db:"parent_id"`
	RootFolderId     string `db:"root_folder_id"`
	IsRoot           bool   `db:"is_root"`
	Enabled          bool   `db:"enabled"`
	FilesCount       int    `db:"files_count"`
	VideoCount       int    `db:"video_count"`
	ChildFolderCount int    `db:"child_folder_count"`
	LastScanAt       string `db:"last_scan_at"`
}

const folderTable = "folders"

func (r *Repository) GetFolders(ctx context.Context) ([]entity.Folder, error) {
	sql, args, err := r.qb.Select("id").
		Columns("name", "path", "parent_id", "root_folder_id", "is_root", "enabled", "files_count", "video_count", "child_folder_count", "last_scan_at").
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
