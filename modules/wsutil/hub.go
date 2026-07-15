package wsutil

import (
	"encoding/json"
	"sync"
	"time"

	"agentmanager/modules/logger"
)

type HandlerFunc func(clientID string, payload json.RawMessage)

type inboundItem struct {
	clientID string
	data     []byte
}

type Hub struct {
	mu              sync.RWMutex
	clients         map[string]*Client
	handlers        map[string]HandlerFunc
	cfg             Config
	onConnect       func(clientID string)
	onDisconnect    func(clientID string)
	onPing          func(clientID string, ms int)
	tagger          TaggerFunc
	offlineTimers   map[string]*time.Timer
	latencyActive   map[string]struct{}
	inbound         chan inboundItem
	binaryHandler   func(clientID string, data []byte)
	textHandler     func(clientID string, data []byte)
	onConnectLog    func(id string, ip string)
	onDisconnectLog func(id string)
}

func NewHub(cfg Config) *Hub {
	queue := cfg.InboundQueue
	if queue <= 0 {
		queue = 8192
	}
	h := &Hub{
		clients:       make(map[string]*Client),
		handlers:      make(map[string]HandlerFunc),
		cfg:           cfg,
		offlineTimers: make(map[string]*time.Timer),
		latencyActive: make(map[string]struct{}),
		inbound:       make(chan inboundItem, queue),
	}
	workers := cfg.InboundWorkers
	if workers <= 0 {
		workers = 32
	}
	for i := 0; i < workers; i++ {
		go h.processInbound()
	}
	return h
}

func (h *Hub) SetLatencyActive(clientID string, active bool) {
	h.mu.Lock()
	if active {
		h.latencyActive[clientID] = struct{}{}
	} else {
		delete(h.latencyActive, clientID)
	}
	h.mu.Unlock()
}

func (h *Hub) isLatencyActive(clientID string) bool {
	h.mu.RLock()
	_, ok := h.latencyActive[clientID]
	h.mu.RUnlock()
	return ok
}

func (h *Hub) RegisterHandler(msgType string, fn HandlerFunc) {
	h.mu.Lock()
	h.handlers[msgType] = fn
	h.mu.Unlock()
}

func (h *Hub) RegisterBinaryHandler(fn func(clientID string, data []byte)) {
	h.mu.Lock()
	h.binaryHandler = fn
	h.mu.Unlock()
}

func (h *Hub) RegisterTextHandler(fn func(clientID string, data []byte)) {
	h.mu.Lock()
	h.textHandler = fn
	h.mu.Unlock()
}

func (h *Hub) routeBinary(clientID string, data []byte) {
	h.mu.RLock()
	fn := h.binaryHandler
	h.mu.RUnlock()
	if fn != nil {
		go func() {
			defer func() { recover() }()
			fn(clientID, data)
		}()
	}
}

func (h *Hub) SetOnConnect(fn func(clientID string)) {
	h.mu.Lock()
	h.onConnect = fn
	h.mu.Unlock()
}

func (h *Hub) SetOnDisconnect(fn func(clientID string)) {
	h.mu.Lock()
	h.onDisconnect = fn
	h.mu.Unlock()
}

func (h *Hub) SetConnectLogger(fn func(id string, ip string)) {
	h.mu.Lock()
	h.onConnectLog = fn
	h.mu.Unlock()
}

func (h *Hub) SetDisconnectLogger(fn func(id string)) {
	h.mu.Lock()
	h.onDisconnectLog = fn
	h.mu.Unlock()
}

func (h *Hub) SetOnPing(fn func(clientID string, ms int)) {
	h.mu.Lock()
	h.onPing = fn
	h.mu.Unlock()
}

func (h *Hub) firePing(id string, ms int) {
	h.mu.RLock()
	fn := h.onPing
	h.mu.RUnlock()
	if fn != nil {
		fn(id, ms)
	}
}

func (h *Hub) fireConnect(id string) {
	h.mu.RLock()
	fn := h.onConnect
	h.mu.RUnlock()
	if fn != nil {
		fn(id)
	}
}

func (h *Hub) fireDisconnect(id string) {
	h.mu.RLock()
	grace := h.cfg.OfflineGracePeriod
	h.mu.RUnlock()
	if grace > 0 {
		return
	}
	h.mu.RLock()
	fn := h.onDisconnect
	h.mu.RUnlock()
	if fn != nil {
		fn(id)
	}
}

