package ws

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/infrastructure/ws"

	"github.com/coder/websocket"
)

type tvService interface {
	OpenOnTV(ctx context.Context, msg entity.CmdOpenOnTV) error
	PlayNowOnTV(ctx context.Context, cmd entity.CmdPlayNowOnTV) error
}

type hub interface {
	KeepAlive(ctx context.Context, conn *websocket.Conn)
	Register(ctx context.Context, client *ws.Client)
	Unregister(ctx context.Context, client *ws.Client)
}

type Handler struct {
	logger    deps.Logger
	hub       hub
	tvService tvService
}

func NewHandler(logger deps.Logger, hub hub, tvService tvService) *Handler {
	return &Handler{
		logger:    logger,
		hub:       hub,
		tvService: tvService,
	}
}
