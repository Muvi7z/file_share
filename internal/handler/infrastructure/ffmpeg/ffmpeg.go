package ffmpeg

type FFMPEG struct {
	PathCmd string
}

func NewFFMPEG(pathCmd string) *FFMPEG {
	return &FFMPEG{
		PathCmd: pathCmd,
	}
}
