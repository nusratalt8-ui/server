package fileexplorer

import (
	"encoding/json"
	"time"

	"agentmanager/modules/panelapp"
	"agentmanager/modules/wsutil"
)

type Service struct {
	app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
	s := &Service{app: app}
	app.Panel.RegisterHandler("fs_open", app.Guarded(s.onFsOpen))
	app.Panel.RegisterHandler("fs_close", app.Guarded(s.onFsClose))
	app.Panel.RegisterHandler("fs_run", app.Guarded(s.onFsRun))
	app.Panel.RegisterHandler("fs_op", app.Guarded(s.onFsOp))
	app.Panel.RegisterHandler("fs_download", app.Guarded(s.onFsDownload))
	app.Panel.RegisterHandler("fs_upload", app.Guarded(s.onFsUpload))
	app.Panel.RegisterHandler("fs_download_multi", app.Guarded(s.onFsDownloadMulti))
	app.Panel.RegisterHandler("fs_search", app.Guarded(s.onFsSearch))
	app.Agent.RegisterHandler("fs_changed", wsutil.NewDebouncer(400*time.Millisecond).Wrap(s.onFsChanged))
	app.Agent.RegisterHandler("fs_search_result", s.onFsSearchResult)
	app.Agent.RegisterHandler("fs_download_result", s.onFsDownloadResult)
	app.Agent.RegisterHandler("fs_list_result", s.onFsListResult)
	app.Agent.RegisterHandler("fs_read_result", s.onFsReadResult)
	app.Agent.RegisterHandler("fs_op_result", s.onFsOpResult)
	return s
}

func (s *Service) onFsOpen(panelConnID string, payload json.RawMessage) {
	s.app.AppOpen(panelConnID, payload, "fs", "fs_open", "fs_refresh")
}

func (s *Service) onFsClose(panelConnID string, payload json.RawMessage) {
	s.app.AppClose(panelConnID, payload, "fs", "fs_close")
}

func (s *Service) onFsRun(_ string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
		Path    string `json:"path"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" || p.Path == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_run", map[string]string{"path": p.Path})
	}
}

func (s *Service) onFsOp(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Op      string          `json:"op"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" || p.Op == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_"+p.Op, p.Data)
	}
}

func (s *Service) onFsDownload(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_download", p.Data)
	}
}

func (s *Service) onFsUpload(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID      string   `json:"agent_id"`
		AttachmentID string   `json:"attachment_id"`
		DestPaths    []string `json:"dest_paths"`
		Overwrite    bool     `json:"overwrite"`
		Hidden       bool     `json:"hidden"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_upload", map[string]interface{}{
			"attachment_id": p.AttachmentID,
			"dest_paths":    p.DestPaths,
			"overwrite":     p.Overwrite,
			"hidden":        p.Hidden,
		})
	}
}

func (s *Service) onFsDownloadMulti(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string          `json:"agent_id"`
		Data    json.RawMessage `json:"data"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_download_multi", p.Data)
	}
}

func (s *Service) onFsSearch(panelConnID string, payload json.RawMessage) {
	var p struct {
		AgentID string `json:"agent_id"`
	}
	if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
		return
	}
	for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
		s.app.Agent.EmitTo(conn, "fs_search", payload)
	}
}

func (s *Service) onFsChanged(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_changed", p)
}

func (s *Service) onFsSearchResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_search_result", p)
}

func (s *Service) onFsDownloadResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_download_result", p)
}

func (s *Service) onFsListResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_list_result", p)
}

func (s *Service) onFsReadResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_read_result", p)
}

func (s *Service) onFsOpResult(agentConnID string, payload json.RawMessage) {
	id, ok := s.app.Reg.AgentFor(agentConnID)
	if !ok {
		return
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	p["agent_id"] = id
	s.app.Publish("fs:"+id, "fs_op_result", p)
}
