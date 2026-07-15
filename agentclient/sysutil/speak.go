//go:build windows

package sysutil

import (
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/loader"
)

func speakDLL(fn func(*loader.Plugin)) {
	p, err := loader.Load("speak")
	if err != nil {
		return
	}
	defer loader.Unload(p)
	fn(p)
}

func Speak(text string) bool {
	var ok bool
	speakDLL(func(p *loader.Plugin) {
		n, err := p.Call("speak", uintptr(unsafe.Pointer(loader.BytePtr(text))), uintptr(len(text)))
		ok = err == nil && int(n) == 1
	})
	return ok
}

func SpeakStop() {
	speakDLL(func(p *loader.Plugin) {
		p.Call("speak_stop")
	})
}
