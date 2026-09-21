package scan

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	video2 "file_share/pkg/utils/video"
	"fmt"
	"io/fs"
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
	GetAllVideo(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.Video, error)
	DeleteVideo(ctx context.Context, id string) error
	UpdateFolder(ctx context.Context, folder entity.UpdateFolderRequest, id string) (entity.Folder, error)
	GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot, enabled *bool) ([]entity.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
	GetAllFile(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.File, error)
	CreateFile(ctx context.Context, file entity.File) (entity.File, error)
	DeleteFile(ctx context.Context, id string) error
}

type scanJobRepository interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
	GetScanJob(ctx context.Context, id string) (entity.ScanJob, error)
	UpdateJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type posterGenerator interface {
	GeneratePosterFFmpeg(ctx context.Context, videoPath, videoId, duration string) (entity.PosterFile, error)
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
				if err != nil {
					continue
				}

				_, err := s.ScanFolder(ctx, folder)
				if err != nil {
					continue
				}

				scanJob.Status = entity.StatusCompleted
				scanJob.FinishedAt = time.Now()
				_, err = s.scanJobRepository.UpdateJob(ctx, scanJob)
				if err != nil {
					continue
				}

			}

		}
	}()
}

func (s *Scan) ScanFolder(ctx context.Context, rootFolder entity.Folder) (map[string]entity.FileBrowserEntry, error) {
	browserFileMap := make(map[string]entity.FileBrowserEntry)
	var localFolders []entity.Folder
	var localVideos []entity.Video
	var localFiles []entity.File
	foldersEntries := make(map[string]entity.FileBrowserEntry)
	videosEntries := make(map[string]entity.FileBrowserEntry)
	fileEntries := make(map[string]entity.FileBrowserEntry)
	const numWorkers = 8

	jobs := make(chan string, 200)

	// Запускаем фиксированный пул воркеров.
	for w := 1; w <= numWorkers; w++ {
		go s.WorkerFixFastStart(ctx, jobs)
	}

	folders, err := s.repository.GetFolders(ctx, "", rootFolder.Id, "", nil, nil)
	if err != nil {
		return nil, err
	}

	for _, folderItem := range folders {

		foldersEntries[folderItem.Path] = entity.FileBrowserEntry{
			Type:   entity.FileTypeFolder,
			Folder: &folderItem,
			Video:  nil,
		}
	}

	err = filepath.WalkDir(rootFolder.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Error(ctx, fmt.Errorf("failed walk dir: %v", err))
			return nil // Возвращаем nil, чтобы продолжить сканирование остальных файлов
		}

		// 2. Проверяем, является ли элемент директорией
		if d.IsDir() {
			if rootFolder.Path == path {
				browserFileMap[path] = entity.FileBrowserEntry{
					Type:   entity.FileTypeFolder,
					Folder: &rootFolder,
				}
				localFolders = append(localFolders, rootFolder)
				return nil
			}
			uuidFolder := uuid.New().String()

			folderEntries, ok := foldersEntries[path]
			if ok {
				uuidFolder = folderEntries.Folder.Id
			}
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

			parentFolder, ok := browserFileMap[parentPath]
			if ok {
				parentFolder.Folder.FilesCount++
				parentFolder.Folder.ChildFolderCount++

				browserFileMap[parentPath] = parentFolder

				folder.ParentId = parentFolder.Folder.Id
				folder.RootFolderId = rootFolder.Id
			}

			localFolders = append(localFolders, folder)
			browserFileMap[path] = entity.FileBrowserEntry{
				Type:   entity.FileTypeFolder,
				Folder: &folder,
			}

		} else {
			// Получаем размер файла (требует дополнительного sys call, поэтому вызываем только для файлов)
			fileName := filepath.Base(path) // "my_video.mp4"

			// 2. Получаем расширение
			ext := filepath.Ext(fileName) // ".mp4"
			ok := allowedVideoExts[ext]
			if ok {
				parentPath := filepath.Dir(path)

				folderEntry, ok := browserFileMap[parentPath]
				if ok {
					folderEntry.Folder.VideosCount++

					browserFileMap[parentPath] = folderEntry
				}

				entityVideo, err := s.ScanVideo(ctx, d, rootFolder.Id, folderEntry, path)
				if err != nil {
					s.logger.Error(ctx, fmt.Errorf("failed scan video: %v", err))
					return nil
				}

				localVideos = append(localVideos, *entityVideo.Video)
				browserFileMap[path] = entityVideo

				jobs <- path
				return nil
			}
			//mime, ok := entity.AllowedImageExts[ext]
			parentPath := filepath.Dir(path)

			folderEntry, ok := browserFileMap[parentPath]
			if ok {
				folderEntry.Folder.FilesCount++

				browserFileMap[parentPath] = folderEntry
			}

			entityFile, err := s.ScanFile(ctx, d, rootFolder.Id, folderEntry, "", path)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed scan video: %v", err))
				return nil
			}

			localFiles = append(localFiles, *entityFile.File)
			browserFileMap[path] = entityFile
			return nil

		}

		return nil // Продолжаем обход
	})
	close(jobs)
	if err != nil {
		return nil, err
	}

	for _, folderItem := range folders {
		_, ok := browserFileMap[folderItem.Path]
		if !ok {
			err = s.repository.DeleteFolder(ctx, folderItem.Id)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed delete folder: %v", err))
				return nil, err

			}
			continue
		}
	}

	for _, localFolder := range localFolders {
		folderEntry, ok := foldersEntries[localFolder.Path]
		if !ok {
			_, err = s.repository.CreateFolder(ctx, localFolder)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
				return nil, err

			}
			continue
		}

		browserFile, ok := browserFileMap[localFolder.Path]
		update := entity.UpdateFolderRequest{
			FilesCount:       &browserFile.Folder.FilesCount,
			VideosCount:      &browserFile.Folder.VideosCount,
			ChildFolderCount: &browserFile.Folder.ChildFolderCount,
		}
		_, err = s.repository.UpdateFolder(ctx, update, folderEntry.Folder.Id)
		if err != nil {
			if errors.Is(err, entity.ErrorNoRowsFound) {
				return nil, nil
			}
			s.logger.Error(ctx, fmt.Errorf("failed update video: %v", err))
			return nil, err

		}

	}

	videos, err := s.repository.GetAllVideo(ctx, "", rootFolder.Id, "", 0, 0)
	if err != nil {
		return nil, err

	}

	//удаление видео если видео не найдено в проводнике
	for _, item := range videos {
		_, ok := browserFileMap[item.Path]
		if !ok {

			err = s.repository.DeleteVideo(ctx, item.Id)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed delete folder: %v", err))
				return nil, err

			}
			continue
		}

		videosEntries[item.Path] = entity.FileBrowserEntry{
			Type:   entity.FileTypeVideo,
			Folder: nil,
			File:   nil,
			Video:  &item,
		}
	}

	//создание видео если видео не найдено в базе
	for _, localVideo := range localVideos {
		_, ok := videosEntries[localVideo.Path]
		if !ok {

			file, err := s.posterGenerator.GeneratePosterFFmpeg(ctx, localVideo.Path, localVideo.Id, localVideo.Duration)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create poster: %v", err))
				continue
			}
			err = file.Reader.Close()
			if err != nil {
				return nil, err
			}

			localVideo.PosterUrl = file.Path

			_, err = s.repository.CreateVideo(ctx, localVideo)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
				return nil, err
			}
			continue
		}
	}

	files, err := s.repository.GetAllFile(ctx, "", rootFolder.Id, "", 0, 0)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		_, ok := browserFileMap[file.Path]
		if !ok {
			err = s.repository.DeleteFile(ctx, file.Id)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed delete file: %v", err))
				return nil, err

			}
			continue
		}

		fileEntries[file.Path] = entity.FileBrowserEntry{
			Type: entity.FileTypeOther,
			File: &file,
		}
	}

	for _, localFile := range localFiles {
		_, ok := fileEntries[localFile.Path]
		if !ok {
			_, err = s.repository.CreateFile(ctx, localFile)
			if err != nil {
				s.logger.Error(ctx, fmt.Errorf("failed create file: %v", err))
				return nil, err
			}
			continue
		}
	}

	return browserFileMap, nil
}

