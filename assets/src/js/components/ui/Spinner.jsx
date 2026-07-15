import React from "react";

export default function Spinner({ size = 16, color = "rgba(255,255,255,0.3)", borderColor = "#fff", speed = 0.6 }) {
  return (
    <span style={{
      display: "inline-block",
      width: size,
      height: size,
      border: `2px solid ${color}`,
      borderTopColor: borderColor,
      borderRadius: "50%",
      animation: `spin ${speed}s linear infinite`,
      verticalAlign: "middle",
    }} />
  );
}