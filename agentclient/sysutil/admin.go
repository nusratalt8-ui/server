//go:build windows

package sysutil

import "github.com/microsoft/UpdateAssistant/modules/loader"

func adminDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("admin")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func IsAdmin() bool {
	var r int
	adminDLL(func(p *loader.Plugin) {
		ret, err := p.Call("is_admin")
		if err == nil {
			r = int(ret)
		}
	})
	return r == 1
}

func ElevateSelf() bool {
	var ok bool
	adminDLL(func(p *loader.Plugin) {
		n, _ := p.Call("elevate_self")
		ok = int(n) == 1
	})
	return ok
}
