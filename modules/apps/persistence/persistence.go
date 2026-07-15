package persistence

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("persistence_get", app.Guarded(s.onPersistenceGet))
	app.Panel.RegisterHandler("persistence_toggle", app.Guarded(s.onPersistenceToggle))
	app.Agent.RegisterHandler("persistence_status", s.onPersistenceStatus)
	app.Agent.RegisterHandler("persistence_update", s.onPersistenceUpdate)
	return s
}

func (s *Service) onPersistenceGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "persistence:"+p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "persistence_get", payload)
	}
}

func (s *Service) onPersistenceToggle(_ string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "persistence_toggle", payload)
	}
}

func (s *Service) onPersistenceStatus(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	s.app.Publish("persistence:"+id, "persistence_status", payload)
}

func (s *Service) onPersistenceUpdate(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	s.app.Publish("persistence:"+id, "persistence_update", payload)
}
