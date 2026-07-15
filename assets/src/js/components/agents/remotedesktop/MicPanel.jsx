import { useState } from "react";
import { LED_ON, LED_OFF, LED_GLOW, MT, MM, MHI, MONO } from "./constants";

export default function MicPanel({ devices, device, setDevice, micOn, startMic, stopMic, micMuted, toggleMute, onRequestDevices, requested, inline }) {
  const [expanded, setExpanded] = useState(false);

  if (inline) {
    return (
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        {/* Mic toggle LED button */}
        <button
          onClick={() => {
            if (!requested) { onRequestDevices(); setExpanded(true); return; }
            if (micOn) stopMic(); else if (devices.length > 0) startMic(device);
          }}
          style={{
            display: "flex", alignItems: "center", gap: 6,
            background: micOn ? "rgba(192,57,43,0.15)" : "transparent",
            border: `1px solid ${micOn ? "rgba(192,57,43,0.4)" : MHI}`,
            borderRadius: 4, cursor: "pointer", padding: "4px 10px",
            transition: "all 0.15s",
          }}
        >
          <span style={{
            width: 7, height: 7, borderRadius: "50%",
            background: micOn ? LED_ON : LED_OFF,
            boxShadow: micOn ? `0 0 6px ${LED_GLOW}` : "none",
            flexShrink: 0, transition: "all 0.15s",
          }} />
          <span style={{ fontFamily: MONO, fontSize: 11, color: micOn ? MT : MM }}>
            {!requested ? "🎙 Mic" : micOn ? "Mic ON" : "Mic OFF"}
          </span>
        </button>

        {/* Device picker — shown when requested and devices exist */}
        {requested && devices.length > 0 && (
          <select
            value={device}
            onChange={(e) => {
              const id = parseInt(e.target.value);
              setDevice(id);
              if (micOn) startMic(id);
            }}
            style={{
              fontSize: 11, fontFamily: MONO,
              background: "var(--rd-key-bg)", color: MT,
              border: `1px solid ${MHI}`,
              padding: "3px 6px", borderRadius: 3, cursor: "pointer",
              maxWidth: 160,
            }}
          >
            {devices.map((d) => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </select>
        )}

        {/* Mute toggle — only when streaming */}
        {micOn && (
          <button
            onClick={toggleMute}
            style={{
              fontFamily: MONO, fontSize: 11, padding: "4px 10px",
              background: micMuted ? "rgba(0,0,0,0.3)" : "rgba(39,174,96,0.15)",
              border: `1px solid ${micMuted ? MHI : "rgba(39,174,96,0.4)"}`,
              borderRadius: 4, cursor: "pointer", color: micMuted ? MM : "#27ae60",
              transition: "all 0.15s",
            }}
          >
            {micMuted ? "Unmute" : "Live"}
          </button>
        )}

        {/* Loading state */}
        {requested && devices.length === 0 && (
          <span style={{ fontFamily: MONO, fontSize: 10, color: MM }}>No mics found</span>
        )}
      </div>
    );
  }

  // Non-inline fallback (unused now but kept)
  return null;
}