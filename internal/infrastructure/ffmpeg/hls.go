package ffmpeg

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (f *FFMPEG) CreateHLS(ctx context.Context, video entity.Video, audioTracks []entity.AudioTrack) error {
	args := []string{
		"-i", video.Path,
		"-map", "0:v:0",
	}

	variants := []string{
		"v:0,agroup:audio,name:video",
	}

	outputDir := filepath.Join(f.PathCmd, video.Id)

	err := os.Mkdir(outputDir, 0777)
	if err != nil {
		return err
	}

	for i, track := range audioTracks {
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
			track.Id,
		))

		audioDir := filepath.Join(outputDir, track.Id)

		err = os.Mkdir(audioDir, 0777)
		if err != nil {
			return err
		}
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
		filepath.Join(outputDir, "%v", "segment_%06d.ts"),
		filepath.Join(outputDir, "%v", "index.m3u8"),
	)

	logfilePath := filepath.Join(outputDir, fmt.Sprintf("%v.log", video.Id))

	logFile, err := os.Create(logfilePath)
	if err != nil {
		return fmt.Errorf("создать журнал FFmpeg: %w", err)
	}
	defer logFile.Close()

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Dir = f.PathCmd

	if err = cmd.Run(); err != nil {
		return err
	}

	return nil
}
