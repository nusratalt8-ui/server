package wsutil

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"agentmanager/modules/crypto"
)

type Client struct {
	id         string
	tag        string
	conn       *websocket.Conn
	send       chan []byte
	frame      chan []byte
	done       chan struct{}
	hub        *Hub
	msgLimiter *rate.Limiter
	violations int
	probeAt    int64
}

type TaggerFunc func(c echo.Context) string

func (h *Hub) SetTagger(fn TaggerFunc) {
	h.mu.Lock()
	h.tagger = fn
	h.mu.Unlock()
}

func (h *Hub) runTagger(c echo.Context) string {
	h.mu.RLock()
	fn := h.tagger
	h.mu.RUnlock()
	if fn == nil {
		return ""
	}
	return fn(c)
}

func (c *Client) close() {
	c.conn.Close()
}

func newUpgrader(cfg Config) *websocket.Upgrader {
	rb := cfg.ReadBufferSize
	if rb <= 0 {
		rb = 4096
	}
	wb := cfg.WriteBufferSize
	if wb <= 0 {
		wb = 4096
	}
	return &websocket.Upgrader{
		ReadBufferSize:  rb,
		WriteBufferSize: wb,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
}

func (h *Hub) HandleConnect(c echo.Context) error {
	conn, err := newUpgrader(h.cfg).Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.mu.RLock()
		fn := h.onConnectLog
		h.mu.RUnlock()
		if fn != nil {
			fn("", c.Request().RemoteAddr)
		}
		return ErrUpgradeFailed
	}
	ip := c.Request().RemoteAddr
	client := &Client{
		id:         crypto.RandomHex(6),
		tag:        h.runTagger(c),
		conn:       conn,
		send:       make(chan []byte, h.cfg.SendBuffer),
		frame:      make(chan []byte, 1),
		done:       make(chan struct{}),
		hub:        h,
		msgLimiter: rate.NewLimiter(rate.Limit(h.cfg.MsgPerSecond), h.cfg.MsgBurst),
	}
	h.register(client)
	h.mu.RLock()
	fn := h.onConnectLog
	h.mu.RUnlock()
	if fn != nil {
		fn(client.id, ip)
	}
	go client.writePump()
	go client.readPump()
	if h.cfg.EnableProbe {
		go client.probePump()
	}
	h.fireConnect(client.id)
	return nil
}

func (c *Client) probePump() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if !c.hub.isLatencyActive(c.id) {
				continue
			}
			ts := time.Now().UnixMilli()
			atomic.StoreInt64(&c.probeAt, ts)
			data, _ := json.Marshal(map[string]interface{}{
				"type":    "latency_probe",
				"payload": map[string]int64{"ts": ts},
			})
			select {
			case c.send <- data:
			case <-c.done:
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		close(c.done)
		c.conn.Close()
		c.hub.fireDisconnect(c.id)
	}()
	c.conn.SetReadLimit(c.hub.cfg.MaxFrameSize)
	c.conn.SetReadDeadline(time.Now().Add(c.hub.cfg.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.hub.cfg.PongWait))
		return nil
	})
	for {
		mt, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if mt == websocket.BinaryMessage {
			c.hub.routeBinary(c.id, data)
			continue
		}
		if mt == websocket.TextMessage {
			h := c.hub
			h.mu.RLock()
			fn := h.textHandler
			h.mu.RUnlock()
			if fn != nil {
				go fn(c.id, data)
				continue
			}
		}
		var head struct {
			Type string `json:"type"`
		}
		typed := json.Unmarshal(data, &head) == nil

		if typed && head.Type == "latency_pong" {
			sent := atomic.LoadInt64(&c.probeAt)
			if sent > 0 {
				ms := int(time.Now().UnixMilli() - sent)
				c.hub.firePing(c.id, ms)
			}
			continue
		}

		if int64(len(data)) > c.hub.cfg.MaxMessageSize {
			c.violations++
			if c.violations >= c.hub.cfg.MaxViolations {
				return
			}
			continue
		}
		if head.Type != "liveview_input" {
			if !c.msgLimiter.Allow() {
				c.violations++
				if c.violations >= c.hub.cfg.MaxViolations {
					return
				}
				continue
			}
		}
		c.violations = 0
		c.hub.route(c.id, data)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.hub.cfg.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.cfg.WriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case data := <-c.frame:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.cfg.WriteWait))
			if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.hub.cfg.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}
