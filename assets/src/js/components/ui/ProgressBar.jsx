export default function ProgressBar({ progress, label, error }) {
  const pct = Math.min(Math.round(progress || 0), 100);
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 4, width: "100%" }}>
      {label && (
        <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, color: "var(--muted)" }}>
          <span>{label}</span>
          {!error && <span>{pct}%</span>}
        </div>
      )}
      <div style={{
        height: 8,
        background: "var(--input-bg)",
        borderRadius: 4,
        overflow: "hidden",
        border: "1px solid var(--edge-dark)",
      }}>
        <div style={{
          width: `${error ? 100 : pct}%`,
          height: "100%",
          borderRadius: 3,
          background: error ? "#c0392b" : "var(--titlebar)",
          transition: "width 0.2s ease",
        }} />
      </div>
      {error && (
        <div style={{ fontSize: 11, color: "#c0392b" }}>{error}</div>
      )}
    </div>
  );
}