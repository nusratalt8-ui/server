package ratelimit

import (
	"os"
	"strconv"
)

type Config struct {
	PerSecond float64
	Burst     int
}

func PanelConfig() Config {
	return Config{PerSecond: floatEnv("PANEL_RATE", 20), Burst: intEnv("PANEL_BURST", 40)}
}

func AgentConfig() Config {
	return Config{PerSecond: floatEnv("AGENT_RATE", 10), Burst: intEnv("AGENT_BURST", 20)}
}

func LoginConfig() Config {
	return Config{PerSecond: 0.5, Burst: 5}
}

func floatEnv(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
