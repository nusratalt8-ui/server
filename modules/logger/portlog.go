package logger

import (
	"fmt"
	"os"
	"sync"
)

var (
	portMu      sync.RWMutex
	portLabels  = map[string]string{}
	portDefault string
)

func RegisterPort(id, label string) {
	portMu.Lock()
	defer portMu.Unlock()
	portLabels[id] = label
	if portDefault == "" {
		portDefault = id
	}
}

func PortLog(portID string, level Level, tag, text string) {
	mu.Lock()
	defer mu.Unlock()
	if level < minLevel {
		return
	}
	os.Stdout.WriteString(fmt.Sprintf("[%s] %s %s %s\n", portID, stamp(), tag, text))
}
