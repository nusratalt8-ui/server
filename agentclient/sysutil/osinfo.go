//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func osinfo(fn func(*loader.Plugin)) {
	p, err := loader.Load("osinfo")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func OSName() string {
	var r string
	osinfo(func(p *loader.Plugin) { r = pStr(p, "os_name", config.SmallBuf) })
	return r
}
func OSBuild() string {
	var r string
	osinfo(func(p *loader.Plugin) { r = pStr(p, "os_build", config.SmallBuf) })
	return r
}
func OSArch() string {
	var r string
	osinfo(func(p *loader.Plugin) { r = pStr(p, "os_arch", config.SmallBuf) })
	return r
}
