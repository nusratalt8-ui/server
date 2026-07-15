package attachments

import "agentmanager/modules/router"

func (s *Service) Prefix() string { return "/attachments" }

func (s *Service) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "POST", Path: "", Handler: s.Upload, Auth: true},
		{Method: "GET", Path: "/:id", Handler: s.Download, Auth: true},
	}
}
