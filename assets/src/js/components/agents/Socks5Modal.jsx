import { useState, useEffect } from "react";

export default function Socks5Modal() {
  const [info, setInfo] = useState(null);

  useEffect(() => {
    window.__socks5Modal = (payload) => setInfo(payload);
    return () => { window.__socks5Modal = null; };
  }, []);

  if (!info) return null;

  const host = info.host || "";
  const port = info.port || "";

  const lines = [
    `network.proxy.type: 1`,
    `network.proxy.socks: ${host}`,
    `network.proxy.socks_port: ${port}`,
    `network.proxy.socks_version: 5`,
    `network.proxy.socks_remote_dns: true`,
    `network.proxy.no_proxies_on: ""`,
  ];

  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)", zIndex: 9999, display: "flex", alignItems: "center", justifyContent: "center" }} onClick={() => setInfo(null)}>
      <div style={{ background: "#0f172a", border: "1px solid #334155", borderRadius: 8, padding: 24, maxWidth: 520, width: "90%" }} onClick={(e) => e.stopPropagation()}>
        <h3 style={{ margin: "0 0 16px", fontSize: 16, color: "#e2e8f0" }}>SOCKS5 Proxy Active</h3>
        <div style={{ background: "#1e293b", borderRadius: 6, padding: "12px 16px", marginBottom: 12, fontFamily: "monospace", fontSize: 13, color: "#4ade80", wordBreak: "break-all" }}>
          socks5://{host}:{port}
        </div>

        <h4 style={{ margin: "0 0 8px", fontSize: 13, color: "#94a3b8" }}>Firefox about:config</h4>
        <div style={{ background: "#1e293b", borderRadius: 6, padding: 12, fontFamily: "monospace", fontSize: 11, color: "#e2e8f0", lineHeight: 1.8 }}>
          {lines.map((line, i) => (
            <div key={i} style={{ whiteSpace: "nowrap" }}>{line}</div>
          ))}
        </div>

        <div style={{ fontSize: 11, color: "#64748b", marginTop: 8, lineHeight: 1.4 }}>
          Go to <b>about:config</b>, search each pref, click + to add string/integer/boolean.
        </div>

        <button onClick={() => setInfo(null)} style={{ marginTop: 16, background: "#334155", border: "none", borderRadius: 4, color: "#e2e8f0", padding: "6px 14px", cursor: "pointer", fontSize: 12, float: "right" }}>Close</button>
      </div>
    </div>
  );
}