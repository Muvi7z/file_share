package entity

import "time"

type File struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	FolderId       string    `json:"folderId"`
	ParentFolderId string    `json:"parentFolderId"`
	Extension      string    `json:"extension"`
	Size           string    `json:"size"`
	SizeBytes      int64     `json:"sizeBytes"`
	ModifiedAt     time.Time `json:"modifiedAt"`
	MimeType       string    `json:"mimeType"`
}
