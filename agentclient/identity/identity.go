package identity

import (
	"os"
	"strings"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

func file() string {
	base := config.DataPath()
	if base == "" {
		return ""
	}
	return base + `\id`
}

func Load() string {
	data, err := os.ReadFile(file())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func Save(id string) error {
	p := file()
	if p == "" {
		return nil
	}
	dir := config.DataPath()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	sysutil.HidePath(dir)
	if err := os.WriteFile(p, []byte(id), 0o600); err != nil {
		return err
	}
	sysutil.HidePath(p)
	return nil
}
