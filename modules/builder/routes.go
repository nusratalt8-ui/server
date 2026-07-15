package builder

import "agentmanager/modules/router"

func (h *Handler) Prefix() string { return "/build" }

func (h *Handler) Routes() []router.RouteConfig {
	return []router.RouteConfig{
		{Method: "POST", Path: "/start", Handler: h.StartBuild, Auth: true},
		{Method: "GET", Path: "/download/:file", Handler: h.DownloadExe, Auth: true},
		{Method: "POST", Path: "/icon", Handler: h.UploadIcon, Auth: true},
	}
}
