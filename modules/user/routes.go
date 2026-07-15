package user

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/users" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "GET", Path: "", Handler: s.ListUsers, Auth: true},
		{Method: "GET", Path: "/:id", Handler: s.GetProfile, Auth: true},
	}
}
