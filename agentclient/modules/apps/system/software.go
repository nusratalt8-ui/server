//go:build windows

package system

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
)

func sendSoftware(s *session) {
	buf := make([]byte, config.CmdBuf*16)
	n, err := s.d.software.Call("software_list", uintptr(unsafe.Pointer(&buf[0])), uintptr(config.CmdBuf*16))
	data := ""
	if err == nil {
		data = string(buf[:int(n)])
	}
	s.client.Send("system_software", map[string]string{"json": data})
}