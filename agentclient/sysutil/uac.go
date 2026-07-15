//go:build windows

package sysutil

import (
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	uacMu   sync.Mutex
	uacPlug *loader.Plugin
)

func uacDLL(fn func(*loader.Plugin)) {
	uacMu.Lock()
	if uacPlug == nil {
		p, err := loader.Load("uac")
		if err == nil {
			uacPlug = p
		}
	}
	p := uacPlug
	uacMu.Unlock()
	if p != nil {
		fn(p)
	}
}

func UACUnload() {
	uacMu.Lock()
	defer uacMu.Unlock()
	if uacPlug != nil {
		loader.Unload(uacPlug)
		uacPlug = nil
	}
}

func UACStatus() string {
	raw := dllStrCall(uacDLL, "uac_status")
	r := parseDLLResult(raw)
	if r.Detail != "" {
		logger.Error("uac status: %s (%s)", r.Status, r.Detail)
	}
	return r.Status
}

func UACEnable() bool {
	raw := dllStrCall(uacDLL, "uac_enable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("uac enable failed: %s", r.Detail)
	}
	return r.OK
}

func UACDisable() bool {
	raw := dllStrCall(uacDLL, "uac_disable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("uac disable failed: %s", r.Detail)
	}
	return r.OK
}

func UACToggle() (string, bool) {
	return dllToggle(uacDLL, "uac_status", "uac_enable", "uac_disable")
}
