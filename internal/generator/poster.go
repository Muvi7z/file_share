package generator

import (
	"context"
	"file_share/internal/entity"
	"os"
	"os/exec"
	"path/filepath"
)

type PosterGenerator struct {
	TempDir string
}

func NewPosterGenerator(TempDir string) *PosterGenerator {
	return &PosterGenerator{
		TempDir: TempDir,
	}
}

func (g *PosterGenerator) GeneratePosterFFmpeg(ctx context.Context, videoPath, videoId, duration string) (entity.PosterFile, error) {
	if err := os.MkdirAll(g.TempDir, 0755); err != nil {
		return entity.PosterFile{}, err
	}

	outPath := filepath.Join(g.TempDir, "poster-"+videoId+".jpg")

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", duration,
		"-i", videoPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outPath,
	)

	if err := cmd.Run(); err != nil {
		return entity.PosterFile{}, err
	}

	file, err := os.Open(outPath)
	if err != nil {
		return entity.PosterFile{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return entity.PosterFile{}, err
	}

	return entity.PosterFile{
		FileName:    filepath.Base(outPath),
		ContentType: "image/jpeg",
		ModTime:     stat.ModTime(),
		Reader:      file,
	}, nil
}
