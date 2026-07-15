package controlpanel

import (
	"encoding/json"
	"sync"

	"agentmanager/modules/panelapp"
)

const maxTabs = 8

type Service struct {
	app       *panelapp.App
	appsCache sync.Map
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("panel_apps_get", app.Guarded(s.onPanelAppsGet))
	app.Panel.RegisterHandler("agent_delete", app.Guarded(s.onAgentDelete))
	app.Panel.RegisterHandler("panel_open", app.Guarded(s.onPanelOpen))
	app.Panel.RegisterHandler("panel_close", app.Guarded(s.onPanelClose))
	app.Panel.RegisterHandler("panel_get", app.Guarded(s.onPanelGet))
	app.Panel.RegisterHandler("panel_action", app.Guarded(s.onPanelAction))
	app.Panel.RegisterHandler("panel_tabs_get", app.Guarded(s.onPanelTabsGet))
	app.Panel.RegisterHandler("panel_tab_action", app.Guarded(s.onPanelTabAction))
	app.Agent.RegisterHandler("panel_buttons", s.onPanelButtons)
	app.Agent.RegisterHandler("panel_result", s.onPanelResult)
	app.Agent.RegisterHandler("panel_apps", s.onPanelApps)
	app.Agent.RegisterHandler("panel_tabs", s.onPanelTabs)
	return s
}

func (s *Service) onAgentDelete(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
    uid := s.app.Panel.TagOf(panelConnID)
    s.app.Reg.DeleteAndBroadcast(p.AgentID, uid)
}

func (s *Service) onPanelOpen(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "controlpanel:"+p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_open", nil)
	}
}

func (s *Service) onPanelClose(panelConnID string, payload json.RawMessage) {
	s.app.AppClose(panelConnID, payload, "controlpanel", "panel_close")
}

func (s *Service) onPanelGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "controlpanel:"+p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_get", nil)
	}
}

func (s *Service) onPanelAction(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		ID      string `json:"id"`
		Level   int    `json:"level"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" || p.ID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_action", p)
	}
}

func (s *Service) onPanelApps(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p struct {
		Apps []string `json:"apps"`
	}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	s.appsCache.Store(id, p.Apps)
	msg := map[string]interface{}{"agent_id": id, "apps": p.Apps}
	s.app.Publish("apps:"+id, "panel_apps", msg)
	s.app.Panel.EmitToTag(s.app.Agent.TagOf(agentConnID), "panel_apps", msg)
}

func (s *Service) onPanelAppsGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "apps:"+p.AgentID)
	if cached, ok := s.appsCache.Load(p.AgentID); ok {
		s.app.Panel.EmitTo(panelConnID, "panel_apps", map[string]interface{}{
			"agent_id": p.AgentID,
			"apps":     cached,
		})
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_apps_get", nil)
	}
}

func (s *Service) onPanelTabs(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p struct {
		AppID string `json:"app_id"`
		Tabs  []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"tabs"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AppID == "" {
		return
	}
	if len(p.Tabs) > maxTabs {
		p.Tabs = p.Tabs[:maxTabs]
	}
	s.app.Publish("tabs:"+id+":"+p.AppID, "panel_tabs", map[string]interface{}{
		"agent_id": id,
		"app_id":   p.AppID,
		"tabs":     p.Tabs,
	})
}

func (s *Service) onPanelTabsGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		AppID   string `json:"app_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" || p.AppID == "" {
		return
	}
	topic := "tabs:" + p.AgentID + ":" + p.AppID
	s.app.Sess.Subscribe(panelConnID, topic)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_tabs_get", map[string]string{"app_id": p.AppID})
	}
}

func (s *Service) onPanelTabAction(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		AppID   string `json:"app_id"`
		TabID   string `json:"tab_id"`
		MsgType string `json:"msg_type"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" || p.AppID == "" || p.TabID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "panel_tab_action", map[string]string{
			"app_id":   p.AppID,
			"tab_id":   p.TabID,
			"msg_type": p.MsgType,
		})
	}
}

func (s *Service) onPanelButtons(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var status interface{}
	if json.Unmarshal(payload, &status) != nil {
		return
	}
	s.app.Publish("controlpanel:"+id, "panel_buttons", map[string]interface{}{"agent_id": id, "result": status})
}

func (s *Service) onPanelResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var result interface{}
	if json.Unmarshal(payload, &result) != nil {
		return
	}
	s.app.Publish("controlpanel:"+id, "panel_result", map[string]interface{}{"agent_id": id, "result": result})
}
