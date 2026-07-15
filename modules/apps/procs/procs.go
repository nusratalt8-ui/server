package procs

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("procwatch_start", app.Guarded(s.onProcWatchStart))
	app.Panel.RegisterHandler("procwatch_stop", app.Guarded(s.onProcWatchStop))
	app.Panel.RegisterHandler("proclist_get", app.Guarded(s.onProcListGet))
	app.Panel.RegisterHandler("proc_kill", app.Guarded(s.onProcKill))
	app.Agent.RegisterHandler("proclist_result", s.onProcListResult)
	app.Agent.RegisterHandler("task_update", s.onTaskUpdate)
	app.Agent.RegisterHandler("proc_kill_result", s.onProcKillResult)
	return s
}

func (s *Service) onProcWatchStart(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "procs:"+p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "procwatch_start", nil)
	}
}

func (s *Service) onProcWatchStop(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Unsubscribe(panelConnID, "procs:"+p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "procwatch_stop", nil)
	}
}

func (s *Service) onProcListGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "proclist_get", nil)
	}
}

func (s *Service) onProcKill(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "proc_kill", p.Data)
	}
}

func (s *Service) onProcListResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("procs:"+id, "proclist_result", p)
}

func (s *Service) onTaskUpdate(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("procs:"+id, "task_update", p)
}

func (s *Service) onProcKillResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("procs:"+id, "proc_kill_result", p)
}
