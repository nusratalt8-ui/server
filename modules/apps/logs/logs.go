package logs

import (
	"encoding/json"
	"sync"

	"agentmanager/modules/panelapp"
)

const maxLogBuf = 500

type logEntry struct {
	AgentID string `json:"agent_id"`
	Level   string `json:"level"`
	Msg     string `json:"msg"`
	Time    int64  `json:"time"`
}

var (
	logMu  sync.RWMutex
	logBuf = map[string][]logEntry{}
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("logs_get", app.Guarded(s.onLogsGet))
	app.Agent.RegisterHandler("agent_log", s.onAgentLog)
	return s
}

func (s *Service) onAgentLog(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var e struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Time  int64  `json:"time"`
	}
	if json.Unmarshal(payload, &e) != nil {
		return
	}
	entry := logEntry{AgentID: id, Level: e.Level, Msg: e.Msg, Time: e.Time}
	logMu.Lock()
	buf := logBuf[id]
	buf = append(buf, entry)
	if len(buf) > maxLogBuf {
		buf = buf[len(buf)-maxLogBuf:]
	}
	logBuf[id] = buf
	logMu.Unlock()
	s.app.Publish("logs:"+id, "agent_log", entry)
}

func (s *Service) onLogsGet(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "logs:"+p.AgentID)
	logMu.RLock()
	entries := logBuf[p.AgentID]
	out := make([]logEntry, len(entries))
	copy(out, entries)
	logMu.RUnlock()
	s.app.Panel.EmitTo(panelConnID, "logs_history", map[string]interface{}{
		"agent_id": p.AgentID,
		"entries":  out,
	})
}
