package wsutil

import "time"

type Config struct {
	WriteWait          time.Duration
	PongWait           time.Duration
	PingPeriod         time.Duration
	MaxMessageSize     int64
	MaxFrameSize       int64
	SendBuffer         int
	ReadBufferSize     int
	WriteBufferSize    int
	MsgPerSecond       float64
	MsgBurst           int
	MaxViolations      int
	OfflineGracePeriod time.Duration
	EnableProbe        bool
	InboundQueue       int
	InboundWorkers     int
}

func DefaultConfig() Config {
	return Config{
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         20 * time.Second,
		MaxMessageSize:     4 * 1024 * 1024,
		MaxFrameSize:       32 * 1024 * 1024,
		SendBuffer:         256,
		ReadBufferSize:     256 * 1024,
		WriteBufferSize:    256 * 1024,
		MsgPerSecond:       15,
		MsgBurst:           30,
		MaxViolations:      20,
		OfflineGracePeriod: 10 * time.Second,
		InboundQueue:       8192,
		InboundWorkers:     32,
	}
}
