import { useState } from "react";
import { MLO, KEY_BG, KEY_HI, KEY_LO, KEY_T, MONO } from "./constants";

export default function Key({ label, danger, onClick }) {
  const [pressed, setPressed] = useState(false);
  return (
    <button
      onMouseDown={() => setPressed(true)}
      onMouseUp={() => { setPressed(false); onClick(); }}
      onMouseLeave={() => setPressed(false)}
      style={{
        background: pressed ? MLO : KEY_BG,
        border: "none",
        borderRadius: 4,
        boxShadow: pressed
          ? `inset 0 2px 5px ${MLO}, 0 1px 0 ${KEY_HI}`
          : `0 4px 0 ${KEY_LO}, 0 1px 0 ${KEY_HI}, inset 0 1px 0 rgba(255,255,255,0.07)`,
        padding: "7px 14px",
        cursor: "pointer",
        flex: 1,
        transform: pressed ? "translateY(3px)" : "translateY(0)",
        userSelect: "none",
      }}
    >
      <span style={{
        fontSize: 12, fontFamily: MONO, fontWeight: 700,
        color: danger ? "#e87070" : KEY_T,
        letterSpacing: "0.04em",
        display: "block", textAlign: "center",
      }}>
        {label}
      </span>
    </button>
  );
}