package agentkey

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/agent-key" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "GET", Path: "", Handler: s.GetInfo, Auth: true},
		{Method: "POST", Path: "/rotate", Handler: s.RotateKey, Auth: true},
	}
}
