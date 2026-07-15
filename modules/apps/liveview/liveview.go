package liveview

import (
	"encoding/json"

	"agentmanager/modules/logger"
	"agentmanager/modules/panelapp"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("liveview_start", app.Guarded(s.onLiveviewStart))
	app.Panel.RegisterHandler("liveview_stop", app.Guarded(s.onLiveviewStop))
	app.Panel.RegisterHandler("liveview_input", app.Guarded(s.onLiveviewInput))
	app.Panel.RegisterHandler("mic_start", app.Guarded(s.onMicStart))
	app.Panel.RegisterHandler("mic_stop", app.Guarded(s.onMicStop))
	app.Panel.RegisterHandler("mic_list", app.Guarded(s.onMicList))
	app.Agent.RegisterHandler("mic_devices", s.onMicDevices)
    app.RegisterBinaryMsg("liveview_frame", s.onLiveviewFrame)
    app.RegisterBinaryMsg("mic_frame", s.onMicFrame)
	return s
}

func (s *Service) onLiveviewStart(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		SessID  string `json:"sess_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" || p.SessID == "" {
		return
	}
	logger.Infof("liveview: start agent=%s sess=%s panel=%s", p.AgentID, p.SessID, panelConnID)
	topic := "liveview:" + p.AgentID
	alreadyStreaming := s.app.Sess.Count(topic) > 0
	s.app.Sess.Subscribe(panelConnID, topic)
	if !alreadyStreaming {
		for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
			s.app.Agent.EmitTo(conn, "liveview_start", map[string]interface{}{
				"sess_id": p.SessID,
			})
		}
	}
}

func (s *Service) onLiveviewStop(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		SessID  string `json:"sess_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" || p.SessID == "" {
		return
	}
	topic := "liveview:" + p.AgentID
	s.app.Sess.Unsubscribe(panelConnID, topic)
	if s.app.Sess.Count(topic) > 0 {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "liveview_stop", map[string]interface{}{
			"sess_id": p.SessID,
		})
	}
}

func (s *Service) onLiveviewInput(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Input   json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "liveview_input", p.Input)
	}
}

func (s *Service) onLiveviewFrame(agentConnID string, payload []byte) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	tagged := make([]byte, 1+len(payload))
	tagged[0] = 0x01
	copy(tagged[1:], payload)
	for _, conn := range s.app.Sess.Subscribers("liveview:" + id) {
		s.app.Panel.EmitFrameRaw(conn, tagged)
	}
}

func (s *Service) onMicStart(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID  string `json:"agent_id"`
		SessID   string `json:"sess_id"`
		DeviceID int    `json:"device_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "mic_start", map[string]interface{}{
			"device_id": p.DeviceID,
		})
	}
}

func (s *Service) onMicStop(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		SessID  string `json:"sess_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "mic_stop", nil)
	}
}

func (s *Service) onMicList(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		logger.Infof("ic] onMicList: bad payload")
		return
	}
	conns := s.app.Reg.ConnsFor(p.AgentID)
	logger.Infof("ic] onMicList: agent=%s conns=%d", p.AgentID, len(conns))
	if len(conns) > 0 {
		s.app.Agent.EmitTo(conns[0], "mic_list", nil)
	}
}

func (s *Service) onMicDevices(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		logger.Infof("ic] onMicDevices: no agent for conn=%s", agentConnID)
		return
	}
	var devices interface{}
	if json.Unmarshal(payload, &devices) != nil {
		logger.Infof("ic] onMicDevices: bad payload agent=%s raw=%s", id, string(payload))
		return
	}
	topic := "liveview:" + id
	subs := s.app.Sess.Subscribers(topic)
	logger.Infof("ic] onMicDevices: agent=%s topic=%s subscribers=%d payload=%s", id, topic, len(subs), string(payload))
	s.app.Publish(topic, "mic_devices", map[string]interface{}{"agent_id": id, "devices": devices})
}

func (s *Service) onMicFrame(agentConnID string, payload []byte) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	tagged := make([]byte, 1+len(payload))
	tagged[0] = 0x02
	copy(tagged[1:], payload)
	for _, conn := range s.app.Sess.Subscribers("liveview:" + id) {
		s.app.Panel.EmitFrameRaw(conn, tagged)
	}
}
