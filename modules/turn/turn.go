package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"agentmanager/modules/config"
)

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

func ICEServers() []ICEServer {
	host := config.PublicHost()
	if host == "" || config.TURNSecret == "" {
		return nil
	}
	exp := time.Now().Unix() + 3600
	username := fmt.Sprintf("%d", exp)
	mac := hmac.New(sha1.New, []byte(config.TURNSecret))
	mac.Write([]byte(username))
	credential := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return []ICEServer{
		{
			URLs: []string{
				fmt.Sprintf("turn:%s:3478?transport=udp", host),
				fmt.Sprintf("turn:%s:3478?transport=tcp", host),
				fmt.Sprintf("turns:%s:5349", host),
			},
			Username:   username,
			Credential: credential,
		},
	}
}
