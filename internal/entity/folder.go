package entity

type ScanJon struct {
	Id               string `json:"id"`
	FolderId         string `json:"folderId"`
	Status           string `json:"status"`
	ProcessedVideos  int    `json:"processedVideos"`
	ProcessedFolders int    `json:"processedFolders"`
	StartedAt        string `json:"startedAt"`
	FinishedAt       string `json:"finishedAt"`
	Error            string `json:"error"`
}

type Folder struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	ParentId         string `json:"parentId"`
	RootFolderId     string `json:"rootFolderId"`
	IsRoot           bool   `json:"isRoot"`
	Enabled          bool   `json:"enabled"`
	FilesCount       int    `json:"filesCount"`
	VideosCount      int    `json:"videosCount"`
	ChildFolderCount int    `json:"childFolderCount"`
	LastScanAt       string `json:"lastScanAt"`
}

type FileBrowserEntry struct {
	Type   string `json:"type"`
	Folder Folder `json:"folder"`
	Video  Video  `json:"video"`
}
