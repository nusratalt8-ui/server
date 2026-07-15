export default function Tabs({ tabs, active, onChange, actions }) {
  return (
    <div style={{ display: "flex", alignItems: "center", padding: "4px 8px 0", background: "var(--window-face)", borderBottom: "1px solid var(--edge-dark)", flexShrink: 0 }}>
      {tabs.map(t => (
        <button key={t} onClick={() => onChange(t)} style={{
          padding: "3px 12px", fontSize: 11, fontWeight: 600, cursor: "pointer",
          background: active === t ? "var(--input-bg)" : "transparent",
          color: active === t ? "var(--text)" : "var(--muted)",
          border: "1px solid " + (active === t ? "var(--edge-dark)" : "transparent"),
          borderBottom: active === t ? "1px solid var(--input-bg)" : "none",
          borderRadius: "3px 3px 0 0", marginBottom: -1,
        }}>{t}</button>
      ))}
      {actions && <><div style={{ flex: 1 }} />{actions}</>}
    </div>
  );
}