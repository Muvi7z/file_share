package ws

import (
	"context"
	"file_share/internal/entity"
	"time"

	"github.com/coder/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	pingInterval   = 25 * time.Second
	pingTimeout    = 15 * time.Second
	maxMessageSize = 64 * 1024 // 64 KB
)

type IncomingHandler func(ctx context.Context, client *Client, env entity.Message)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	device    entity.Device
	onMessage IncomingHandler
}
