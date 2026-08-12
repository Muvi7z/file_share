package video

import (
	"fmt"
	"math"
	"mime"
	"path/filepath"
	"strings"
)

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

func VideoContentType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".mp4":
		return "video/mp4"
	case ".m4v":
		return "video/x-m4v"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	default:
		if contentType := mime.TypeByExtension(filepath.Ext(fileName)); contentType != "" {
			return contentType
		}
		return "application/octet-stream"
	}
}

func FormatFileSize(bytes int64) string {
	if bytes < 0 {
		return "-" + FormatFileSize(-bytes)
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d Б", bytes)
	}

	// Единицы измерения в порядке возрастания
	units := []string{"КБ", "МБ", "ГБ", "ТБ", "ПБ", "ЭБ"}

	// Находим подходящую единицу
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}

	// Считаем значение с плавающей точкой
	value := float64(bytes) / float64(div)

	// Округляем до 2 знаков после запятой
	value = math.Round(value*100) / 100

	// Убираем лишние нули: 1.50 МБ -> 1.5 МБ, 2.00 МБ -> 2 МБ
	formatted := fmt.Sprintf("%.2f", value)
	// Отрезаем нули в конце
	for formatted[len(formatted)-1] == '0' {
		formatted = formatted[:len(formatted)-1]
	}
	// Если осталась точка в конце — убираем и её
	if formatted[len(formatted)-1] == '.' {
		formatted = formatted[:len(formatted)-1]
	}

	return fmt.Sprintf("%s %s", formatted, units[exp])
}
