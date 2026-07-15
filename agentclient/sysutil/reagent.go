//go:build windows

package sysutil

import (
	"strings"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func reagentc(fn func(*loader.Plugin)) {
	p, err := loader.Load("reagentc")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func ReagentcStatus() string {
	var r string
	reagentc(func(p *loader.Plugin) { r = pStr(p, "reagentc_status", config.SmallBuf) })
	return r
}

func ReagentcEnabled() bool {
	s := ReagentcStatus()
	return strings.Contains(strings.ToLower(s), "enabled")
}

func ReagentcEnable() string {
	var r string
	reagentc(func(p *loader.Plugin) { r = pStr(p, "reagentc_enable", config.SmallBuf) })
	return r
}

func ReagentcDisable() string {
	var r string
	reagentc(func(p *loader.Plugin) { r = pStr(p, "reagentc_disable", config.SmallBuf) })
	return r
}

func ReagentcToggle() (enabled bool, output string, ok bool) {
	if ReagentcEnabled() {
		output = ReagentcDisable()
		enabled = false
	} else {
		output = ReagentcEnable()
		enabled = true
	}
	lower := strings.ToLower(output)
	ok = !strings.Contains(lower, "error") && !strings.Contains(lower, "fail") && output != ""
	return
}
