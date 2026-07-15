package auth

import "time"

type Config struct {
	CookieName  string
	SessionTTL  time.Duration
	Secure      bool
	MaxName     int
	MinPassword int
}

func DefaultConfig() Config {
	return Config{
		CookieName:  "chocolatechipcookie",
		SessionTTL:  7 * 24 * time.Hour,
		Secure:      false,
		MaxName:     24,
		MinPassword: 6,
	}
}
