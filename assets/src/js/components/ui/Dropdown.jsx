import { useState, useRef, useEffect } from "react";

export default function Dropdown({ trigger, children, align = "left" }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e) => { if (ref.current && !ref.current.contains(e.target)) setOpen(false); };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  return (
    <div ref={ref} style={{ position: "relative", display: "inline-block" }}>
      <div onClick={() => setOpen(o => !o)}>{trigger}</div>
      {open && (
        <div style={{
          position: "absolute",
          top: "calc(100% + 4px)",
          [align === "right" ? "right" : "left"]: 0,
          zIndex: 200,
          minWidth: 160,
          background: "var(--window-face)",
          border: "1px solid var(--edge-dark)",
          borderRadius: 4,
          boxShadow: "0 4px 16px rgba(0,0,0,0.35)",
          padding: "3px 0",
          overflow: "hidden",
        }}>
          {children}
        </div>
      )}
    </div>
  );
}

export function DropdownItem({ children, onClick, danger, disabled, checked }) {
  return (
    <button
      disabled={disabled}
      onClick={() => onClick?.()}
      style={{
        display: "flex", alignItems: "center", gap: 6,
        width: "100%", textAlign: "left",
        padding: "6px 12px", fontSize: 12, fontWeight: 500,
        background: "transparent", border: "none", cursor: disabled ? "default" : "pointer",
        color: disabled ? "var(--muted)" : danger ? "#e87070" : "var(--text)",
        opacity: disabled ? 0.5 : 1,
        transition: "background 0.1s",
      }}
      onMouseEnter={e => { if (!disabled) e.currentTarget.style.background = "var(--hover)"; }}
      onMouseLeave={e => { e.currentTarget.style.background = "transparent"; }}
    >
      <span style={{ width: 12, flexShrink: 0, opacity: checked ? 1 : 0, fontSize: 10 }}>✓</span>
      {children}
    </button>
  );
}

export function DropdownSep() {
  return <div style={{ height: 1, background: "var(--edge-dark)", margin: "3px 0" }} />;
}