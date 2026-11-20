package entity

import "encoding/json"

const (
	DeviceTypeTV      = "tv"
	DeviceTypeDesktop = "desktop"
	DeviceTypeMobile  = "mobile"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Device struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
