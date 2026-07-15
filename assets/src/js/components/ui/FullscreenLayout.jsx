import { useState, useEffect } from "react";

export default function FullscreenLayout({ title, tabs, onClose, defaultTab, children }) {
  const [activeTab, setActiveTab] = useState(defaultTab || tabs[0]?.id);

  useEffect(() => {
    const onKey = (e) => { if (e.key === "Escape") onClose(); };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex" style={{ background: "var(--bg)" }}>
      <div className="flex flex-row w-full h-full">
        {/* Sidebar */}
        <div className="flex flex-col w-56 shrink-0" style={{ background: "var(--sidebar-bg)", borderRight: "2px solid", borderColor: "var(--edge-dark)" }}>
          <div className="flex items-center justify-between px-4 py-3" style={{ borderBottom: "2px solid", borderColor: "var(--edge-dark)" }}>
            <span style={{ fontSize: 10, fontWeight: 700, color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.08em" }}>{title}</span>
            <button onClick={onClose} style={{ color: "var(--muted)", background: "none", border: "none", cursor: "pointer", fontSize: 14, padding: 4 }}>✕</button>
          </div>
          <div className="flex flex-col flex-1 p-2 gap-0.5 overflow-y-auto">
            {tabs.map((tab, idx) => {
              const prevGroup = idx > 0 ? tabs[idx - 1].group : undefined;
              const showGroupHeader = tab.group && tab.group !== prevGroup;
              return (
                <div key={tab.id}>
                  {showGroupHeader && (
                    <div className="flex items-center gap-2 px-3 pt-3 pb-1.5">
                      <span style={{ fontSize: 9, fontWeight: 700, color: "var(--muted)", textTransform: "uppercase", letterSpacing: "0.08em" }}>{tab.groupLabel || tab.group}</span>
                      <div className="flex-1 h-px" style={{ background: "var(--border)" }} />
                    </div>
                  )}
                  <button
                    onClick={() => setActiveTab(tab.id)}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
                    style={{
                      borderRadius: 4,
                      fontSize: 12,
                      fontWeight: activeTab === tab.id ? 600 : 400,
                      background: activeTab === tab.id ? "var(--titlebar)" : "transparent",
                      color: activeTab === tab.id ? "var(--titlebar-text)" : "var(--muted)",
                    }}
                  >
                    {tab.icon && <span>{tab.icon}</span>}
                    {tab.label}
                  </button>
                </div>
              );
            })}
          </div>
          <div className="p-2" style={{ borderTop: "2px solid", borderColor: "var(--edge-dark)" }}>
            <button onClick={onClose} style={{ width: "100%", display: "flex", alignItems: "center", gap: 8, padding: "6px 10px", fontSize: 11, color: "var(--muted)", background: "none", border: "none", cursor: "pointer", textAlign: "left" }}>
              ← Back
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 flex flex-col min-w-0 min-h-0">
          <div className="flex-1 overflow-y-auto p-6">
            {children(activeTab)}
          </div>
        </div>
      </div>
    </div>
  );
}