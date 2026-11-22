package tv

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/infrastructure/ws"
)

type Hub interface {
	SendOpenOnTV(ctx context.Context, cmd entity.CmdOpenOnTV) error
	SendPlayNowOnTV(ctx context.Context, cmd entity.CmdPlayNowOnTV) error
	Register(ctx context.Context, client *ws.Client)
	Unregister(ctx context.Context, client *ws.Client)
}

type videoService interface {
	GetVideoById(ctx context.Context, id string) (entity.Video, error)
}

type Service struct {
	logger       deps.Logger
	hub          Hub
	videoService videoService
}

func NewService(logger deps.Logger, hub Hub, videoService videoService) *Service {
	return &Service{
		logger:       logger,
		hub:          hub,
		videoService: videoService,
	}
}

func (s *Service) PlayNowOnTV(ctx context.Context, cmd entity.CmdPlayNowOnTV) error {
	if cmd.Entry.Id == "" {
		return entity.ClosePlayer
	}
	_, err := s.videoService.GetVideoById(ctx, cmd.Entry.Id)
	if err != nil {
		s.logger.Error(ctx, err)
		return err
	}

	err = s.hub.SendPlayNowOnTV(ctx, cmd)
	if err != nil {
		s.logger.Error(ctx, err)
		return err
	}

	return nil
}

func (s *Service) OpenOnTV(ctx context.Context, msg entity.CmdOpenOnTV) error {
	if msg.Mode == entity.ModeReplace || msg.Mode == entity.ModeAppend {
		if len(msg.Entries) < 0 {
			return errors.New("no entries")
		}
	}

	if msg.Mode == "" {
		msg.Mode = entity.ModeReplace
	}

	for _, entry := range msg.Entries {
		switch entry.Type {
		case entity.FileTypeVideo:
			_, err := s.videoService.GetVideoById(ctx, entry.Id)
			if err != nil {
				s.logger.Error(ctx, err)
				return err
			}

		}
	}

	err := s.hub.SendOpenOnTV(ctx, msg)
	if err != nil {
		s.logger.Error(ctx, err)
		return err
	}
	return nil
}
