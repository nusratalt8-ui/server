package transport

import "encoding/json"

type message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type startupPayload struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Hostname    string `json:"hostname"`
	Username    string `json:"username"`
}

type Handler func(msgType string, payload json.RawMessage)
