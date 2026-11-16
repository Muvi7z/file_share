package scan

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	video2 "file_share/pkg/utils/video"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/vansante/go-ffprobe"
)

type repository interface {
	GetAllScanJob(ctx context.Context, status string) ([]entity.ScanJob, error)
	GetFolderById(ctx context.Context, id string) (entity.Folder, error)
	CreateFolder(ctx context.Context, folder entity.Folder) (entity.Folder, error)
	CreateVideo(ctx context.Context, video entity.Video) (entity.Video, error)
	GetAllVideo(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64, enabled *bool) ([]entity.Video, error)
	DeleteVideo(ctx context.Context, id string) error
	UpdateFolder(ctx context.Context, folder entity.UpdateFolderRequest, id string) (entity.Folder, error)
	GetFolders(ctx context.Context, query, rootFolderId, parentFolderId string, isRoot, enabled *bool) ([]entity.Folder, error)
	DeleteFolder(ctx context.Context, id string) error
	GetAllFile(ctx context.Context, query, rootFolderId, parentFolderId string, limit uint64, offset uint64) ([]entity.File, error)
	CreateFile(ctx context.Context, file entity.File) (entity.File, error)
	DeleteFile(ctx context.Context, id string) error
	CreateAudioTracks(ctx context.Context, audioTracks []entity.AudioTrack) ([]entity.AudioTrack, error)
}

type scanJobRepository interface {
	CreateScanJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
	GetScanJob(ctx context.Context, id string) (entity.ScanJob, error)
	UpdateJob(ctx context.Context, job entity.ScanJob) (entity.ScanJob, error)
}

type playbackService interface {
	PrepareHls(ctx context.Context, video entity.Video) ([]entity.AudioTrack, func() error, error)
	DeleteHLS(_ context.Context, id string) error
}

type posterGenerator interface {
	GeneratePosterFFMPEG(ctx context.Context, videoPath, videoId, duration string) (entity.PosterFile, error)
}

type fsWatcher interface {
	Add(ctx context.Context, path string) error
	Remove(ctx context.Context, path string) error
	GetEventsChannel() <-chan fsnotify.Event
}

type watchFile struct {
	Status string
	Name   string
	Path   string
	Size   int64
}

const (
	StatusWatchWriting   = "writing"
	StatusWatchCompleted = "completed"
)

var allowedVideoExtension = map[string]bool{
	".mp4":     true,
	".avi":     true,
	entity.MKV: true,
	".mov":     true,
	".webm":    true,
	".flv":     true,
	".wmv":     true,
	".m4v":     true,
	".mpeg":    true,
	".mpg":     true,
	".3gp":     true,
}

type Scan struct {
	logger            deps.Logger
	repository        repository
	scanJobRepository scanJobRepository
	posterGenerator   posterGenerator
	playbackService   playbackService
	fsWatcher         fsWatcher
}

