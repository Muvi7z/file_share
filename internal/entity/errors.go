package entity

import "errors"

var (
	ErrorTokenNotFound = errors.New("token not found")
	ErrorForbidden     = errors.New("access denied")
	ErrorNoRowsFound   = errors.New("no rows found")
	ErrorBadRequest    = errors.New("incorrect fields")

	ErrorNotFoundFolder = errors.New("folder not found")
	ErrorGetFolders     = errors.New("error getting folders")
	ErrorGetFolder      = errors.New("error getting folder")
	ErrorUpdateFolder   = errors.New("error update folder")
	ErrorDeleteFolder   = errors.New("error delete folder")

	ErrorGetVideosEntries = errors.New("error getting video entries")

	ErrorCreateScanJob = errors.New("error create scan job")
	ErrorGetScanJob    = errors.New("error get scan job")
	ErrorNotFoundScan  = errors.New("scan job not found")

	ErrorNotFoundVideo  = errors.New("video not found")
	ErrorGetVideos      = errors.New("error get videos")
	ErrorCreateVideo    = errors.New("error create video")
	ErrorGetVideo       = errors.New("error get video")
	ErrorGetVideoStream = errors.New("error get video stream")
	ErrorUpdateVideo    = errors.New("error update video")
	ErrorCreatePoster   = errors.New("error create poster")
	ErrorDeletePoster   = errors.New("error delete poster")

	ErrorGetFile    = errors.New("error get file")
	ErrorDeleteFile = errors.New("error delete file")
	ErrorUpdateFile = errors.New("error update file")
	ErrorGetFiles   = errors.New("error get file")
	ErrorCreateFile = errors.New("error create file")

	ErrorGetUser            = errors.New("error get user")
	ErrorLoginUser          = errors.New("error login user")
	ErrorAuthMe             = errors.New("error auth me")
	ErrorInvalidCredentials = errors.New("invalid credentials")

	ErrSessionNotFound = errors.New("cache not found")
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
