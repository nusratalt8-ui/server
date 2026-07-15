//go:build windows

package sysutil

import (
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

var (
	volMu   sync.Mutex
	volPlug *loader.Plugin
)

func volLoad() {
	volMu.Lock()
	defer volMu.Unlock()
	if volPlug != nil {
		return
	}
	p, err := loader.Load("vol")
	if err != nil {
		return
	}
	volPlug = p
}

func VolUnload() {
	volMu.Lock()
	defer volMu.Unlock()
	if volPlug != nil {
		loader.Unload(volPlug)
		volPlug = nil
	}
}

func VolGet() int {
	volLoad()
	volMu.Lock()
	p := volPlug
	volMu.Unlock()
	if p == nil {
		return -1
	}
	r, _ := p.Call("vol_get")
	v := int(int32(r))
	if v < 0 || v > 100 {
		return -1
	}
	return v
}

func VolSet(pct int) bool {
	volLoad()
	volMu.Lock()
	p := volPlug
	volMu.Unlock()
	if p == nil {
		return false
	}
	r, _ := p.Call("vol_set", uintptr(pct))
	return int(int32(r)) == 1
}
