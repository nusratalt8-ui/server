package wsutil

import (
	"encoding/json"
	"sync"
	"time"
)

type Debouncer struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
	d      time.Duration
}

func NewDebouncer(d time.Duration) *Debouncer {
	return &Debouncer{
		timers: make(map[string]*time.Timer),
		d:      d,
	}
}

// Wrap returns a HandlerFunc that debounces calls per clientID.
// Only the last call within the window fires.
func (db *Debouncer) Wrap(fn HandlerFunc) HandlerFunc {
	return func(clientID string, payload json.RawMessage) {
		db.mu.Lock()
		defer db.mu.Unlock()
		if t, ok := db.timers[clientID]; ok {
			t.Stop()
		}
		p := make(json.RawMessage, len(payload))
		copy(p, payload)
		db.timers[clientID] = time.AfterFunc(db.d, func() {
			fn(clientID, p)
			db.mu.Lock()
			delete(db.timers, clientID)
			db.mu.Unlock()
		})
	}
}