func (s *Scan) ScanFile(ctx context.Context, fileEntry fs.DirEntry, rootFolderId string, folderEntry entity.FileBrowserEntry, mime string, path string) (entity.FileBrowserEntry, error) {
	info, _ := fileEntry.Info()
	// 3. Обрезаем расширение
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName) // ".mp4"
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	file := entity.File{
		Id:             uuid.New().String(),
		Name:           nameWithoutExt,
		Path:           path,
		Extension:      ext,
		FolderId:       rootFolderId,
		ParentFolderId: folderEntry.Folder.Id,
		Size:           video2.FormatFileSize(info.Size()),
		SizeBytes:      info.Size(),
		ModifiedAt:     info.ModTime(),
		MimeType:       mime,
	}

	return entity.FileBrowserEntry{
		Type: entity.FileTypeOther,
		File: &file,
	}, nil

}

func (s *Scan) ScanVideo(ctx context.Context, videoFile fs.DirEntry, rootFolderId string, folderEntry entity.FileBrowserEntry, path string) (entity.FileBrowserEntry, error) {
	info, _ := videoFile.Info()
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName) // ".mp4"
	if !allowedVideoExts[ext] {
		return entity.FileBrowserEntry{}, nil
	}

	data, err := ffprobe.GetProbeDataContext(ctx, path)
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed walk dir: %s %v", path, err))
		return entity.FileBrowserEntry{}, nil
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

	video.ParentFolderId = folderEntry.Folder.Id
	video.FolderName = folderEntry.Folder.Name
	video.FolderId = rootFolderId

	for _, stream := range data.Streams {
		switch stream.CodecType {
		case "video":
			video.Codec = stream.CodecName
			video.Resolution = fmt.Sprintf("%dx%d", stream.Width, stream.Height)
			video.Size = video2.FormatFileSize(info.Size())
			duration, err := strconv.Atoi(stream.Duration)
			if err == nil {
				video.Duration = video2.FormatDuration(int64(duration))
			} else {
				video.Duration = stream.Duration
			}
		case "audio":
		case "subtitle":
		}
	}

	return entity.FileBrowserEntry{
		Type:  entity.FileTypeVideo,
		Video: &video,
	}, nil
}

func (s *Scan) WorkerFixFastStart(ctx context.Context, jobs <-chan string) {
	for job := range jobs {
		s.logger.Info(ctx, fmt.Sprintf("%v: started fix", job))
		err := video2.FixFastStart(ctx, job)
		if err != nil {
			s.logger.Error(ctx, fmt.Errorf("%v: failed fix fast start: %v", job, err))
		}
		s.logger.Info(ctx, fmt.Sprintf("%v: end fix", job))
	}

}
