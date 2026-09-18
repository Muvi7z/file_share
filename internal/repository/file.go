package repository

import "time"

type FileRow struct {
	Id             string    `db:"id"`
	Name           string    `db:"name"`
	Path           string    `db:"path"`
	Extension      string    `db:"extension"`
	FolderId       string    `db:"folder_id"`
	ParentFolderId string    `db:"parent_folder_id"`
	Size           string    `db:"size"`
	SizeBytes      int64     `db:"size_bytes"`
	ModifiedAt     time.Time `db:"modified_at"`
}

const fileTable = "file"
