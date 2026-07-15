package terminal

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("terminal_exec", app.Guarded(s.onTerminalExec))
	app.Panel.RegisterHandler("terminal_ps_exec", app.Guarded(s.onTerminalPSExec))
	app.Agent.RegisterHandler("terminal_output", s.onTerminalOutput)
	app.Agent.RegisterHandler("terminal_ps_output", s.onTerminalPSOutput)
	return s
}

func (s *Service) onTerminalExec(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "terminal_exec", p.Data)
	}
}

func (s *Service) onTerminalPSExec(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "terminal_ps_exec", p.Data)
	}
}

func (s *Service) onTerminalOutput(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("terminal:"+id, "terminal_output", p)
}

func (s *Service) onTerminalPSOutput(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("terminal:"+id, "terminal_ps_output", p)
}
