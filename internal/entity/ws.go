package entity

import "encoding/json"

const (
	DeviceTypeTV     = "tv"
	DeviceTypeWeb    = "web"
	DeviceTypeMobile = "mobile"
)

const (
	ModeReplace = "replace"
	ModeAppend  = "append"
	ModeSync    = "sync"
	ModeStop    = "stop"
)

const (
	CmdFileDelete = "delete"
	CmdFileAdd    = "add"
)

type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type MessageRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Entry struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type CmdOpenOnTV struct {
	DeviceId string  `json:"deviceId"`
	Mode     string  `json:"mode"`
	Entries  []Entry `json:"entries"`
}

type CmdPlayNowOnTV struct {
	DeviceId string `json:"deviceId"`
	Entry    Entry  `json:"entry"`
}

type CmdUpdateLibrary struct {
	File FileBrowserEntry `json:"file"`
	Type string           `json:"type"`
}

type Device struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
