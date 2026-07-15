package wsutil

import "encoding/json"

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

func encode(msg Message) ([]byte, bool) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, false
	}
	return data, true
}
