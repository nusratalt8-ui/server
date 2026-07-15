import Window from "./Window";
import Button from "./Button";

export default function InfoModal({ title, children, onClose }) {
  return (
    <div onClick={onClose} style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,0.4)", zIndex: 9999, display: "flex", alignItems: "center", justifyContent: "center" }}>
      <div onClick={(e) => e.stopPropagation()}>
        <Window title={title || "Info"} onClose={onClose} style={{ width: 420, maxHeight: "70vh" }}>
        <div style={{ fontSize: 11, lineHeight: 1.6, color: "var(--text)", overflowY: "auto", maxHeight: "calc(70vh - 60px)" }}>
          {children}
        </div>
        <div style={{ marginTop: 12, display: "flex", justifyContent: "flex-end" }}>
          <Button onClick={onClose}>Close</Button>
        </div>
        </Window>
      </div>
    </div>
  );
}