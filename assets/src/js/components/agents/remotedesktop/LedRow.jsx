import { LED_ON, LED_GLOW, MT, MM, MHI, MLO } from "./constants";

export default function LedRow({ label, active, onClick }) {
  const kx = active ? 26 : 10;
  return (
    <button
      onClick={onClick}
      style={{
        display: "flex", alignItems: "center", gap: 10,
        background: "transparent", border: "none", cursor: "pointer",
        padding: "5px 6px", width: "100%", textAlign: "left",
        borderRadius: 4, transition: "background 0.15s",
      }}
      onMouseEnter={e => e.currentTarget.style.background = "var(--rd-metal-hi)"}
      onMouseLeave={e => e.currentTarget.style.background = "transparent"}
    >
      <span style={{ flex: 1, fontFamily: '"Consolas", "Courier New", monospace', fontSize: 12, color: MT }}>
        {label}
      </span>
      <svg width="40" height="22" viewBox="0 0 40 22" style={{ flexShrink: 0, overflow: "visible" }}>
        <defs>
          <linearGradient id="sw-track-on" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#1a3a1a" />
            <stop offset="100%" stopColor="#0d200d" />
          </linearGradient>
          <linearGradient id="sw-track-off" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#111411" />
            <stop offset="100%" stopColor="#1a1d1a" />
          </linearGradient>
          <linearGradient id="sw-knob" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#c8d4c8" />
            <stop offset="35%" stopColor="#9aaa9a" />
            <stop offset="100%" stopColor="#5a665a" />
          </linearGradient>
          <filter id="sw-glow">
            <feGaussianBlur stdDeviation="1.5" result="blur" />
            <feComposite in="SourceGraphic" in2="blur" operator="over" />
          </filter>
          <filter id="sw-shadow">
            <feDropShadow dx="0" dy="1" stdDeviation="1" floodColor="rgba(0,0,0,0.7)" />
          </filter>
        </defs>

        <rect x="1" y="5" width="38" height="13" rx="6.5"
          fill={active ? "url(#sw-track-on)" : "url(#sw-track-off)"}
          stroke={active ? "rgba(40,90,40,0.9)" : "rgba(0,0,0,0.8)"}
          strokeWidth="1"
        />
        <rect x="2" y="6" width="36" height="5" rx="4"
          fill="rgba(0,0,0,0.35)"
        />
        {active && (
          <>
            <rect x="1" y="5" width="38" height="13" rx="6.5"
              fill="none" stroke={LED_ON} strokeWidth="0.75" opacity="0.5"
            />
            <circle cx="10" cy="11.5" r="2.5" fill={LED_ON} opacity="0.85" filter="url(#sw-glow)" />
          </>
        )}

        <g filter="url(#sw-shadow)" style={{ transition: "transform 0.12s ease" }}
          transform={`translate(${kx}, 11.5)`}>
          <circle r="7.5" fill="url(#sw-knob)"
            stroke="rgba(0,0,0,0.5)" strokeWidth="0.75"
          />
          <ellipse cx="-1" cy="-2.5" rx="3.5" ry="2"
            fill="rgba(255,255,255,0.22)"
          />
          <line x1="-2.5" y1="-3" x2="-2.5" y2="3" stroke="rgba(60,70,60,0.7)" strokeWidth="0.8" />
          <line x1="0" y1="-3.5" x2="0" y2="3.5" stroke="rgba(60,70,60,0.7)" strokeWidth="0.8" />
          <line x1="2.5" y1="-3" x2="2.5" y2="3" stroke="rgba(60,70,60,0.7)" strokeWidth="0.8" />
          <circle cx="0" cy="0" r="1.8"
            fill={active ? LED_ON : "rgba(30,40,30,0.8)"}
            stroke={active ? "rgba(255,255,255,0.2)" : "rgba(0,0,0,0.4)"}
            strokeWidth="0.5"
          />
          {active && (
            <circle cx="0" cy="0" r="1.8" fill={LED_ON} opacity="0.6" filter="url(#sw-glow)" />
          )}
        </g>
      </svg>
    </button>
  );
}