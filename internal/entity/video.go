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

type VideoStream struct {
	FileName string
	Size     int64
	ModTime  time.Time
	Reader   ReadSeekCloser
}

type PosterFile struct {
	VideoId     string
	FileName    string
	ContentType string
	ModTime     time.Time
	Reader      ReadSeekCloser
	Path        string
}

type ReadSeekCloser interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

const (
	RepostTypeOpen     = "open"
	RepostTypeAuto     = "auto"
	RepostTypeResolved = "resolved"
)

type ReportVideo struct {
	Id             string    `json:"id"`
	VideoId        string    `json:"videoId"`
	Title          string    `json:"title"`
	Reason         string    `json:"reason"`
	PositionSecond int64     `json:"positionSecond"`
	Comment        string    `json:"comment"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateReportVideo struct {
	VideoId        string    `json:"videoId"`
	Title          string    `json:"title"`
	Reason         string    `json:"reason"`
	PositionSecond int64     `json:"positionSecond"`
	Comment        string    `json:"comment"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
}
