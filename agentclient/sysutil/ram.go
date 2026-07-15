//go:build windows

package sysutil

import "github.com/microsoft/UpdateAssistant/modules/loader"

func ram(fn func(*loader.Plugin)) {
	p, err := loader.Load("ram")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func RAMTotal() uint64 {
	var r uint64
	ram(func(p *loader.Plugin) { r = pU64(p, "ram_total") })
	return r
}
func RAMAvailable() uint64 {
	var r uint64
	ram(func(p *loader.Plugin) { r = pU64(p, "ram_available") })
	return r
}
func RAMUsagePct() int {
	var r int
	ram(func(p *loader.Plugin) { r = pNum(p, "ram_usage_pct") })
	return r
}
