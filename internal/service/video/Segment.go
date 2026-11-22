package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (s *Service) Segment(inputPath, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	manifestPath := filepath.Join(outputDir, "index.m3u8")
	segmentPattern := filepath.Join(outputDir, "segment%03d.ts")

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-codec:", "copy", // Copy streams without re-encoding (fast, no quality loss)
		"-start_number", "0",
		"-hls_time", "6", // Target segment duration in seconds
		"-hls_list_size", "0", // 0 = keep all segments in manifest (VOD)
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		manifestPath,
	)

	cmd.Stderr = os.Stderr

	if s.logger != nil {
		s.logger.Info(context.Background(), "segmenting %s -> %s", inputPath, outputDir)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}

	if s.logger != nil {
		s.logger.Info(context.Background(), "segmentation complete: %s", manifestPath)
	}
	return nil
}
