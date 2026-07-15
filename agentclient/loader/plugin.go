//go:build windows

package loader

/*
#include "memload.h"
extern void install_crash_handler(void);
*/
import "C"
import (
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

type Plugin struct {
	Name   string
	handle unsafe.Pointer
	refs   atomic.Int64
	mu     sync.Mutex
}

func Open(name string, data []byte) (*Plugin, error) {
	if len(data) == 0 {
		return nil, ErrLoadFailed
	}
	h := C.mod_load(unsafe.Pointer(&data[0]), C.size_t(len(data)))
	if h == nil {
		return nil, ErrLoadFailed
	}
	return &Plugin{Name: name, handle: unsafe.Pointer(h)}, nil
}

func (p *Plugin) Call(export string, args ...uintptr) (uintptr, error) {
	p.mu.Lock()
	if p.handle == nil {
		p.mu.Unlock()
		return 0, ErrProcMissing
	}
	n := C.CString(export)
	proc := C.mod_sym(p.handle, n)
	C.free(unsafe.Pointer(n))
	if proc == nil {
		p.mu.Unlock()
		return 0, ErrProcMissing
	}
	p.refs.Add(1)
	p.mu.Unlock()

	ret, _, _ := syscall.SyscallN(uintptr(unsafe.Pointer(proc)), args...)

	p.refs.Add(-1)
	return ret, nil
}

func (p *Plugin) Close() {
	p.mu.Lock()
	h := p.handle
	p.handle = nil
	p.mu.Unlock()

	if h == nil {
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	for p.refs.Load() > 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	C.mod_free(h)
}

func init() {
	C.install_crash_handler()
}

func BytePtr(s string) *byte {
	p := C.CString(s)
	return (*byte)(unsafe.Pointer(p))
}
