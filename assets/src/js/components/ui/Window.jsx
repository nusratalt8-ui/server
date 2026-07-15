import { useState, useRef, forwardRef, useImperativeHandle } from "react";
import useDraggable from "../../hooks/useDraggable";
import useResizable from "../../hooks/useResizable";

let topZ = 100;

const Window = forwardRef(function Window({ title, children, className = "", bodyClassName = "p-5", draggable = false, resizable = false, pos: initialPos, size: initialSize, onClose, onLayoutChange, style, tabs, activeTab, onTabChange }, ref) {
  const layoutRef = useRef(onLayoutChange);
  layoutRef.current = onLayoutChange;
  const { pos, onPointerDown } = useDraggable(initialPos, (p) => layoutRef.current?.({ pos: p }));
  const { size, onResizeDown } = useResizable(initialSize, undefined, (s) => layoutRef.current?.({ size: s }));
  const [z, setZ] = useState(() => ++topZ);
  const floating = draggable || resizable;

  const bringToFront = () => {
    if (z < topZ) setZ(++topZ);
  };

  useImperativeHandle(ref, () => ({ bringToFront }), [z]);

  const floatStyle = floating
    ? {
        position: "fixed",
        left: pos.x,
        top: pos.y,
        zIndex: z,
        ...(resizable
          ? { width: size.w, height: size.h, maxWidth: "95vw", maxHeight: "90vh", display: "flex", flexDirection: "column" }
          : {}),
      }
    : {};

  return (
    <div
      className={"y2k-window " + className}
      style={{ ...floatStyle, ...style }}
      onPointerDownCapture={floating ? bringToFront : undefined}
    >
      <div
        className="y2k-titlebar flex items-center gap-2 flex-shrink-0"
        onPointerDown={draggable ? onPointerDown : undefined}
        style={draggable ? { cursor: "move", userSelect: "none", touchAction: "none" } : undefined}
      >
        <span className="truncate" style={{ fontSize: 12, fontWeight: 600, opacity: 0.8 }}>{title}</span>
        {tabs && tabs.length > 0 && (
          <div className="flex items-center gap-0.5 flex-1 ml-2 overflow-x-auto" style={{ scrollbarWidth: "none" }}>
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onPointerDown={(e) => e.stopPropagation()}
                onClick={() => onTabChange?.(tab.id)}
                style={{
                  padding: "2px 10px",
                  fontSize: 11,
                  fontWeight: 500,
                  borderRadius: "4px 4px 0 0",
                  border: "none",
                  background: activeTab === tab.id ? "var(--window-face)" : "transparent",
                  color: activeTab === tab.id ? "var(--text)" : "var(--muted)",
                  cursor: "pointer",
                  whiteSpace: "nowrap",
                  borderBottom: activeTab === tab.id ? "2px solid var(--accent)" : "2px solid transparent",
                  transition: "color 0.1s, border-color 0.1s",
                }}
              >
                {tab.label}
              </button>
            ))}
          </div>
        )}
        {!tabs && <span className="flex-1" />}
        {onClose && (
          <button
            onPointerDown={(e) => e.stopPropagation()}
            onClick={onClose}
            style={{
              width: 22,
              height: 22,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              borderRadius: 4,
              border: "none",
              background: "transparent",
              color: "var(--muted)",
              cursor: "pointer",
              fontSize: 14,
              lineHeight: 1,
              transition: "background 0.1s, color 0.1s",
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = "rgba(239,68,68,0.2)"; e.currentTarget.style.color = "#ef4444"; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = "transparent"; e.currentTarget.style.color = "var(--muted)"; }}
          >
            ✕
          </button>
        )}
      </div>
      <div className={(resizable ? "flex-1 min-h-0 overflow-hidden " : "") + bodyClassName}>{children}</div>
      {resizable && (
        <div
          onPointerDown={onResizeDown}
          className="absolute bottom-0 right-0 w-4 h-4 flex items-end justify-end p-0.5"
          style={{ cursor: "nwse-resize", touchAction: "none" }}
        >
          <svg width="10" height="10" viewBox="0 0 10 10">
            <path d="M9 1L1 9M9 5L5 9M9 9h0" stroke="var(--edge-dark)" strokeWidth="1.5" strokeLinecap="round" fill="none" />
          </svg>
        </div>
      )}
    </div>
  );
});

export default Window;