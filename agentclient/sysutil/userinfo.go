//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func userinfo(fn func(*loader.Plugin)) {
	p, err := loader.Load("userinfo")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func UserName() string {
	var r string
	userinfo(func(p *loader.Plugin) { r = pStr(p, "user_name", config.SmallBuf) })
	return r
}
func UserDomain() string {
	var r string
	userinfo(func(p *loader.Plugin) { r = pStr(p, "user_domain", config.SmallBuf) })
	return r
}
func UserIsAdmin() bool {
	var r int
	userinfo(func(p *loader.Plugin) { r = pNum(p, "user_is_admin") })
	return r == 1
}
