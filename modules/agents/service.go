package agents

import (
	"encoding/json"
	"sync"
	"time"

	"agentmanager/modules/admin"
	"agentmanager/modules/crypto"
	"agentmanager/modules/dbutil"
	"agentmanager/modules/wsutil"
)

type Agent struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	OwnerUsername string `json:"owner_username,omitempty"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Hostname      string `json:"hostname"`
	Username      string `json:"username"`
	Online        bool   `json:"online"`
	LastSeen      int64  `json:"last_seen"`
}

type startupPayload struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Hostname    string `json:"hostname"`
	Username    string `json:"username"`
}

type connInfo struct {
	agentID string
	userID  string
}

type Service struct {
	db       *dbutil.DB
	agent    *wsutil.Hub
	panel    *wsutil.Hub
	mu       sync.Mutex
	conns    map[string]connInfo
	onPingCb func(connID, agentID string, ms int)
}

const agentsPageSize = 50

type UserCounts struct {
	Total    int    `json:"total"`
	Online   int    `json:"online"`
	Username string `json:"username,omitempty"`
}

type agentsPageResponse struct {
	Items       []Agent               `json:"items"`
	Total       int                   `json:"total"`
	OnlineTotal int                   `json:"online_total"`
	Offset      int                   `json:"offset"`
	PerUser     map[string]UserCounts `json:"per_user,omitempty"`
}

func NewService(db *dbutil.DB, agentHub, panelHub *wsutil.Hub) *Service {
	s := &Service{db: db, agent: agentHub, panel: panelHub, conns: make(map[string]connInfo)}
	s.resetAllOffline()
	agentHub.RegisterHandler("startup", s.onStartup)
	agentHub.SetOnDisconnect(s.onDisconnect)
	agentHub.SetOnPing(s.onPing)
	panelHub.SetOnConnect(s.onPanelConnect)
	return s
}

func (s *Service) onStartup(connID string, payload json.RawMessage) {
	var p startupPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	owner := s.agent.TagOf(connID)
	if owner == "" {
		return
	}

	id := p.ID
	generated := false
	if id == "" || !s.exists(id, owner) {
		id = crypto.RandomHex(16)
		generated = true
	}

	name := p.DisplayName
	if name == "" {
		name = "Agent " + id
	}
	a := Agent{ID: id, UserID: owner, Name: name, Description: p.Description, Hostname: p.Hostname, Username: p.Username}
	if err := s.upsert(a, time.Now().Unix()); err != nil {
		return
	}

	s.mu.Lock()
	s.conns[connID] = connInfo{agentID: id, userID: owner}
	s.mu.Unlock()

	if generated {
		s.agent.EmitTo(connID, "identity", map[string]string{"id": id})
	}
	s.broadcast(owner)
}

func (s *Service) onDisconnect(connID string) {
	s.mu.Lock()
	info, ok := s.conns[connID]
	delete(s.conns, connID)
	stillOnline := false
	for _, v := range s.conns {
		if v.agentID == info.agentID {
			stillOnline = true
			break
		}
	}
	s.mu.Unlock()
	if !ok || stillOnline {
		return
	}
	s.setOffline(info.agentID, time.Now().Unix())
	s.broadcast(info.userID)
}

func (s *Service) onPanelConnect(panelClientID string) {
	owner := s.panel.TagOf(panelClientID)
	items, err := s.listPaged(owner, agentsPageSize, 0)
	if err != nil {
		return
	}
	total, online, err := s.countsForUser(owner)
	if err != nil {
		return
	}
	s.panel.EmitTo(panelClientID, "agents", agentsPageResponse{
		Items: items, Total: total, OnlineTotal: online, Offset: 0,
	})
	if admins.IsAdmin(owner) {
		allItems, err := s.listAllPaged(agentsPageSize, 0)
		if err != nil {
			return
		}
		perUser, totalAll, onlineAll, err := s.countsPerUser()
		if err != nil {
			return
		}
		s.panel.EmitTo(panelClientID, "admin_agents", agentsPageResponse{
			Items: allItems, Total: totalAll, OnlineTotal: onlineAll, Offset: 0, PerUser: perUser,
		})
	}
}

func (s *Service) ConnsFor(agentID string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, 1)
	for conn, info := range s.conns {
		if info.agentID == agentID {
			out = append(out, conn)
		}
	}
	return out
}

func (s *Service) AgentFor(connID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, ok := s.conns[connID]
	return info.agentID, ok
}

func (s *Service) onPing(connID string, ms int) {
	s.mu.Lock()
	info, ok := s.conns[connID]
	cb := s.onPingCb
	s.mu.Unlock()
	if !ok {
		return
	}
	if cb != nil {
		cb(connID, info.agentID, ms)
	}
}

func (s *Service) SetOnPingCb(fn func(connID, agentID string, ms int)) {
	s.mu.Lock()
	s.onPingCb = fn
	s.mu.Unlock()
}

func (s *Service) OwnerOf(agentID string) (string, bool) {
	return s.ownerOf(agentID)
}

func (s *Service) DeleteAndBroadcast(agentID, requesterUID string) bool {
	a, ok := s.getByID(agentID)
	if !ok {
		return false
	}
	if a.UserID != requesterUID && !admins.IsAdmin(requesterUID) {
		return false
	}
	if err := s.deleteAgent(agentID); err != nil {
		return false
	}
	s.broadcast(a.UserID)
	return true
}

func (s *Service) broadcast(userID string) {
	if userID == "" {
		return
	}
	items, err := s.listPaged(userID, agentsPageSize, 0)
	if err != nil {
		return
	}
	total, online, err := s.countsForUser(userID)
	if err != nil {
		return
	}
	s.panel.EmitToTag(userID, "agents", agentsPageResponse{
		Items: items, Total: total, OnlineTotal: online, Offset: 0,
	})
	s.broadcastAdmin()
}

func (s *Service) broadcastAdmin() {
	allItems, err := s.listAllPaged(agentsPageSize, 0)
	if err != nil {
		return
	}
	perUser, totalAll, onlineAll, err := s.countsPerUser()
	if err != nil {
		return
	}
	resp := agentsPageResponse{Items: allItems, Total: totalAll, OnlineTotal: onlineAll, Offset: 0, PerUser: perUser}
	for _, adminID := range admins.AdminIDs() {
		s.panel.EmitToTag(adminID, "admin_agents", resp)
	}
}
