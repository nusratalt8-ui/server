//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func cpu(fn func(*loader.Plugin)) {
	p, err := loader.Load("cpu")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func CPUName() string {
	var r string
	cpu(func(p *loader.Plugin) { r = pStr(p, "cpu_name", config.SmallBuf) })
	return r
}
func CPUCores() int { var r int; cpu(func(p *loader.Plugin) { r = pNum(p, "cpu_cores") }); return r }
func CPUUsage() int { var r int; cpu(func(p *loader.Plugin) { r = pNum(p, "cpu_usage") }); return r }
