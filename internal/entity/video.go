package entity

import "time"

const (
	MKV = ".mkv"
)

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
	Extension      string    `json:"extension"`
}

type AudioTrack struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	VideoId     string `json:"videoId"`
	StreamIndex int    `json:"streamIndex"`
	Language    string `json:"language"`
	Codec       string `json:"codec"`
	Channels    int    `json:"channels"`
	IsDefault   bool   `json:"isDefault"`
	Status      string `json:"status"`
}

type AudioTrackInfo struct {
	AudioTrackStream []AudioTrackStream `json:"streams"`
}

type AudioTrackStream struct {
	Index       int                   `json:"index"`
	CodecName   string                `json:"codec_name"`
	CodecType   string                `json:"codec_type"`
	Disposition AudioTrackDisposition `json:"disposition"`
	Tags        AudioTrackTags        `json:"tags"`
}

type AudioTrackDisposition struct {
	Default int `json:"default"`
}

type AudioTrackTags struct {
	Language string `json:"language"`
	Title    string `json:"title"`
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

const (
	AudioTrackPending    = "pending"
	AudioTrackProcessing = "processing"
	AudioTrackReady      = "ready"
	AudioTrackFailed     = "failed"
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
