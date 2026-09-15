package storage

import (
	"context"
	"file_share/internal/entity"
	"io"
	"os"
	"path/filepath"
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

func (s *Storage) CreatePoster(ctx context.Context, image io.ReadSeeker, path string) error {
	if err := os.MkdirAll(filepath.Dir(s.PosterPath), 0777); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, image)
	if err != nil {
		return err
	}

	return nil

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
