package entity

import "time"

type File struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Path           string    `json:"path"`
	FolderId       string    `json:"folderId"`
	FolderName     string    `json:"folderName"`
	ParentFolderId string    `json:"parentFolderId"`
	Extension      string    `json:"extension"`
	Size           string    `json:"size"`
	SizeBytes      int64     `json:"sizeBytes"`
	ModifiedAt     time.Time `json:"modifiedAt"`
	MimeType       string    `json:"mimeType"`
}

var AllowedImageExts = map[string]string{
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
}
