//go:build windows

package loader

import (
	"io/fs"
	"sync"

	"github.com/microsoft/UpdateAssistant/modules/logger"
)

type Loader struct {
	src fs.FS
}

func New(src fs.FS) *Loader { return &Loader{src: src} }

func (l *Loader) Load(name string) (*Plugin, error) {
	data, err := loadBytes(l.src, name)
	if err != nil {
		logger.Error("load %s: %v", name, err)
		return nil, err
	}
	p, err := Open(name, data)
	if err != nil {
		logger.Error("open %s: %v", name, err)
		return nil, err
	}
	logger.Info("loaded %s", name)
	return p, nil
}

func (l *Loader) Unload(p *Plugin) {
	if p == nil {
		return
	}
	logger.Info("unloaded %s", p.Name)
	p.Close()
}

var (
	defMu sync.RWMutex
	def   *Loader
)

func SetFS(src fs.FS) {
	defMu.Lock()
	def = New(src)
	defMu.Unlock()
}

func Load(name string) (*Plugin, error) {
	defMu.RLock()
	l := def
	defMu.RUnlock()
	if l == nil {
		return nil, ErrLoadFailed
	}
	return l.Load(name)
}

func Unload(p *Plugin) {
	defMu.RLock()
	l := def
	defMu.RUnlock()
	if l == nil {
		if p != nil {
			p.Close()
		}
		return
	}
	l.Unload(p)
}
