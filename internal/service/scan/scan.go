package scan

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vansante/go-ffprobe"
)

type repository interface {
	GetAllScanJob(ctx context.Context, status string) ([]entity.ScanJob, error)
	GetFolderById(ctx context.Context, id string) (entity.Folder, error)
	CreateFolder(ctx context.Context, folder entity.Folder) (entity.Folder, error)
	CreateVideo(ctx context.Context, video entity.Video) (entity.Video, error)
	UpdateFolder(ctx context.Context, folder entity.UpdateFolderRequest, id string) (entity.Folder, error)
}

type scanJobRepository interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
	GetScanJob(ctx context.Context, id string) (entity.ScanJob, error)
	UpdateJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type posterGenerator interface {
	GeneratePosterFFmpeg(ctx context.Context, videoPath, videoId string) (entity.PosterFile, error)
}

var allowedVideoExts = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mkv":  true,
	".mov":  true,
	".webm": true,
	".flv":  true,
	".wmv":  true,
	".m4v":  true,
	".mpeg": true,
	".mpg":  true,
	".3gp":  true,
	".m3u8": true,
}

type Scan struct {
	logger            deps.Logger
	repository        repository
	scanJobRepository scanJobRepository
	posterGenerator   posterGenerator
}

func New(logger deps.Logger, repository repository, scanJobRepository scanJobRepository, posterGenerator posterGenerator) *Scan {
	return &Scan{
		logger:            logger,
		repository:        repository,
		scanJobRepository: scanJobRepository,
		posterGenerator:   posterGenerator,
	}
}

func (s *Scan) GetScanJob(ctx context.Context, id string) (entity.ScanJob, error) {
	scanJob, err := s.scanJobRepository.GetScanJob(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundScan) {
			return entity.ScanJob{}, entity.ErrorNotFoundScan
		}
		return entity.ScanJob{}, entity.ErrorGetScanJob
	}

	return scanJob, nil
}

func (s *Scan) CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error) {
	scan, err := s.scanJobRepository.CreateScanJob(ctx, job)
	if err != nil {
		return entity.ScanJob{}, entity.ErrorCreateScanJob
	}

	return scan, nil
}

func (s *Scan) StartProcessScan(ctx context.Context, handlePeriod time.Duration) {
	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Info(ctx, "Stop scan process")
				return
			case <-ticker.C:
			}

			scanJobs, err := s.repository.GetAllScanJob(ctx, entity.StatusQueued)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed get jobs: %v", err))
				continue
			}

			for _, scanJob := range scanJobs {
				var folder entity.Folder

				folder, err = s.repository.GetFolderById(ctx, scanJob.FolderId)

				err = s.ScanFolder(ctx, folder)
				if err != nil {
					continue
				}

				scanJob.Status = entity.StatusCompleted
				scanJob.FinishedAt = time.Now()
				_, err = s.scanJobRepository.UpdateJob(ctx, scanJob)
				if err != nil {
					continue
				}

				//создание папок или файлов

			}

			//обновить данные корневой папки

			//удалить файлы или папки, которых больше нет
		}
	}()
}

func (s *Scan) ScanFolder(ctx context.Context, rootFolder entity.Folder) error {
	folderMap := make(map[string]entity.Folder)

	err := filepath.WalkDir(rootFolder.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Error(ctx, fmt.Errorf("failed walk dir: %v", err))
			return nil // Возвращаем nil, чтобы продолжить сканирование остальных файлов
		}

		// 2. Проверяем, является ли элемент директорией
		if d.IsDir() {
			if rootFolder.Path == path {
				folderMap[path] = rootFolder

				return nil
			}
			uuidFolder := uuid.New().String()

			folder := entity.Folder{
				Id:               uuidFolder,
				Name:             d.Name(),
				Path:             path,
				ParentId:         "",
				RootFolderId:     "",
				IsRoot:           false,
				Enabled:          false,
				FilesCount:       0,
				VideosCount:      0,
				ChildFolderCount: 0,
				LastScanAt:       time.Time{},
			}

			parentPath := filepath.Dir(path)

			parentFolder, ok := folderMap[parentPath]
			if ok {
				parentFolder.FilesCount++
				parentFolder.ChildFolderCount++

				folderMap[parentPath] = parentFolder

				folder.ParentId = parentFolder.Id
				folder.RootFolderId = rootFolder.Id
			}
			_, err = s.repository.CreateFolder(ctx, folder)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
				return err
			}

			folderMap[path] = folder

		} else {
			// Получаем размер файла (требует дополнительного sys call, поэтому вызываем только для файлов)
			info, _ := d.Info()
			data, err := ffprobe.GetProbeDataContext(ctx, path)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed walk dir: %v", err))
				return nil
			}

			fileName := filepath.Base(path) // "my_video.mp4"

			// 2. Получаем расширение
			ext := filepath.Ext(fileName) // ".mp4"

			if !allowedVideoExts[ext] {
				return nil
			}

			// 3. Обрезаем расширение
			nameWithoutExt := strings.TrimSuffix(fileName, ext)

			uuidVideo := uuid.New().String()

			video := entity.Video{
				Id:         uuidVideo,
				Title:      nameWithoutExt,
				PosterUrl:  fmt.Sprintf("/api/videos/%s/poster", uuidVideo),
				StreamUrl:  fmt.Sprintf("/api/videos/%s/stream", uuidVideo),
				SizeBytes:  info.Size(),
				Path:       path,
				ModifiedAt: info.ModTime(),
				Size:       data.Format.Size,
			}

			parentPath := filepath.Dir(path)

			folder, ok := folderMap[parentPath]
			if ok {
				folder.VideosCount++

				folderMap[parentPath] = folder

				video.ParentFolderId = folder.Id
				video.FolderName = folder.Name
				video.FolderId = folder.Id
			}

			for _, stream := range data.Streams {
				switch stream.CodecType {
				case "video":
					video.Codec = stream.CodecName
					video.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)
					video.Size = FormatFileSize(info.Size())
					duration, err := strconv.Atoi(stream.Duration)
					if err == nil {
						video.Duration = FormatDuration(int64(duration))
					} else {
						video.Duration = stream.Duration
					}
				case "audio":
				case "subtitle":
				}
			}

			file, err := s.posterGenerator.GeneratePosterFFmpeg(ctx, path, uuidVideo)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create poster: %v", err))
				return err
			}
			err = file.Reader.Close()
			if err != nil {
				return err
			}
			_, err = s.repository.CreateVideo(ctx, video)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
				return err
			}

		}

		return nil // Продолжаем обход
	})

	folder, ok := folderMap[rootFolder.Path]
	if ok {
		update := entity.UpdateFolderRequest{
			FilesCount:       folder.FilesCount,
			VideosCount:      folder.VideosCount,
			ChildFolderCount: folder.ChildFolderCount,
		}
		_, err = s.repository.UpdateFolder(ctx, update, folder.Id)
		if err != nil {
			return err
		}

	}

	if err != nil {
		return err
	}

	return nil
}

func FormatDuration(totalSeconds int64) string {
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	// %d — часы без ведущего нуля (2, а не 02)
	// %02d — минуты и секунды с ведущим нулём (05, а не 5)
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
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
