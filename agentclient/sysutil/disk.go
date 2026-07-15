//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func disk(fn func(*loader.Plugin)) {
	p, err := loader.Load("disk")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func DiskDrives() string {
	var r string
	disk(func(p *loader.Plugin) { r = pStr(p, "disk_drives", config.SmallBuf) })
	return r
}
func DiskTotal(drive string) uint64 {
	var r uint64
	disk(func(p *loader.Plugin) { r = pU64Drive(p, "disk_total", drive) })
	return r
}
func DiskFree(drive string) uint64 {
	var r uint64
	disk(func(p *loader.Plugin) { r = pU64Drive(p, "disk_free", drive) })
	return r
}
func DiskUsagePct(drive string) int {
	var r int
	disk(func(p *loader.Plugin) { r = pNumDrive(p, "disk_usage_pct", drive) })
	return r
}
