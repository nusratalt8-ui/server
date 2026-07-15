//go:build windows

package apps

import (
	"encoding/json"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/transport"
)

// HandlerFunc handles an incoming message for an app.
type HandlerFunc func(client *transport.Client, msgType string, payload json.RawMessage)

// App defines a registered app's message handler and lifecycle.
type App struct {
	// Handle is called for every message type this app owns.
	Handle HandlerFunc
	// Stop is called on explicit stop message or WS disconnect. Optional.
	Stop func()
}

var (
	regMu    sync.Mutex
	registry = map[string]*App{}   // keyed by message prefix/name e.g. "liveview", "fs"
	msgMap   = map[string]string{} // maps exact msgType -> app name
)

// RegisterApp registers an app by name.
func RegisterApp(name string, app *App, messages ...string) {
	regMu.Lock()
	registry[name] = app
	for _, msg := range messages {
		msgMap[msg] = name
	}
	regMu.Unlock()
}

// Dispatch routes a message to the correct app handler.
// Returns true if handled.
func Dispatch(client *transport.Client, msgType string, payload json.RawMessage) bool {
	regMu.Lock()
	name, ok := msgMap[msgType]
	var app *App
	if ok {
		app = registry[name]
	}
	regMu.Unlock()
	if app == nil || app.Handle == nil {
		return false
	}
	app.Handle(client, msgType, payload)
	return true
}

// StopApp calls Stop on the named app.
func StopApp(name string) {
	regMu.Lock()
	app := registry[name]
	regMu.Unlock()
	if app != nil && app.Stop != nil {
		app.Stop()
	}
}

// StopAll calls Stop on every registered app.
func StopAll() {
	regMu.Lock()
	apps := make([]*App, 0, len(registry))
	for _, a := range registry {
		apps = append(apps, a)
	}
	regMu.Unlock()
	for _, a := range apps {
		if a.Stop != nil {
			a.Stop()
		}
	}
}
