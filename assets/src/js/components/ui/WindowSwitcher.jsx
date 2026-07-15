export default function WindowSwitcher({ apps, openIds, index }) {
  if (!openIds.length) return null;
  const openApps = apps.filter((a) => openIds.includes(a.id));
  return (
    <div className="fixed inset-0 z-[9999] flex items-center justify-center pointer-events-none">
      <div className="vista-glass" style={{ padding: "16px 20px", display: "flex", flexDirection: "column", alignItems: "center", gap: 12 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          {openApps.map((app, i) => (
            <div key={app.id} className={"vista-glass-item" + (i === index ? " active" : "")}>
              <app.Icon style={{ width: 40, height: 40 }} />
              <span className="vista-glass-label">{app.title}</span>
            </div>
          ))}
        </div>
        <span className="vista-glass-label" style={{ fontSize: 12, fontWeight: 700, opacity: 0.9 }}>
          {openApps[index]?.title}
        </span>
      </div>
    </div>
  );
}