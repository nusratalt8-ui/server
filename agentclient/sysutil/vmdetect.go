//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

func vmdetect(fn func(*loader.Plugin)) {
	p, err := loader.Load("vmdetect")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func IsVM() bool {
	var result bool
	vmdetect(func(p *loader.Plugin) {
		flags, _ := p.Call("sandbox_flags")
		f := int(flags)

		type check struct {
			bit   int
			name  string
			score int
		}
		checks := []check{
			{1, "registry", 2},
			{2, "mac", 2},
			{4, "modules", 2},
			{8, "vmprocs", 2},
			{16, "bios", 2},
			{32, "files", 1},
			{64, "sleep", 2},
			{128, "username", 1},
			{256, "hostname", 1},
			{512, "ram", 2},
			{1024, "cpu", 1},
			{2048, "processes", 1},
			{4096, "timing", 1},
		}

		score := 0
		for _, c := range checks {
			if f&c.bit != 0 {
				logger.Info("vmdetect: %-12s FAIL (+%d)", c.name, c.score)
				score += c.score
			} else {
				logger.Info("vmdetect: %-12s ok", c.name)
			}
		}
		logger.Info("vmdetect: score=%d blocked=%v", score, score >= 3)
		result = score >= 3
	})
	return result
}
