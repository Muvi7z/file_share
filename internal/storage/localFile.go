package storage

import (
	"context"
	"file_share/internal/entity"
	"os"
)

func (s *Storage) Open(ctx context.Context, path string) (entity.VideoStream, error) {
	file, err := os.Open(path)
	if err != nil {
		return entity.VideoStream{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		if err = file.Close(); err != nil {
		}
		return entity.VideoStream{}, err
	}

	if stat.IsDir() {
		if err = file.Close(); err != nil {
		}
		return entity.VideoStream{}, os.ErrNotExist
	}

	return entity.VideoStream{
		FileName: stat.Name(),
		Size:     stat.Size(),
		ModTime:  stat.ModTime(),
		Reader:   file,
	}, nil
}

func (s *Storage) OpenPoster(ctx context.Context, path string) (entity.PosterFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return entity.PosterFile{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		if err = file.Close(); err != nil {

		}
		return entity.PosterFile{}, err
	}

	if stat.IsDir() {
		if err = file.Close(); err != nil {

		}
		return entity.PosterFile{}, os.ErrNotExist
	}

	return entity.PosterFile{
		FileName:    stat.Name(),
		ContentType: "image/jpeg",
		ModTime:     stat.ModTime(),
		Reader:      file,
	}, nil

}
