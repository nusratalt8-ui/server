package webcam

import (
    "encoding/json"

    "agentmanager/modules/logger"
    "agentmanager/modules/panelapp"
)

type Service struct {
    app *panelapp.App
}

func NewService(app *panelapp.App) *Service {
    s := &Service{app: app}
    app.Panel.RegisterHandler("cam_start", app.Guarded(s.onCamStart))
    app.Panel.RegisterHandler("cam_stop", app.Guarded(s.onCamStop))
    app.Panel.RegisterHandler("cam_list", app.Guarded(s.onCamList))
    app.Agent.RegisterHandler("cam_devices", s.onCamDevices)
    app.RegisterBinaryMsg("cam_frame", s.onCamFrame)
    return s
}

func (s *Service) onCamFrame(agentConnID string, payload []byte) {
    id, ok := s.app.Reg.AgentFor(agentConnID)
    if !ok {
        return
    }
    tagged := make([]byte, 1+len(payload))
    tagged[0] = 0x03
    copy(tagged[1:], payload)
    for _, conn := range s.app.Sess.Subscribers("webcam:" + id) {
        s.app.Panel.EmitFrameRaw(conn, tagged)
    }
}
func (s *Service) onCamStart(panelConnID string, payload json.RawMessage) {
    var p struct {
        AgentID  string `json:"agent_id"`
        DeviceID int    `json:"device_id"`
    }
    if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
        return
    }
    s.app.Sess.Subscribe(panelConnID, "webcam:"+p.AgentID)
    for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
        s.app.Agent.EmitTo(conn, "cam_start", map[string]interface{}{
            "device_id": p.DeviceID,
        })
    }
}

func (s *Service) onCamStop(panelConnID string, payload json.RawMessage) {
    var p struct {
        AgentID string `json:"agent_id"`
    }
    if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
        return
    }
    s.app.Sess.Unsubscribe(panelConnID, "webcam:"+p.AgentID)
    for _, conn := range s.app.Reg.ConnsFor(p.AgentID) {
        s.app.Agent.EmitTo(conn, "cam_stop", nil)
    }
}

func (s *Service) onCamList(panelConnID string, payload json.RawMessage) {
    var p struct {
        AgentID string `json:"agent_id"`
    }
    if json.Unmarshal(payload, &p) != nil || p.AgentID == "" {
        return
    }
    conns := s.app.Reg.ConnsFor(p.AgentID)
    if len(conns) > 0 {
        s.app.Agent.EmitTo(conns[0], "cam_list", nil)
    }
}

func (s *Service) onCamDevices(agentConnID string, payload json.RawMessage) {
    id, ok := s.app.Reg.AgentFor(agentConnID)
    if !ok {
        return
    }
    var devices interface{}
    if json.Unmarshal(payload, &devices) != nil {
        logger.Infof("[cam] onCamDevices: bad payload agent=%s raw=%s", id, string(payload))
        return
    }
    s.app.Publish("webcam:"+id, "cam_devices", map[string]interface{}{"agent_id": id, "devices": devices})
}