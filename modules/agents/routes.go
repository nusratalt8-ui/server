package agents

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/agents" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "GET", Path: "", Handler: s.ListAgents, Auth: true},
		{Method: "GET", Path: "/all", Handler: s.ListAllAgents, Auth: true},
		{Method: "GET", Path: "/user/:uid", Handler: s.ListUserAgents, Auth: true},
		{Method: "GET", Path: "/:id", Handler: s.GetAgent, Auth: true},
	}
}
