package entity

import "time"

const (
	StatusQueued    string = "queued"
	StatusRunning   string = "running"
	StatusCompleted string = "completed"
	StatusFailed    string = "failed"
)

type ScanJob struct {
	Id               string    `json:"id"`
	FolderId         string    `json:"folderId"`
	Status           string    `json:"status"`
	ProcessedVideos  int       `json:"processedVideos"`
	ProcessedFolders int       `json:"processedFolders"`
	StartedAt        time.Time `json:"startedAt"`
	FinishedAt       time.Time `json:"finishedAt"`
	Error            string    `json:"error"`
}

type Folder struct {
	Id               string    `json:"id"`
	Name             string    `json:"name"`
	Path             string    `json:"path"`
	ParentId         string    `json:"parentId"`
	RootFolderId     string    `json:"rootFolderId"`
	IsRoot           bool      `json:"isRoot"`
	Enabled          bool      `json:"enabled"`
	FilesCount       int       `json:"filesCount"`
	VideosCount      int       `json:"videosCount"`
	ChildFolderCount int       `json:"childFolderCount"`
	LastScanAt       time.Time `json:"lastScanAt"`
}

type FileBrowserEntry struct {
	Type   string `json:"type"`
	Folder Folder `json:"folder"`
	Video  Video  `json:"video"`
}

type CreateRootFolderRequest struct {
	Path string `json:"path"`
}

type CreateRootFolderResponse struct {
	Folder  Folder  `json:"folder"`
	ScanJob ScanJob `json:"scanJob"`
}
