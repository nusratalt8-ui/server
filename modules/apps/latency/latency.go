package latency

import (
	"encoding/json"
	"sync"
	"time"

	"agentmanager/modules/panelapp"
)

const maxLatencyBuf = 300

type latencyPoint struct {
	AgentID string `json:"agent_id"`
	MS      int    `json:"ms"`
	T       int64  `json:"t"`
}

var (
	latMu  sync.RWMutex
	latBuf = map[string][]latencyPoint{}
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("latency_open", app.Guarded(s.onLatencyOpen))
	app.Panel.RegisterHandler("latency_close", app.Guarded(s.onLatencyClose))
	return s
}

func (s *Service) HandlePing(agentConnID string, ms int) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	pt := latencyPoint{AgentID: id, MS: ms, T: time.Now().Unix()}
	latMu.Lock()
	buf := latBuf[id]
	buf = append(buf, pt)
	if len(buf) > maxLatencyBuf {
		buf = buf[len(buf)-maxLatencyBuf:]
	}
	latBuf[id] = buf
	latMu.Unlock()
	s.app.Publish("latency:"+id, "agent_ping", map[string]interface{}{"id": id, "ms": ms, "t": pt.T})
}

func (s *Service) onLatencyOpen(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.SetLatencyActive(conn, true)
	}
	s.app.Sess.Subscribe(panelConnID, "latency:"+p.AgentID)
	latMu.RLock()
	pts := latBuf[p.AgentID]
	out := make([]latencyPoint, len(pts))
	copy(out, pts)
	latMu.RUnlock()
	s.app.Panel.EmitTo(panelConnID, "latency_history", map[string]interface{}{
		"agent_id": p.AgentID,
		"points":   out,
	})
}

func (s *Service) onLatencyClose(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Unsubscribe(panelConnID, "latency:"+p.AgentID)
	if s.app.Sess.Count("latency:"+p.AgentID) > 0 {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.SetLatencyActive(conn, false)
	}
}
