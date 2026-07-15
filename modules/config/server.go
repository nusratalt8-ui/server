package config

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const APIPrefix = "/api/v1"

var (
	publicIPOnce sync.Once
	publicIP     string
)

func PublicHost() string {
	publicIPOnce.Do(func() {
		cl := &http.Client{Timeout: 5 * time.Second}
		for _, svc := range []string{
			"https://api.ipify.org",
			"https://ifconfig.me/ip",
			"https://icanhazip.com",
		} {
			resp, err := cl.Get(svc)
			if err != nil {
				continue
			}
			b, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}
			if ip := strings.TrimSpace(string(b)); ip != "" {
				publicIP = ip
				return
			}
		}
	})
	return publicIP
}

func PanelHost() string { return "0.0.0.0" }
func AgentHost() string { return "0.0.0.0" }
func PanelPort() string { return "8080" }
func AgentPort() string { return "9090" }
func CertFile() string  { return "assets/certs/cert.pem" }
func KeyFile() string   { return "assets/certs/key.pem" }

func AllowedDomains() []string {
	host := PublicHost()
	if host == "" {
		return []string{}
	}
	return []string{
		"https://" + host + ":8080",
		"http://" + host + ":8080",
	}
}

const (
	MasterKey  = "b3f1a2e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2d3c4b5a6f7e8d9c0b1a2"
	TURNSecret = "9fc07c3d5969e67bc1d3764f76814779ea4463ccf4868582c1813449abf6a3f8"
)
