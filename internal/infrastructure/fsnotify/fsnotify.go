package fsnotify

import (
	"context"
	fs "github.com/fsnotify/fsnotify"
)

type FSNotify struct {
	watcher *fs.Watcher
}

func New(watcher *fs.Watcher) *FSNotify {
	return &FSNotify{watcher: watcher}
}

func (f *FSNotify) Add(_ context.Context, path string) error {
	err := f.watcher.Add(path)
	if err != nil {
		return err
	}

	return nil
}

func (f *FSNotify) Remove(_ context.Context, path string) error {
	err := f.watcher.Remove(path)
	if err != nil {
		return err
	}

	return nil
}

func (f *FSNotify) GetEventsChannel() <-chan fs.Event {
	return f.watcher.Events
}