func New(
	logger deps.Logger,
	repository repository,
	scanJobRepository scanJobRepository,
	posterGenerator posterGenerator,
	playbackService playbackService,
	fsWatcher fsWatcher,
) *Scan {
	return &Scan{
		logger:            logger,
		repository:        repository,
		scanJobRepository: scanJobRepository,
		posterGenerator:   posterGenerator,
		playbackService:   playbackService,
		fsWatcher:         fsWatcher,
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

func (s *Scan) StartAutoScan(ctx context.Context) error {
	workers := 4
	// добавить просмотр этих папок

	//

	go func() {
		mapFiles := make(map[string]watchFile)
		mapFolders := make(map[string]entity.Folder)
		hlsChan := make(chan entity.Video, 4)

		defer close(hlsChan)

		go func() {
			channels := s.fanOutHLS(ctx, hlsChan, workers)

			result := s.fanInHLS(ctx, channels...)

			for err := range result {
				if err != nil {
					s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
					return
				}
			}
		}()

		isRoot := true
		rootFolders, err := s.repository.GetFolders(ctx, "", "", "", &isRoot, nil)
		if err != nil {
			s.logger.Info(ctx, "error to get root folders")
			return
		}

		for _, rootFolder := range rootFolders {
			folders, err := s.repository.GetFolders(ctx, "", rootFolder.Id, "", nil, nil)
			if err != nil {
				s.logger.Info(ctx, "error to get root folders")
				continue
			}

			for _, folder := range folders {
				err := s.fsWatcher.Add(ctx, folder.Path)
				if err != nil {
					s.logger.Info(ctx, "error to watch folders")
					continue
				}

				mapFolders[folder.Path] = folder

			}

			err = s.fsWatcher.Add(ctx, rootFolder.Path)
			if err != nil {
				s.logger.Info(ctx, "error to watch folders")
				continue
			}

			mapFolders[rootFolder.Path] = rootFolder
		}

		for {
			select {
			case <-ctx.Done():
				s.logger.Info(ctx, "Stop AutoScan process")
				return
			case event, ok := <-s.fsWatcher.GetEventsChannel():

				if !ok {
					s.logger.Info(ctx, "Stop AutoScan process")
					return
				}

				switch event.Op {
				case fsnotify.Write:
					fileWatch, okFile := mapFiles[event.Name]
					if !okFile {
						s.logger.Error(ctx, errors.New("error to read map files"), event.Name)
						continue
					}
					time.Sleep(time.Second * 1)
					if fileWatch.Status == StatusWatchWriting {
						file, err := os.Open(event.Name)
						if err != nil {
							s.logger.Info(ctx, "error to open file FileWatchWriting", event.Name, event)
							continue
						}
						s.logger.Info(ctx, "open file FileWatchWriting", event.Name, event)

						info, err := file.Stat()
						if err != nil {
							s.logger.Info(ctx, "error to stat file", event.Name)
							continue
						}

						folderPath := filepath.Dir(event.Name)
						fileName := filepath.Base(event.Name) // "my_video.mp4"
						ext := filepath.Ext(fileName)
						parentFolder := mapFolders[folderPath]

						if allowedVideoExtension[ext] {
							entityVideo, err := s.ScanVideo(ctx, info, event.Name)
							if err != nil {
								s.logger.Error(ctx, fmt.Errorf("failed scan video: %v", err))
								continue
							}
							entityVideo.Video.ParentFolderId = parentFolder.Id
							entityVideo.Video.FolderName = parentFolder.Name
							entityVideo.Video.FolderId = parentFolder.RootFolderId

							fileWatch.Status = StatusWatchCompleted

							mapFiles[event.Name] = fileWatch
							if strings.EqualFold(ext, entity.MKV) {
								hlsChan <- *entityVideo.Video

							} else {
								err = s.CreateVideo(ctx, entityVideo.Video)
								if err != nil {
									s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
									continue
								}
							}
						} else {
							entityFile, err := s.ScanFile(ctx, info, "", event.Name)
							if err != nil {
								s.logger.Error(ctx, fmt.Errorf("failed scan video: %v", err))
								continue
							}
							entityFile.File.ParentFolderId = parentFolder.Id
							entityFile.File.FolderName = parentFolder.Name
							entityFile.File.FolderId = parentFolder.RootFolderId

							_, err = s.repository.CreateFile(ctx, *entityFile.File)
							if err != nil {
								s.logger.Error(ctx, fmt.Errorf("failed create file: %v", err))
								continue
							}
						}

						err = file.Close()
						if err != nil {
							return
						}
					}

				case fsnotify.Create:
					fileWatch, okFile := mapFiles[event.Name]

					_ = fileWatch
					if !okFile {

						file, err := os.Open(event.Name)
						if err != nil {
							mapFiles[event.Name] = watchFile{
								Status: StatusWatchWriting,
								Path:   event.Name,
							}

							s.logger.Info(ctx, "error to open file FileWatchCreate", event.Name)
							continue
						}
						s.logger.Info(ctx, "event", event)
						err = file.Close()
						if err != nil {
							s.logger.Error(ctx, errors.New("error to close file"))

							return
						}
					}

				default:
					s.logger.Info(ctx, "event unhandled default case")
				}
				//если файл большой, то при копировании он помечается create
				// нужно отслеживать его путь и смотреть после получения инвента write, размер файла, если он изменился значит файл все еще копируется

			}
		}
	}()

	return nil
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
	ctxScan, cancel := context.WithCancel(ctx)
	workers := 4

	defer cancel()

	folders, err := s.repository.GetFolders(ctxScan, "", rootFolder.Id, "", nil, nil)
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
			s.logger.Error(ctxScan, fmt.Errorf("failed walk dir: %v", err))
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
			parentPath := filepath.Dir(path)

			folderEntry, ok := browserFileMap[parentPath]
			if ok {
				folderEntry.Folder.VideosCount++

				browserFileMap[parentPath] = folderEntry
			}

			if allowedVideoExtension[ext] {
				info, _ := d.Info()

				entityVideo, err := s.ScanVideo(ctxScan, info, path)
				if err != nil {
					s.logger.Error(ctxScan, fmt.Errorf("failed scan video: %v", err))
					return nil
				}

				entityVideo.Video.ParentFolderId = folderEntry.Folder.Id
				entityVideo.Video.FolderName = folderEntry.Folder.Name
				entityVideo.Video.FolderId = rootFolder.Id

				localVideos = append(localVideos, *entityVideo.Video)
				browserFileMap[path] = entityVideo

			} else {
				info, err := d.Info()
				if err != nil {
					s.logger.Error(ctx, fmt.Errorf("failed get info: %v", err))
					return nil
				}
				entityFile, err := s.ScanFile(ctxScan, info, "", path)
				if err != nil {
					s.logger.Error(ctxScan, fmt.Errorf("failed scan video: %v", err))
					return nil
				}

				entityFile.File.FolderId = rootFolder.Id
				entityFile.File.ParentFolderId = folderEntry.Folder.Id

				localFiles = append(localFiles, *entityFile.File)
				browserFileMap[path] = entityFile
			}

			return nil
		}

		return nil // Продолжаем обход
	})
	if err != nil {
		return nil, err
	}

	for _, folderItem := range folders {
		_, ok := browserFileMap[folderItem.Path]
		if !ok {
			err = s.repository.DeleteFolder(ctxScan, folderItem.Id)
			if err != nil {
				s.logger.Error(ctxScan, fmt.Errorf("failed delete folder: %v", err))
				return nil, err

			}
			continue
		}
	}

	for _, localFolder := range localFolders {
		folderEntry, ok := foldersEntries[localFolder.Path]
		if !ok {
			_, err = s.repository.CreateFolder(ctxScan, localFolder)
			if err != nil {
				s.logger.Error(ctxScan, fmt.Errorf("failed create video: %v", err))
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
		_, err = s.repository.UpdateFolder(ctxScan, update, folderEntry.Folder.Id)
		if err != nil {
			if errors.Is(err, entity.ErrorNoRowsFound) {
				return nil, nil
			}
			s.logger.Error(ctxScan, fmt.Errorf("failed update video: %v", err))
			return nil, err

		}

	}

	videos, err := s.repository.GetAllVideo(ctxScan, "", rootFolder.Id, "", 0, 0, nil)
	if err != nil {
		return nil, err
	}

	//удаление видео если видео не найдено в проводнике
	for _, item := range videos {
		_, ok := browserFileMap[item.Path]
		if !ok {

			err = s.repository.DeleteVideo(ctxScan, item.Id)
			if err != nil {
				s.logger.Error(ctxScan, fmt.Errorf("failed delete folder: %v", err))
				return nil, err

			}

			if item.Extension == entity.MKV {
				err = s.playbackService.DeleteHLS(ctxScan, item.Id)
				if err != nil {
					s.logger.Error(ctxScan, fmt.Errorf("failed delete video: %v", err))
					return nil, err
				}
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

	hlsVideos := make([]entity.Video, 0)

	//создание видео если видео не найдено в базе
	for _, localVideo := range localVideos {
		_, ok := videosEntries[localVideo.Path]
		if !ok {

			if strings.EqualFold(localVideo.Extension, entity.MKV) {

				hlsVideos = append(hlsVideos, localVideo)

			} else {
				err = s.CreateVideo(ctxScan, &localVideo)
				if err != nil {
					s.logger.Error(ctxScan, fmt.Errorf("failed create video: %v", err))
					continue
				}
			}
		}
	}

	//err = s.CreateHLSVideo(ctx, hlsChan)
	//if err != nil {
	//	return nil, err
	//}

	hlsChan, err := s.AddHlsJob(ctxScan, hlsVideos)
	if err != nil {
		s.logger.Error(ctxScan, fmt.Errorf("failed create video: %v", err))
	}
	channels := s.fanOutHLS(ctxScan, hlsChan, workers)

	result := s.fanInHLS(ctxScan, channels...)

	for err = range result {
		if err != nil {
			s.logger.Error(ctx, fmt.Errorf("failed create video: %v", err))
			return nil, err
		}
	}

	files, err := s.repository.GetAllFile(ctxScan, "", rootFolder.Id, "", 0, 0)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		_, ok := browserFileMap[file.Path]
		if !ok {
			err = s.repository.DeleteFile(ctxScan, file.Id)
			if err != nil {
				s.logger.Error(ctxScan, fmt.Errorf("failed delete file: %v", err))
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
			_, err = s.repository.CreateFile(ctxScan, localFile)
			if err != nil {
				s.logger.Error(ctxScan, fmt.Errorf("failed create file: %v", err))
				return nil, err
			}
			continue
		}
	}

	return browserFileMap, nil
}

func (s *Scan) CreateVideo(ctx context.Context, video *entity.Video) error {
	file, err := s.posterGenerator.GeneratePosterFFMPEG(ctx, video.Path, video.Id, video.Duration)
	if err != nil {
		return fmt.Errorf("failed create poster: %v", err)
	}
	err = file.Reader.Close()
	if err != nil {
		return err
	}

	video.PosterUrl = file.Path

	_, err = s.repository.CreateVideo(ctx, *video)
	if err != nil {
		return fmt.Errorf("failed create video: %v", err)
	}
	return err
}

func (s *Scan) ScanFile(_ context.Context, info fs.FileInfo, mime string, path string) (entity.FileBrowserEntry, error) {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName) // ".mp4"
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	file := entity.File{
		Id:         uuid.New().String(),
		Name:       nameWithoutExt,
		Path:       path,
		Extension:  ext,
		Size:       video2.FormatFileSize(info.Size()),
		SizeBytes:  info.Size(),
		ModifiedAt: info.ModTime(),
		MimeType:   mime,
	}

	return entity.FileBrowserEntry{
		Type: entity.FileTypeOther,
		File: &file,
	}, nil

}

func (s *Scan) fanInHLS(ctx context.Context, channels ...chan error) chan error {
	finalStream := make(chan error)

	var wg sync.WaitGroup

	for _, channel := range channels {
		wg.Add(1)

		chCopy := channel

		go func() {
			defer wg.Done()
			for err := range chCopy {
				select {
				case <-ctx.Done():
					return
				case finalStream <- err:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalStream)
	}()

	return finalStream
}

func (s *Scan) fanOutHLS(ctx context.Context, inputChan chan entity.Video, workers int) []chan error {
	resultChannels := make([]chan error, workers)
	for i := 0; i < workers; i++ {
		resultChannels[i] = s.StartHls(ctx, inputChan)
	}

	return resultChannels
}

func (s *Scan) StartHls(ctx context.Context, hlsChan chan entity.Video) chan error {
	resultChan := make(chan error)
	go func() {
		defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				return
			case localVideo, ok := <-hlsChan:
				if !ok {
					s.logger.Info(ctx, "video hls close: %v", localVideo.Title)

					return
				}
				if err := ctx.Err(); err != nil {
					s.logger.Info(ctx, "video hls cancel: %v", localVideo.Title)
					return
				}
				s.logger.Info(ctx, "video hls started: %v", localVideo.Title)
				audioTracks, cleanup, err := s.playbackService.PrepareHls(ctx, localVideo)
				if err != nil {
					//s.logger.Error(ctx, fmt.Errorf("failed create hls: %v", err))
					resultChan <- err
					continue
				}

				s.logger.Info(ctx, "video hls created: %v", localVideo.Title)
				err = s.CreateVideo(ctx, &localVideo)
				if err != nil {
					resultChan <- err
					err = cleanup()
					if err != nil {
						resultChan <- err
					}
					continue
				}

				_, err = s.repository.CreateAudioTracks(ctx, audioTracks)
				if err != nil {
					resultChan <- err
					err = cleanup()
					if err != nil {
						resultChan <- err
					}
					continue
				}
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
		}
	}()

	return resultChan
}

func (s *Scan) AddHlsJob(ctx context.Context, videos []entity.Video) (chan entity.Video, error) {
	hlsChan := make(chan entity.Video, 4)
	go func() {
		defer close(hlsChan)
		for _, localVideo := range videos {
			select {
			case <-ctx.Done():
				return
			case hlsChan <- localVideo:
			}

		}
	}()

	return hlsChan, nil
}

func (s *Scan) ScanVideo(ctx context.Context, info os.FileInfo, path string) (entity.FileBrowserEntry, error) {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName) // ".mp4"
	if !allowedVideoExtension[ext] {
		return entity.FileBrowserEntry{}, errors.New("error extension not allowed")
	}

	data, err := ffprobe.GetProbeDataContext(ctx, path)
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed walk dir: %s %v", path, err))
		return entity.FileBrowserEntry{}, err
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
		Extension:  ext,
	}

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
