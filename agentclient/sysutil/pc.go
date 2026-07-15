//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

func pcDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("pc")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func LockPC() bool {
	var ok bool
	pcDLL(func(p *loader.Plugin) {
		r := parseDLLResult(pStr(p, "lock_pc", 64))
		if !r.OK {
			logger.Error("lock failed: %s", r.Detail)
		}
		ok = r.OK
	})
	return ok
}

func ShutdownPC() { pcDLL(func(p *loader.Plugin) { p.Call("shutdown_pc") }) }
func RestartPC()  { pcDLL(func(p *loader.Plugin) { p.Call("restart_pc") }) }
