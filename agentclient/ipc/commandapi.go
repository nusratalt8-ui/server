package ipc

import (
	"strings"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/config"
)

type command struct {
	desc string
	cat  string
	h    Handler
}

type Commands struct {
	mu       sync.Mutex
	prefix   string
	handlers map[string]command
	notFound Handler
}

func New() *Commands {
	cm := &Commands{prefix: config.Prefix, handlers: make(map[string]command)}
	cm.AddCommand("help", "list commands", "general", cm.help)
	cm.notFound = func(c *Context) error {
		return c.Reply("unknown command " + c.prefix + c.Name + " — try " + c.prefix + "help")
	}
	return cm
}

func (cm *Commands) AddCommand(name, desc, cat string, h Handler) {
	cm.mu.Lock()
	cm.handlers[strings.ToLower(name)] = command{desc: desc, cat: strings.ToLower(cat), h: h}
	cm.mu.Unlock()
}

func (cm *Commands) OnNotFound(h Handler) {
	cm.notFound = h
}

func (cm *Commands) Dispatch(raw string, inAtts []string, send func(text string, embed *Embed, attachments []string) error) bool {
	text := strings.TrimSpace(raw)
	if !strings.HasPrefix(text, cm.prefix) {
		return false
	}
	fields := strings.Fields(text[len(cm.prefix):])
	if len(fields) == 0 {
		return false
	}
	name := strings.ToLower(fields[0])
	cm.mu.Lock()
	c, ok := cm.handlers[name]
	cm.mu.Unlock()
	ctx := &Context{Name: name, Args: fields[1:], Attachments: inAtts, Raw: text, prefix: cm.prefix, send: send}
	if !ok {
		if cm.notFound != nil {
			cm.notFound(ctx)
		}
		return false
	}
	c.h(ctx)
	return true
}
