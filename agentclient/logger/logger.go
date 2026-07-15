package logger

import (
	"fmt"
	"sync"
	"time"
)

type Sender interface {
	Send(msgType string, payload interface{}) error
}

var (
	mu     sync.Mutex
	client Sender
	buf    []entry
)

type entry struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Time  int64  `json:"time"`
}

func Init(c Sender) {
	mu.Lock()
	client = c
	pending := buf
	buf = nil
	mu.Unlock()
	for _, e := range pending {
		c.Send("agent_log", e)
	}
}

func Info(format string, args ...interface{}) {
	emit("info", fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	emit("error", fmt.Sprintf(format, args...))
}

func emit(level, msg string) {
	e := entry{Level: level, Msg: msg, Time: time.Now().Unix()}
	mu.Lock()
	c := client
	if c == nil {
		if len(buf) < 200 {
			buf = append(buf, e)
		}
		mu.Unlock()
		return
	}
	mu.Unlock()
	c.Send("agent_log", e)
}
