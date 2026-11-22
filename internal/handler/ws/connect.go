package ws

import (
	"file_share/internal/entity"
	"file_share/internal/infrastructure/ws"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) Connect(c *gin.Context) {
	ctx := c.Request.Context()

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		h.logger.Error(ctx, err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Автоматически читает Ping/Pong кадры и отменяет ctx, когда клиент отключился
	ctx = conn.CloseRead(ctx)

	deviceId := c.Query("deviceId")
	if deviceId == "" {
		deviceId = uuid.New().String()
	}

	device := entity.Device{
		Id:   deviceId,
		Name: c.DefaultQuery("deviceName", "Device"),
		Type: c.DefaultQuery("deviceType", entity.DeviceTypeWeb),
	}

	client := ws.Client{
		Conn:   conn,
		Device: device,
	}

	h.hub.Register(ctx, &client)
	defer h.hub.Unregister(ctx, &client)

	//for {
	//	var req entity.MessageRequest
	//
	//	if err := wsjson.Read(ctx, conn, &req); err != nil {
	//		h.logger.Error(ctx, err)
	//		return
	//	}
	//	switch req.Type {
	//	case ws.MsgTypeOpenOnTV:
	//		var msgOpenOnTV entity.CmdOpenOnTV
	//
	//		if err := json.Unmarshal(req.Payload, &msgOpenOnTV); err != nil {
	//			h.logger.Error(ctx, err)
	//			continue
	//		}
	//
	//		err = h.tvService.OpenToTV(ctx, msgOpenOnTV)
	//		if err != nil {
	//			h.logger.Error(ctx, err)
	//		}
	//	}
	//}
	h.hub.KeepAlive(ctx, conn)
}
