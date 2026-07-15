import { MONO } from "./constants";

export default function StreamView({ canvasRef, wrapRef, stale, onMouseMove, onMouseDown, onMouseUp, onWheel }) {
  return (
    <div style={{ flex: 1, minWidth: 0, position: "relative", background: "#000", overflow: "hidden" }}>
      <div
        ref={wrapRef}
        style={{ position: "absolute", inset: 0, cursor: "crosshair" }}
        onMouseMove={onMouseMove}
        onMouseDown={onMouseDown}
        onMouseUp={onMouseUp}
        onWheel={onWheel}
        onContextMenu={(e) => e.preventDefault()}
      >
        <canvas
          ref={canvasRef}
          tabIndex={-1}
          style={{ position: "absolute", display: "block", left: 0, top: 0, width: "100%", height: "100%" }}
        />
      </div>
      {stale && (
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", background: "rgba(0,0,0,0.75)", zIndex: 20, pointerEvents: "none" }}>
          <span style={{ color: "#fff", fontFamily: MONO, fontSize: 13, letterSpacing: "0.06em" }}>Reconnecting...</span>
        </div>
      )}
    </div>
  );
}