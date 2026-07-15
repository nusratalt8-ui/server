package system

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("system_open", app.Guarded(s.onSystemOpen))
	app.Panel.RegisterHandler("system_close", app.Guarded(s.onSystemClose))
	app.Panel.RegisterHandler("system_refresh", app.Guarded(s.onSystemRefresh))
	app.Panel.RegisterHandler("system_clipboard_get", app.Guarded(s.onSystemClipGet))
	app.Panel.RegisterHandler("system_clipboard_set", app.Guarded(s.onSystemClipSet))
	app.Panel.RegisterHandler("system_startup_list", app.Guarded(s.onSystemStartupList))
	app.Panel.RegisterHandler("system_startup_add", app.Guarded(s.onSystemStartupAdd))
	app.Panel.RegisterHandler("system_startup_remove", app.Guarded(s.onSystemStartupRemove))
	app.Panel.RegisterHandler("system_export", app.Guarded(s.onSystemExport))
	app.Panel.RegisterHandler("system_software_list", app.Guarded(s.onSystemSoftwareList))
	app.Agent.RegisterHandler("system_static", s.onSystemResultType("static"))
	app.Agent.RegisterHandler("system_live", s.onSystemResultType("live"))
	app.Agent.RegisterHandler("system_clipboard", s.onSystemResultType("clipboard"))
	app.Agent.RegisterHandler("system_startup", s.onSystemResultType("startup"))
	app.Agent.RegisterHandler("system_export", s.onSystemResultType("export"))
	app.Agent.RegisterHandler("system_software", s.onSystemResultType("software"))
	return s
}

func (s *Service) onSystemOpen(panelConnID string, payload json.RawMessage) {
	s.app.AppOpen(panelConnID, payload, "system", "system_open", "system_refresh")
}

func (s *Service) onSystemClose(panelConnID string, payload json.RawMessage) {
	s.app.AppClose(panelConnID, payload, "system", "system_close")
}

func (s *Service) forward(agentMsg string) func(string, json.RawMessage) {
	return func(_ string, payload json.RawMessage) {
		var p struct {
			AgentID string `json:"agent_id"`
		}
		if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
			return
		}
		for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
			s.app.Agent.EmitTo(conn, agentMsg, nil)
		}
	}
}

func (s *Service) onSystemRefresh(panelConnID string, payload json.RawMessage) {
	s.forward("system_refresh")(panelConnID, payload)
}

func (s *Service) onSystemClipGet(panelConnID string, payload json.RawMessage) {
	s.forward("system_clipboard_get")(panelConnID, payload)
}

func (s *Service) onSystemClipSet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		Text    string `json:"text"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "system_clipboard_set", map[string]string{"text": p.Text})
	}
}

func (s *Service) onSystemStartupList(panelConnID string, payload json.RawMessage) {
	s.forward("system_startup_list")(panelConnID, payload)
}

func (s *Service) onSystemStartupAdd(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "system_startup_add", p.Data)
	}
}

func (s *Service) onSystemStartupRemove(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "system_startup_remove", p.Data)
	}
}

func (s *Service) onSystemExport(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "system_export", nil)
	}
}

func (s *Service) onSystemSoftwareList(panelConnID string, payload json.RawMessage) {
	s.forward("system_software_list")(panelConnID, payload)
}

func (s *Service) onSystemResultType(resultType string) func(string, json.RawMessage) {
	return func(agentConnID string, payload json.RawMessage) {
		id, ok := s.app.Reg.AgentFor(agentConnID)
		if !ok {
			return
		}
		var p map[string]interface{}
		if json.Unmarshal(payload, &p) != nil {
			return
		}
		p["agent_id"] = id
		s.app.Publish("system:"+id, "system_"+resultType, p)
	}
}
