package message

import (
    "encoding/json"
    "time"

    "agentmanager/modules/admin"
    "agentmanager/modules/crypto"
    "agentmanager/modules/dbutil"
    "agentmanager/modules/session"
    "agentmanager/modules/wsutil"
)

type Registry interface {
	ConnsFor(agentID string) []string
	AgentFor(connID string) (string, bool)
	OwnerOf(agentID string) (string, bool)
}

type Resolver interface {
	Meta(id string) (filename, contentType string, size int64, ok bool)
}

type Service struct {
    db    *dbutil.DB
    cfg   Config
    agent *wsutil.Hub
    panel *wsutil.Hub
    reg   Registry
    sess  *session.Registry
    att   Resolver
}

func NewService(db *dbutil.DB, cfg Config, agentHub, panelHub *wsutil.Hub, reg Registry, sess *session.Registry, att Resolver) *Service {
    s := &Service{db: db, cfg: cfg, agent: agentHub, panel: panelHub, reg: reg, sess: sess, att: att}
    panelHub.RegisterHandler("chat_send", s.guarded(s.onPanelChat))
    agentHub.RegisterHandler("chat", s.onAgentChat)
    panelHub.RegisterHandler("ping", func(connID string, payload json.RawMessage) {
        panelHub.EmitTo(connID, "pong", payload)
    })
    return s
}

func (s *Service) publish(topic, msgType string, payload interface{}) {
    for _, connID := range s.sess.Subscribers(topic) {
        s.panel.EmitTo(connID, msgType, payload)
    }
}

func (s *Service) guarded(fn func(string, json.RawMessage)) func(string, json.RawMessage) {
    return func(panelConnID string, payload json.RawMessage) {
        var p struct {
            AgentID string `json:"agent_id"`
        }
        if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
            return
        }
        owner, ok := s.reg.OwnerOf(p.AgentID)
        if !ok || owner == "" {
            return
        }
        panelUser := s.panel.TagOf(panelConnID)
        if owner != panelUser && !admins.IsAdmin(panelUser) {
            return
        }
        fn(panelConnID, payload)
    }
}

func (s *Service) hydrate(ids []string) []Attachment {
	out := make([]Attachment, 0, len(ids))
	for _, id := range ids {
		a := Attachment{ID: id}
		if fn, ct, sz, ok := s.att.Meta(id); ok {
			a.Filename, a.ContentType, a.Size = fn, ct, sz
		}
		out = append(out, a)
	}
	return out
}

func (s *Service) validate(body string, ids []string, embed json.RawMessage) bool {
	if len(body) > s.cfg.MaxLength {
		return false
	}
	return body != "" || len(ids) > 0 || len(embed) > 0
}

func (s *Service) store(agentID, sender, body string, ids []string, embed json.RawMessage) (Message, bool) {
	if ids == nil {
		ids = []string{}
	}
	if !s.validate(body, ids, embed) {
		return Message{}, false
	}
	m := Message{
		ID:          crypto.RandomHex(12),
		AgentID:     agentID,
		Sender:      sender,
		Body:        body,
		Attachments: s.hydrate(ids),
		Embed:       embed,
		CreatedAt:   time.Now().Unix(),
	}
	if err := s.insert(m); err != nil {
		return Message{}, false
	}
	return m, true
}

func (s *Service) onPanelChat(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID     string   `json:"agent_id"`
		Text        string   `json:"text"`
		Attachments []string `json:"attachments"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.AgentID == "" {
		return
	}
	senderUID := s.panel.TagOf(panelConnID)
	isAdmin := admins.IsAdmin(senderUID)

	sender := senderUID
	if isAdmin {
		sender = "admin"
	}
	m, ok := s.store(p.AgentID, sender, p.Text, p.Attachments, nil)
	if !ok {
		return
	}

	replyTopic := "chat:" + p.AgentID + ":" + panelConnID
	s.sess.Subscribe(panelConnID, replyTopic)

	for _, conn := range s.reg.ConnsFor(p.AgentID) {
		s.agent.EmitTo(conn, "chat", map[string]interface{}{
			"text":        m.Body,
			"attachments": p.Attachments,
			"reply_to":    replyTopic,
		})
	}
	s.panel.EmitTo(panelConnID, "message", m)
	if !isAdmin {
		for _, adminID := range admins.AdminIDs() {
			s.panel.EmitToTag(adminID, "message", m)
		}
	}
}

func (s *Service) onAgentChat(agentConnID string, payload json.RawMessage) {
	var p struct {
		Text        string          `json:"text"`
		Attachments []string        `json:"attachments"`
		Embed       json.RawMessage `json:"embed"`
		ReplyTo     string          `json:"reply_to"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}
	id, ok := s.reg.AgentFor(agentConnID)
	if !ok {
		return
	}

	subscribers := s.sess.Subscribers(p.ReplyTo)

	sender := "agent"
	if len(subscribers) > 0 && admins.IsAdmin(s.panel.TagOf(subscribers[0])) {
		sender = "agent_admin"
	}

	m, ok := s.store(id, sender, p.Text, p.Attachments, p.Embed)
	if !ok {
		return
	}

	if len(subscribers) > 0 {
		s.publish(p.ReplyTo, "message", m)
		if sender != "agent_admin" {
			for _, adminID := range admins.AdminIDs() {
				s.panel.EmitToTag(adminID, "message", m)
			}
		}
	} else {
		s.panel.EmitToTag(s.agent.TagOf(agentConnID), "message", m)
		for _, adminID := range admins.AdminIDs() {
			s.panel.EmitToTag(adminID, "message", m)
		}
	}
}
