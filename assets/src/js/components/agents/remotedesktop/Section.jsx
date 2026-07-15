import { useState } from "react";
import { MONO, MHI, MM } from "./constants";

export default function Section({ label, children, defaultOpen = true, onOpen }) {
  const [open, setOpen] = useState(defaultOpen);
  const toggle = () => {
    const next = !open;
    setOpen(next);
    if (next && onOpen) onOpen();
  };
  return (
    <div>
      <button
        onClick={toggle}
        style={{
          display: "flex", alignItems: "center", justifyContent: "space-between",
          width: "100%", background: "transparent", border: "none", cursor: "pointer",
          padding: "4px 0", borderBottom: `1px solid ${MHI}`, marginBottom: open ? 8 : 0,
        }}
      >
        <span style={{ fontFamily: MONO, fontSize: 10, color: MM, letterSpacing: "0.1em", textTransform: "uppercase", fontWeight: 600 }}>
          {label}
        </span>
        <span style={{ fontFamily: MONO, fontSize: 10, color: MM, lineHeight: 1 }}>{open ? "▾" : "▸"}</span>
      </button>
      {open && <div style={{ display: "flex", flexDirection: "column", gap: 2, marginBottom: 4 }}>{children}</div>}
    </div>
  );
}