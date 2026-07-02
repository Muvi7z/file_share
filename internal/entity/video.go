package entity

import "time"

type Video struct {
	Id             string    `json:"id"`
	Title          string    `json:"title"`
	FolderId       string    `json:"folderId"`
	FolderName     string    `json:"folderName"`
	ParentFolderId string    `json:"parentFolderId"`
	Size           string    `json:"size"`
	SizeBytes      int64     `json:"sizeBytes"`
	Duration       string    `json:"duration"`
	ModifiedAt     time.Time `json:"modifiedAt"`
	Codec          string    `json:"codec"`
	Resolution     string    `json:"resolution"`
	PosterUrl      string    `json:"posterUrl"`
	StreamUrl      string    `json:"streamUrl"`
	Path           string    `json:"path"`
}
