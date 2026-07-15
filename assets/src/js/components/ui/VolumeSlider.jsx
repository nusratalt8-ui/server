import { useRef, useState } from "react";

export default function VolumeSlider({ value, onChange }) {
  const [local, setLocal] = useState(value);
  const dragging = useRef(false);
  const display = dragging.current ? local : value;

  return (
    <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="var(--rd-metal-muted)" strokeWidth="2">
        {display === 0 ? (
          <>
            <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
            <line x1="23" y1="9" x2="17" y2="15" />
            <line x1="17" y1="9" x2="23" y2="15" />
          </>
        ) : (
          <>
            <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
            {display > 30 && <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />}
            {display > 60 && <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />}
          </>
        )}
      </svg>
      <input
        type="range" min={0} max={100} value={display}
        onMouseDown={() => { dragging.current = true; setLocal(value); }}
        onChange={(e) => setLocal(parseInt(e.target.value))}
        onMouseUp={(e) => { dragging.current = false; onChange(parseInt(e.target.value)); }}
        style={{
          width: 90, cursor: "pointer", height: 3,
          accentColor: "var(--rd-seg-fill)",
          background: `linear-gradient(to right, var(--rd-seg-fill) ${display}%, var(--rd-seg-empty) ${display}%)`,
        }}
      />
      <span style={{ fontFamily: '"Consolas", "Courier New", monospace', fontSize: 11, color: "var(--rd-metal-muted)", minWidth: 26, textAlign: "right" }}>
        {display}
      </span>
    </div>
  );
}