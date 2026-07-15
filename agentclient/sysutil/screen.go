//go:build windows

package sysutil

import "github.com/microsoft/UpdateAssistant/modules/loader"

func screenDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("screen")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func ScreenWidth() int {
	var r int
	screenDLL(func(p *loader.Plugin) { r = pNum(p, "screen_width") })
	return r
}
func ScreenHeight() int {
	var r int
	screenDLL(func(p *loader.Plugin) { r = pNum(p, "screen_height") })
	return r
}
