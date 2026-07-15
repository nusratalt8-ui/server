package auth

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "POST", Path: "/register", Handler: s.Register},
		{Method: "POST", Path: "/login", Handler: s.Login},
		{Method: "POST", Path: "/logout", Handler: s.Logout, Auth: true},
		{Method: "GET", Path: "/me", Handler: s.Me, Auth: true},
		{Method: "GET", Path: "/ice-servers", Handler: s.ICEServers, Auth: true},
		{Method: "GET", Path: "/uiprefs", Handler: s.GetUIPrefs, Auth: true},
		{Method: "PUT", Path: "/uiprefs", Handler: s.SetUIPrefs, Auth: true},
	}
}
