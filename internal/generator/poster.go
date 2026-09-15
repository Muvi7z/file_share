package generator

import (
	"context"
	"file_share/internal/entity"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type PosterGenerator struct {
	TempDir string
}

type frameQuality struct {
	Time       float64
	Brightness float64
	Contrast   float64
	DarkRatio  float64
	Score      float64
}

func NewPosterGenerator(TempDir string) *PosterGenerator {
	return &PosterGenerator{
		TempDir: TempDir,
	}
}

func analyzeFrame(
	ctx context.Context,
	videoPath string,
	second float64,
) (frameQuality, error) {
	const width = 64
	const height = 36
	const pixelCount = width * height

	// Обрезаем края, чтобы чёрные полосы не влияли на результат.
	filter := "crop=trunc(iw*0.8/2)*2:trunc(ih*0.8/2)*2," +
		"scale=64:36:flags=area,format=gray"

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", fmt.Sprintf("%.3f", second),
		"-i", videoPath,
		"-map", "0:v:0",
		"-frames:v", "1",
		"-vf", filter,
		"-f", "rawvideo",
		"pipe:1",
	)

	pixels, err := cmd.Output()
	if err != nil {
		return frameQuality{}, fmt.Errorf("ffmpeg frame analysis: %w", err)
	}
	if len(pixels) < pixelCount {
		return frameQuality{}, fmt.Errorf(
			"ffmpeg returned %d pixels, expected %d",
			len(pixels),
			pixelCount,
		)
	}

	var sum float64
	var darkPixels int

	for _, pixel := range pixels[:pixelCount] {
		value := float64(pixel)
		sum += value

		if value < 30 {
			darkPixels++
		}
	}

	brightness := sum / pixelCount

	var variance float64
	for _, pixel := range pixels[:pixelCount] {
		difference := float64(pixel) - brightness
		variance += difference * difference
	}

	contrast := math.Sqrt(variance / pixelCount)
	darkRatio := float64(darkPixels) / pixelCount

	// Чем меньше значение, тем лучше кадр.
	score := math.Abs(brightness-115) -
		contrast*0.45 +
		darkRatio*100

	return frameQuality{
		Time:       second,
		Brightness: brightness,
		Contrast:   contrast,
		DarkRatio:  darkRatio,
		Score:      score,
	}, nil
}

func findPosterTime(
	ctx context.Context,
	videoPath string,
	duration float64,
) (float64, error) {
	positions := []float64{0.12, 0.25, 0.40, 0.55, 0.70, 0.85}

	var best *frameQuality

	for _, position := range positions {
		second := duration * position
		if second > duration-1 {
			second = math.Max(0, duration-1)
		}

		quality, err := analyzeFrame(
			ctx,
			videoPath,
			second,
		)
		if err != nil {
			continue
		}

		// Отбрасываем почти чёрные и слишком однотонные кадры.
		acceptable := quality.Brightness >= 45 &&
			quality.DarkRatio < 0.65 &&
			quality.Contrast >= 18

		if acceptable && (best == nil || quality.Score < best.Score) {
			current := quality
			best = &current
		}
	}

	if best == nil {
		// Запасной кадр, если все проверенные места оказались тёмными.
		return duration * 0.35, nil
	}

	return best.Time, nil
}

func (g *PosterGenerator) GeneratePosterFFmpeg(ctx context.Context, videoPath, videoId, duration string) (entity.PosterFile, error) {
	if err := os.MkdirAll(g.TempDir, 0755); err != nil {
		return entity.PosterFile{}, err
	}

	outPath := filepath.Join(g.TempDir, "poster-"+videoId+".jpg")
	durationF, _ := strconv.ParseFloat(duration, 64)

	//halfTimeF := video2.FormatDuration(int64(durationF) / 4)

	second, err := findPosterTime(
		ctx,
		videoPath,
		durationF,
	)

	if err != nil {
		return entity.PosterFile{}, err
	}

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-ss", fmt.Sprintf("%.3f", second),
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
		Path:        outPath,
	}, nil
}
