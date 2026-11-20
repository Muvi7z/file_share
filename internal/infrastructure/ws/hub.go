package ws

import (
	"context"
	"encoding/json"
	"file_share/internal/entity"
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{}
}

func (h *Hub) Register(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}

	if client.device.Type == entity.DeviceTypeTV {
		h.broadcastTVList()
	}
}

func (h *Hub) Unregister(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, exists := h.clients[client]
	if exists {
		delete(h.clients, client)
		close(client.send)
	}

}

func (h *Hub) GetTVDevices(ctx context.Context) ([]entity.Device, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return nil, nil
}

func (h *Hub) broadcastTVList() {
	devices := h.GetTVDevices(context.Background())
	h.broadcast("tv.devices", devices)
}

func (h *Hub) broadcast(msgType string, payload any) {
	data, err := marshalEnvelope(msgType, payload)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

func marshalEnvelope(msgType string, payload any) ([]byte, error) {
	return json.Marshal(map[string]interface{}{})
}
