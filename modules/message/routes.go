package message

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/messages" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "GET", Path: "/:agent", Handler: s.History, Auth: true},
	}
}
