package ffmpeg

import (
	"encoding/json"
	"file_share/internal/entity"
	"fmt"
	"os/exec"
)

func (f *FFMPEG) GetAudioTrack(inputPath string) ([]entity.AudioTrackStream, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries",
		"stream=index,codec_type,codec_name,channels:stream_tags=language,title:stream_disposition=default",
		"-of", "json",
		inputPath,
	)

	cmd.Dir = f.PathCmd

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg GetAudioTrack failed: %w: %s", err, string(out))
	}

	var audioTrackInfo entity.AudioTrackInfo

	err = json.Unmarshal(out, &audioTrackInfo)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg GetAudioTrack failed: %w", err)
	}

	return audioTrackInfo.AudioTrackStream, nil
}
