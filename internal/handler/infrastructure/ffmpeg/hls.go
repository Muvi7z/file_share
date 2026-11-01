package ffmpeg

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func CreateHLS() {
	args := []string{
		"-i", inputPath,
		"-map", "0:v:0",
	}

	variants := []string{
		"v:0,agroup:audio,name:video",
	}

	for i, track := range tracks {
		args = append(args,
			"-map", fmt.Sprintf("0:%d", track.StreamIndex),
		)

		defaultValue := "no"
		if i == 0 {
			defaultValue = "yes"
		}

		// Предполагается, что ID — UUID, сформированный сервером.
		variants = append(variants, fmt.Sprintf(
			"a:%d,agroup:audio,default:%s,name:%s",
			i,
			defaultValue,
			track.ID,
		))
	}

	args = append(args,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-ac", "2",
		"-f", "hls",
		"-hls_segment_type", "mpegts",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-var_stream_map", strings.Join(variants, " "),
		"-master_pl_name", "master.m3u8",
		"-hls_segment_filename",
		filepath.Join(outputDir, "%v_%06d.ts"),
		filepath.Join(outputDir, "%v.m3u8"),
	)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
}
