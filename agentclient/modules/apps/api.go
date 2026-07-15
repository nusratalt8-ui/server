//go:build windows

package apps

import (
	"encoding/json"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type Tab struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type tabDef struct {
	Tab
	appID   string
	handler func(tabID, msgType string, payload json.RawMessage)
}

type Tabs struct {
	mu     sync.Mutex
	defs   []*tabDef
	byKey  map[string]*tabDef
	appIDs []string
}

func NewTabs() *Tabs {
	return &Tabs{byKey: make(map[string]*tabDef)}
}

func (t *Tabs) DeclareApp(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, existing := range t.appIDs {
		if existing == id {
			return
		}
	}
	t.appIDs = append(t.appIDs, id)
}

func (t *Tabs) AddTab(appID, id, label string, handler func(tabID, msgType string, payload json.RawMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	d := &tabDef{Tab: Tab{ID: id, Label: label}, appID: appID, handler: handler}
	t.defs = append(t.defs, d)
	t.byKey[appID+":"+id] = d
}

func (t *Tabs) SendApps(client *transport.Client) {
	t.mu.Lock()
	ids := make([]string, len(t.appIDs))
	copy(ids, t.appIDs)
	t.mu.Unlock()
	client.Send("panel_apps", map[string]interface{}{"apps": ids})
}

func (t *Tabs) SendTabs(client *transport.Client, appID string) {
	t.mu.Lock()
	out := make([]Tab, 0)
	for _, d := range t.defs {
		if d.appID == appID {
			out = append(out, d.Tab)
		}
	}
	t.mu.Unlock()
	if len(out) == 0 {
		return
	}
	client.Send("panel_tabs", map[string]interface{}{
		"app_id": appID,
		"tabs":   out,
	})
}

func (t *Tabs) HandleTabsGet(client *transport.Client, payload json.RawMessage) {
	var p struct {
		AppID string `json:"app_id"`
	}
	if json.Unmarshal(payload, &p) == nil && p.AppID != "" {
		t.SendTabs(client, p.AppID)
	}
}

func (t *Tabs) HandleTabAction(payload json.RawMessage) {
	var p struct {
		AppID   string          `json:"app_id"`
		TabID   string          `json:"tab_id"`
		MsgType string          `json:"msg_type"`
		Payload json.RawMessage `json:"payload"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AppID == "" || p.TabID == "" {
		return
	}
	t.mu.Lock()
	d := t.byKey[p.AppID+":"+p.TabID]
	t.mu.Unlock()
	if d != nil {
		go d.handler(p.TabID, p.MsgType, p.Payload)
	}
}
