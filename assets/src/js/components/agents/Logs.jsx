import { useEffect, useRef, useState } from "react";
import { on } from "../../socket/events";
import { send } from "../../socket";

export default function Logs({ agent }) {
  const [entries, setEntries] = useState([]);
  const bottomRef = useRef(null);
  const agentId = String(agent.id);

  useEffect(() => {
    send("logs_get", { agent_id: agentId });
  }, [agentId]);

  useEffect(() => on("logs_history", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    setEntries(p.entries || []);
  }), [agentId]);

  useEffect(() => on("agent_log", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    setEntries((prev) => [...prev.slice(-499), p]);
  }), [agentId]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "auto" });
  }, [entries]);

  const fmt = (ts) => {
    const d = new Date(ts * 1000);
    return d.toLocaleTimeString("en-GB", { hour12: false });
  };

  return (
    <div className="h-full flex flex-col font-mono text-xs"
      style={{ background: "#0c0c0c", color: "#cccccc" }}>
      <div className="flex-1 overflow-y-auto p-3 vista-scroll"
        style={{ scrollbarColor: "#444 #0c0c0c" }}>
        {entries.map((e, i) => (
          <div key={i} style={{
            color: e.level === "error" ? "#f47c7c" : "#cccccc",
            whiteSpace: "pre-wrap",
            wordBreak: "break-all",
            lineHeight: 1.6,
          }}>
            <span style={{ color: "#555", userSelect: "none" }}>{fmt(e.time)} </span>
            <span style={{ color: e.level === "error" ? "#f47c7c" : "#4ec9b0", userSelect: "none" }}>
              [{e.level}]{" "}
            </span>
            {e.msg}
          </div>
        ))}
        {entries.length === 0 && (
          <div style={{ color: "#555" }}>no logs yet</div>
        )}
        <div ref={bottomRef} />
      </div>
      <div className="px-3 py-1 text-xs flex-shrink-0"
        style={{ borderTop: "1px solid #222", color: "#555", background: "#0c0c0c" }}>
        {entries.length} entries
      </div>
    </div>
  );
}