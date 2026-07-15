//go:build windows

package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

func Inject(pid uint32, name string) error {
	defMu.RLock()
	l := def
	defMu.RUnlock()
	if l == nil {
		return ErrLoadFailed
	}
	data, err := loadBytes(l.src, name)
	if err != nil {
		return fmt.Errorf("load payload %s: %w", name, err)
	}
	tmpDir := filepath.Join(os.TempDir(), "inject")
	os.MkdirAll(tmpDir, 0755)
	dst := filepath.Join(tmpDir, name+".dll")
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	return injectRemote(pid, dst)
}

func InjectFile(pid uint32, dllPath string) error {
	return injectRemote(pid, dllPath)
}

func FindProcessPID(name string) (uint32, error) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(snap)
	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := syscall.Process32First(snap, &pe); err != nil {
		return 0, err
	}
	for {
		if matchProcessName(pe, name) {
			return pe.ProcessID, nil
		}
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return 0, fmt.Errorf("process %q not found", name)
}

func matchProcessName(pe syscall.ProcessEntry32, name string) bool {
	exe := syscall.UTF16ToString(pe.ExeFile[:])
	if len(exe) != len(name) {
		return false
	}
	for i := 0; i < len(exe); i++ {
		if lower(exe[i]) != lower(name[i]) {
			return false
		}
	}
	return true
}

func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

func injectRemote(pid uint32, dllPath string) error {
	const (
		procCreateThread     = 0x0002
		procQueryInformation = 0x0400
		procVMOperation      = 0x0008
		procVMWrite          = 0x0020
		procVMRead           = 0x0010
		memCommit            = 0x1000
		memReserve           = 0x2000
		pageReadWrite        = 0x04
	)
	hProc, err := syscall.OpenProcess(
		procCreateThread|procQueryInformation|
			procVMOperation|procVMWrite|procVMRead,
		false, pid)
	if err != nil {
		return fmt.Errorf("openprocess: %v", err)
	}
	defer syscall.CloseHandle(hProc)
	pathBytes := append([]byte(dllPath), 0)
	pathLen := uintptr(len(pathBytes))
	remoteAddr, err := virtualAllocEx(hProc, 0, pathLen, memCommit|memReserve, pageReadWrite)
	if err != nil {
		return fmt.Errorf("virtualalloc: %v", err)
	}
	var written uintptr
	if err := writeProcessMemory(hProc, remoteAddr, &pathBytes[0], pathLen, &written); err != nil {
		return fmt.Errorf("writememory: %v", err)
	}
	loadLib, err := getProcAddress("kernel32.dll", "LoadLibraryA")
	if err != nil {
		return fmt.Errorf("getprocaddr: %v", err)
	}
	threadH, _, err := createRemoteThread(hProc, 0, 0, loadLib, remoteAddr, 0, nil)
	if err != nil {
		return fmt.Errorf("remotethread: %v", err)
	}
	syscall.WaitForSingleObject(syscall.Handle(threadH), syscall.INFINITE)
	syscall.CloseHandle(syscall.Handle(threadH))
	return nil
}

func virtualAllocEx(h syscall.Handle, addr, size, allocType, protect uintptr) (uintptr, error) {
	p := syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualAllocEx")
	r, _, err := p.Call(uintptr(h), addr, size, allocType, protect)
	if r == 0 {
		return 0, err
	}
	return r, nil
}

func writeProcessMemory(h syscall.Handle, addr uintptr, buf *byte, size uintptr, written *uintptr) error {
	p := syscall.NewLazyDLL("kernel32.dll").NewProc("WriteProcessMemory")
	r, _, err := p.Call(uintptr(h), addr, uintptr(unsafe.Pointer(buf)), size, uintptr(unsafe.Pointer(written)))
	if r == 0 {
		return err
	}
	return nil
}

func getProcAddress(dll, proc string) (uintptr, error) {
	h, err := syscall.LoadLibrary(dll)
	if err != nil {
		return 0, err
	}
	defer syscall.FreeLibrary(h)
	addr, err := syscall.GetProcAddress(h, proc)
	if err != nil {
		return 0, err
	}
	return uintptr(addr), nil
}

func createRemoteThread(h syscall.Handle, sa, stackSize, startAddr, param, flags uintptr, tid *uint32) (uintptr, uintptr, error) {
	p := syscall.NewLazyDLL("kernel32.dll").NewProc("CreateRemoteThread")
	r, _, err := p.Call(uintptr(h), sa, stackSize, startAddr, param, flags, uintptr(unsafe.Pointer(tid)))
	if r == 0 {
		return 0, 0, err
	}
	return r, 0, nil
}
