package panelapp

import (
    "encoding/json"
    "sync"

    "agentmanager/modules/admin"
    "agentmanager/modules/agents"
    "agentmanager/modules/session"
    "agentmanager/modules/wsutil"
)

type PlanChecker interface {
    IsPaid(userID string) bool
}

type App struct {
    Agent *wsutil.Hub
    Panel *wsutil.Hub
    Reg   *agents.Service
    Sess  *session.Registry
    Plan  PlanChecker

    binMu       sync.RWMutex
    binHandlers map[string]func(agentConnID string, payload []byte)
}

func (a *App) RegisterBinaryMsg(msgType string, fn func(agentConnID string, payload []byte)) {
    a.binMu.Lock()
    defer a.binMu.Unlock()
    if a.binHandlers == nil {
        a.binHandlers = map[string]func(string, []byte){}
        a.Agent.RegisterBinaryHandler(a.dispatchBinary)
    }
    a.binHandlers[msgType] = fn
}

func (a *App) dispatchBinary(agentConnID string, data []byte) {
    if len(data) == 0 {
        return
    }
    typeLen := int(data[0])
    if len(data) < 1+typeLen {
        return
    }
    msgType := string(data[1 : 1+typeLen])
    payload := data[1+typeLen:]
    a.binMu.RLock()
    fn := a.binHandlers[msgType]
    a.binMu.RUnlock()
    if fn != nil {
        fn(agentConnID, payload)
    }
}

func (a *App) Paid(fn func(string, json.RawMessage)) func(string, json.RawMessage) {
	return func(panelConnID string, payload json.RawMessage) {
		var p struct {
			AgentID string `json:"agent_id"`
		}
		if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
			return
		}
		owner, ok := a.Reg.OwnerOf(p.AgentID)
		if !ok || owner == "" {
			return
		}
		if a.Plan != nil && !a.Plan.IsPaid(owner) && !admins.IsAdmin(owner) {
			a.Panel.EmitTo(panelConnID, "plan_error", map[string]string{
				"msg": "upgrade your plan broke ass nigga",
			})
			return
		}
		fn(panelConnID, payload)
	}
}

func (a *App) EmitFrameRaw(panelConnID string, data []byte) {
	a.Panel.EmitFrameRaw(panelConnID, data)
}

func (a *App) Guarded(fn func(string, json.RawMessage)) func(string, json.RawMessage) {
	return func(panelConnID string, payload json.RawMessage) {
		var p struct {
			AgentID string `json:"agent_id"`
		}
		if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
			return
		}
		owner, ok := a.Reg.OwnerOf(p.AgentID)
		if !ok || owner == "" {
			return
		}
		panelUser := a.Panel.TagOf(panelConnID)
		if owner != panelUser && !admins.IsAdmin(panelUser) {
			return
		}
		fn(panelConnID, payload)
	}
}

func (a *App) Publish(topic, msgType string, payload interface{}) {
	for _, connID := range a.Sess.Subscribers(topic) {
		a.Panel.EmitTo(connID, msgType, payload)
	}
}

func (a *App) AppOpen(panelConnID string, payload json.RawMessage, topicPrefix, openMsg, refreshMsg string) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	topic := topicPrefix + ":" + p.AgentID
	alreadyOpen := a.Sess.Count(topic) > 0
	a.Sess.Subscribe(panelConnID, topic)
	if alreadyOpen {
		for _, conn := range a.Reg.ConnsFor(p.AgentID) {
			a.Agent.EmitTo(conn, refreshMsg, nil)
		}
	} else {
		for _, conn := range a.Reg.ConnsFor(p.AgentID) {
			a.Agent.EmitTo(conn, openMsg, nil)
		}
	}
}

func (a *App) AppClose(panelConnID string, payload json.RawMessage, topicPrefix, closeMsg string) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	topic := topicPrefix + ":" + p.AgentID
	a.Sess.Unsubscribe(panelConnID, topic)
	if a.Sess.Count(topic) > 0 {
		return
	}
	for _, conn := range a.Reg.ConnsFor(p.AgentID) {
		a.Agent.EmitTo(conn, closeMsg, nil)
	}
}
