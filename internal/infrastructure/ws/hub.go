package ws

import (
	"context"
	"file_share/internal/entity"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Client struct {
	Device entity.Device
	Conn   *websocket.Conn
}

const (
	writeTimeout = 5 * time.Second
	pingInterval = 25 * time.Second
	pingTimeout  = 10 * time.Second
)

const (
	MsgTypeTVDevices           = "tv.devices"
	MsgTypeOpenOnTV            = "tv.open"
	MsgTypeTVNowPlaying        = "tv.playing"
	MsgTypeNotifyLibraryUpdate = "library.update"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		mu:      sync.RWMutex{},
	}
}

func (h *Hub) Register(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	old, exists := h.clients[client.Device.Id]

	h.clients[client.Device.Id] = client
	if exists {
		h.Disconnect(ctx, old)
	}
	if client.Device.Type == entity.DeviceTypeTV {
		go h.broadcastTVList()
	} else {
		go h.SentTVListTo(client)
	}

}

func (h *Hub) Unregister(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	current, exists := h.clients[client.Device.Id]
	if exists {
		if current == client {
			delete(h.clients, client.Device.Id)
		}
		if client.Device.Type == entity.DeviceTypeTV {
			go h.broadcastTVList()
		}
	}

}

func (h *Hub) Disconnect(_ context.Context, client *Client) {
	_, exists := h.clients[client.Device.Id]
	if exists {
		delete(h.clients, client.Device.Id)
		_ = client.Conn.Close(websocket.StatusNormalClosure, "client disconnected")
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for id, client := range h.clients {
		_ = client.Conn.Close(websocket.StatusGoingAway, "server shutting down")
		delete(h.clients, id)
	}
}

func (h *Hub) NotifyLibraryUpdated(ctx context.Context, entry entity.CmdUpdateLibrary) {
	h.broadcast(ctx, entity.Message{
		Type:    MsgTypeNotifyLibraryUpdate,
		Payload: entry,
	}, "")
}

func (h *Hub) KeepAlive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return // Клиент не ответил на Ping — выходим, сработает defer Unregister и conn.Close
			}
		}
	}
}

func (h *Hub) GetTVDevices(_ context.Context) []entity.Device {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices := make([]entity.Device, 0, len(h.clients))
	for _, client := range h.clients {
		if client.Device.Type == entity.DeviceTypeTV {
			devices = append(devices, client.Device)

		}
	}

	return devices
}

func (h *Hub) broadcast(ctx context.Context, msg entity.Message, senderDeviceId string) {
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		if senderDeviceId != client.Device.Id {
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	var wg sync.WaitGroup

	for _, item := range targets {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			defer cancel()
			_ = wsjson.Write(writeCtx, c.Conn, msg)
		}(item)
	}

	wg.Wait()
	return
}

func (h *Hub) broadcastTVList() {
	ctx := context.Background()

	h.broadcast(ctx, entity.Message{
		Type:    MsgTypeTVDevices,
		Payload: h.GetTVDevices(ctx),
	}, "")
}

func (h *Hub) SentTVListTo(c *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	_ = wsjson.Write(ctx, c.Conn, entity.Message{
		Type:    MsgTypeTVDevices,
		Payload: h.GetTVDevices(ctx),
	})
}

func (h *Hub) SendTVNowPlaying(ctx context.Context, cmd entity.CmdPlayNowOnTV) error {
	msg := entity.Message{
		Type:    MsgTypeTVNowPlaying,
		Payload: cmd,
	}

	h.broadcast(ctx, msg, cmd.DeviceId)

	return nil
}

func (h *Hub) SendOpenOnTV(ctx context.Context, cmd entity.CmdOpenOnTV) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[cmd.DeviceId]

	if ok {
		if client.Device.Type == entity.DeviceTypeTV {
			msg := entity.Message{
				Type:    MsgTypeOpenOnTV,
				Payload: cmd,
			}
			_ = wsjson.Write(ctx, client.Conn, msg)
		}
	}

	return nil
}
