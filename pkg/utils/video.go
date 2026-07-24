package utils

import "fmt"

func FormatDuration(totalSeconds int64) string {
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	// %d — часы без ведущего нуля (2, а не 02)
	// %02d — минуты и секунды с ведущим нулём (05, а не 5)
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

func GetHalfTimeVideo(duration int64) string {
	return FormatDuration(duration / 3)
}
