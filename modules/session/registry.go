package session

import "sync"

// Registry tracks panel sessions (one per panel websocket connection,
// keyed by the hub-assigned connID) and which per-agent topics each one
// is subscribed to. A topic is "<feature>:<agentID>", e.g.
// "liveview:ab12cd". Multiple sessions may subscribe to the same topic
// (two people watching one agent's live view) — Subscribers returns all
// of them, which is how shared viewing stays intact.
type Registry struct {
	mu      sync.RWMutex
	byConn  map[string]map[string]struct{} // connID -> set of topics
	byTopic map[string]map[string]struct{} // topic  -> set of connIDs
}

func New() *Registry {
	return &Registry{
		byConn:  make(map[string]map[string]struct{}),
		byTopic: make(map[string]map[string]struct{}),
	}
}

func (r *Registry) Subscribe(connID, topic string) {
	if connID == "" || topic == "" {
		return
	}
	r.mu.Lock()
	if r.byConn[connID] == nil {
		r.byConn[connID] = make(map[string]struct{})
	}
	r.byConn[connID][topic] = struct{}{}
	if r.byTopic[topic] == nil {
		r.byTopic[topic] = make(map[string]struct{})
	}
	r.byTopic[topic][connID] = struct{}{}
	r.mu.Unlock()
}

func (r *Registry) Unsubscribe(connID, topic string) {
	r.mu.Lock()
	if topics, ok := r.byConn[connID]; ok {
		delete(topics, topic)
		if len(topics) == 0 {
			delete(r.byConn, connID)
		}
	}
	if conns, ok := r.byTopic[topic]; ok {
		delete(conns, connID)
		if len(conns) == 0 {
			delete(r.byTopic, topic)
		}
	}
	r.mu.Unlock()
}

// Subscribers returns the connIDs currently subscribed to topic.
func (r *Registry) Subscribers(topic string) []string {
	r.mu.RLock()
	conns := r.byTopic[topic]
	out := make([]string, 0, len(conns))
	for c := range conns {
		out = append(out, c)
	}
	r.mu.RUnlock()
	return out
}

// Count reports how many sessions are subscribed to topic. Used to
// reference-count agent-side streams so the last viewer leaving is the
// one that tells the agent to stop.
func (r *Registry) Count(topic string) int {
	r.mu.RLock()
	n := len(r.byTopic[topic])
	r.mu.RUnlock()
	return n
}

// Remove drops a session entirely (wired to the panel hub's
// OnDisconnect) so a closed tab can't leak subscriptions.
// Returns topics that hit zero subscribers so callers can
// stop any agent-side streams that lost their last viewer.
func (r *Registry) Remove(connID string) []string {
	var emptied []string
	r.mu.Lock()
	for topic := range r.byConn[connID] {
		if conns, ok := r.byTopic[topic]; ok {
			delete(conns, connID)
			if len(conns) == 0 {
				delete(r.byTopic, topic)
				emptied = append(emptied, topic)
			}
		}
	}
	delete(r.byConn, connID)
	r.mu.Unlock()
	return emptied
}
