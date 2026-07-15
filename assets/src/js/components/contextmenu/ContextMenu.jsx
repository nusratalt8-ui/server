import { createContext, useContext, useState, useCallback, useRef, useEffect } from "react";

const Ctx = createContext(null);
const EDGE_GAP = 8;

function SubMenu({ items, onClose }) {
  return (
    <div className="win-swish" style={{
      minWidth: 180,
      background: "color-mix(in srgb, var(--window-face) 92%, transparent)",
      backdropFilter: "blur(12px)", border: "1px solid var(--border)",
      borderRadius: 8, boxShadow: "0 8px 32px rgba(0,0,0,0.6)", padding: "4px",
    }}>
      {items.map((it, i) => it === "—" ? (
        <div key={i} style={{ margin: "3px 8px", borderTop: "1px solid var(--border)" }} />
      ) : (
        <button key={i} onClick={() => { it.action?.(); onClose(); }} style={{
          display: "block", width: "100%", textAlign: "left", padding: "6px 12px",
          fontSize: 12.5, borderRadius: 5, border: "none", background: "transparent",
          color: it.danger ? "#f87171" : "var(--text)", cursor: "pointer", transition: "background 0.08s",
        }}
          onMouseEnter={(e) => e.currentTarget.style.background = it.danger ? "rgba(248,113,113,0.12)" : "color-mix(in srgb, var(--text) 8%, transparent)"}
          onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}>
          {it.label}
        </button>
      ))}
    </div>
  );
}

function MenuItem({ it, onClose }) {
  const [hovered, setHovered] = useState(false);
  const [showSub, setShowSub] = useState(false);
  const ref = useRef(null);
  const hasSub = Array.isArray(it.submenu) && it.submenu.length > 0;
  const leaveTimer = useRef(null);

  if (it === "—") return <div style={{ margin: "3px 8px", borderTop: "1px solid var(--border)" }} />;

  const onEnter = () => {
    setHovered(true);
    if (leaveTimer.current) { clearTimeout(leaveTimer.current); leaveTimer.current = null; }
    if (hasSub) setShowSub(true);
  };

  const onLeave = () => {
    setHovered(false);
    if (hasSub) {
      leaveTimer.current = setTimeout(() => setShowSub(false), 200);
    }
  };

  const onSubEnter = () => {
    if (leaveTimer.current) { clearTimeout(leaveTimer.current); leaveTimer.current = null; }
  };

  const onSubLeave = () => {
    leaveTimer.current = setTimeout(() => setShowSub(false), 200);
  };

  return (
    <div style={{ position: "relative" }} ref={ref}
      onMouseEnter={onEnter}
      onMouseLeave={onLeave}>
      <button onClick={() => { if (!hasSub) { it.action?.(); onClose(); } }} style={{
        display: "flex", width: "100%", textAlign: "left", padding: "6px 12px",
        fontSize: 12.5, borderRadius: 5, border: "none", background: hovered ? (it.danger ? "rgba(248,113,113,0.12)" : "color-mix(in srgb, var(--text) 8%, transparent)") : "transparent",
        color: it.danger ? "#f87171" : "var(--text)", cursor: "pointer", transition: "background 0.08s", alignItems: "center", justifyContent: "space-between",
      }}>
        <span>{it.label}</span>
        {hasSub && <span className="opacity-40 ml-3">▸</span>}
      </button>
      {showSub && hasSub && (
        <div style={{ position: "absolute", left: "100%", top: 0, zIndex: 10000, padding: "4px 8px 4px 0" }}
             onMouseEnter={onSubEnter} onMouseLeave={onSubLeave}>
          <SubMenu items={it.submenu} onClose={onClose} />
        </div>
      )}
    </div>
  );
}

function Menu({ x, y, items, onClose }) {
  const ref = useRef(null);
  const [pos, setPos] = useState({ x, y });
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    let nx = x, ny = y;
    if (x + rect.width > window.innerWidth - EDGE_GAP) nx = Math.max(EDGE_GAP, x - rect.width);
    if (y + rect.height > window.innerHeight - EDGE_GAP) ny = Math.max(EDGE_GAP, y - rect.height);
    setPos({ x: nx, y: ny });
  }, [x, y]);
  useEffect(() => {
    const onDown = (e) => { if (!ref.current || !ref.current.contains(e.target)) onClose(); };
    window.addEventListener("pointerdown", onDown, true);
    return () => window.removeEventListener("pointerdown", onDown, true);
  }, [onClose]);
  return (
    <div ref={ref} className="win-swish fixed z-[9999]" style={{
      left: pos.x, top: pos.y, minWidth: 180,
      background: "color-mix(in srgb, var(--window-face) 92%, transparent)",
      backdropFilter: "blur(12px)", border: "1px solid var(--border)",
      borderRadius: 8, boxShadow: "0 8px 32px rgba(0,0,0,0.6)", padding: "4px",
    }} onContextMenu={(e) => e.preventDefault()}>
      {items.map((it, i) => <MenuItem key={i} it={it} onClose={onClose} />)}
    </div>
  );
}

export function ContextMenuProvider({ children }) {
  const [menu, setMenu] = useState(null);
  const open = useCallback((e, items) => { e.preventDefault(); setMenu({ x: e.clientX, y: e.clientY, items }); }, []);
  const close = useCallback(() => setMenu(null), []);
  return (
    <Ctx.Provider value={open}>
      {children}
      {menu && <Menu {...menu} onClose={close} />}
    </Ctx.Provider>
  );
}

export function useContextMenu() { return useContext(Ctx); }