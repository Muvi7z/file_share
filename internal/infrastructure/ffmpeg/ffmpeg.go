package ffmpeg

type FFMPEG struct {
	PathCmd string
	LogPath string
}

func NewFFMPEG(pathCmd string) *FFMPEG {
	return &FFMPEG{
		PathCmd: pathCmd,
		LogPath: pathCmd,
	}
}