func (h *Hub) IDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for id := range h.clients {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	if t, ok := h.offlineTimers[c.id]; ok {
		t.Stop()
		delete(h.offlineTimers, c.id)
	}
	h.clients[c.id] = c
	h.mu.Unlock()
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	if cur := h.clients[c.id]; cur == c {
		delete(h.clients, c.id)
		if h.onDisconnectLog != nil {
			h.onDisconnectLog(c.id)
		}
		grace := h.cfg.OfflineGracePeriod
		fn := h.onDisconnect
		if fn != nil && grace > 0 {
			id := c.id
			h.offlineTimers[id] = time.AfterFunc(grace, func() {
				h.mu.Lock()
				_, pending := h.offlineTimers[id]
				if pending {
					delete(h.offlineTimers, id)
				}
				h.mu.Unlock()
				if pending {
					fn(id)
				}
			})
			h.mu.Unlock()
			return
		}
	}
	h.mu.Unlock()
}

func (h *Hub) processInbound() {
	for item := range h.inbound {
		h.dispatch(item.clientID, item.data)
	}
}

// route hands an inbound message to the worker pool instead of
// dispatching it on the connection's read goroutine. It never blocks
// the read loop: if the pool is saturated the message is dropped (and
// logged) rather than stalling reads — which is what kept a flood of
// live-view frames from holding up fs/proc/panel handlers on the same
// connection.
func (h *Hub) route(clientID string, data []byte) {
	var head struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(data, &head) == nil && head.Type == "liveview_input" {
		go h.dispatch(clientID, data)
		return
	}
	select {
	case h.inbound <- inboundItem{clientID: clientID, data: data}:
	default:
		logger.Warnf("ws: inbound queue full, dropped message from %s", clientID)
	}
}

func (h *Hub) dispatch(clientID string, raw []byte) {
	var in struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return
	}
	h.mu.RLock()
	fn := h.handlers[in.Type]
	h.mu.RUnlock()
	if fn != nil {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Warnf("ws: handler panic (type=%s client=%s): %v", in.Type, clientID, rec)
			}
		}()
		fn(clientID, in.Payload)
	}
}

func (h *Hub) deliver(c *Client, data []byte) {
	select {
	case c.send <- data:
	case <-c.done:
	default:
	}
}

func (h *Hub) Emit(msgType string, payload interface{}) {
	data, ok := encode(Message{Type: msgType, Payload: payload})
	if !ok {
		return
	}
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients))
	for _, c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()
	for _, c := range targets {
		h.deliver(c, data)
	}
}

func (h *Hub) EmitTo(clientID string, msgType string, payload interface{}) {
	data, ok := encode(Message{Type: msgType, Payload: payload})
	if !ok {
		return
	}
	h.mu.RLock()
	c := h.clients[clientID]
	h.mu.RUnlock()
	if c != nil {
		h.deliver(c, data)
	}
}

// EmitToTag sends to every connection carrying the given tag. For the panel
// hub the tag is the account id, so this is how a broadcast reaches only the
// connections belonging to one account instead of Emit's send-to-everyone.
func (h *Hub) EmitToTag(tag string, msgType string, payload interface{}) {
	if tag == "" {
		return
	}
	data, ok := encode(Message{Type: msgType, Payload: payload})
	if !ok {
		return
	}
	h.mu.RLock()
	targets := make([]*Client, 0)
	for _, c := range h.clients {
		if c.tag == tag {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range targets {
		h.deliver(c, data)
	}
}

func (h *Hub) deliverFrame(c *Client, data []byte) {
	for {
		select {
		case c.frame <- data:
			return
		case <-c.done:
			return
		default:
			select {
			case <-c.frame:
			default:
			}
		}
	}
}

func (h *Hub) EmitFrameRaw(clientID string, data []byte) {
	h.mu.RLock()
	c := h.clients[clientID]
	h.mu.RUnlock()
	if c != nil {
		h.deliverFrame(c, data)
	}
}

func (h *Hub) EmitTextRaw(clientID string, data []byte) {
	h.mu.RLock()
	c := h.clients[clientID]
	h.mu.RUnlock()
	if c != nil {
		h.deliver(c, data)
	}
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) TagOf(clientID string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if c, ok := h.clients[clientID]; ok {
		return c.tag
	}
	return ""
}

func (h *Hub) ConnsByTag(tag string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]string, 0)
	for id, c := range h.clients {
		if c.tag == tag {
			out = append(out, id)
		}
	}
	return out
}

func (h *Hub) HasClient(clientID string) bool {
	h.mu.RLock()
	_, ok := h.clients[clientID]
	h.mu.RUnlock()
	return ok
}

func (h *Hub) CloseByTag(tag string) {
	if tag == "" {
		return
	}
	h.mu.RLock()
	targets := make([]*Client, 0)
	for _, c := range h.clients {
		if c.tag == tag {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range targets {
		c.close()
	}
}
