package plan

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/plans" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "PUT", Path: "/user/:id", Handler: s.SetUserPlan, Auth: true},
	}
}
