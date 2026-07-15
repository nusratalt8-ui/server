import { useEffect, useRef, useState } from "react";
import { on, off } from "../../socket/events";
import { send } from "../../socket";

const MAX_POINTS = 300;

function drawGraph(canvas, points) {
  if (!canvas || points.length < 2) return;
  const ctx = canvas.getContext("2d");
  const dpr = window.devicePixelRatio || 1;
  const rect = canvas.getBoundingClientRect();
  canvas.width = rect.width * dpr;
  canvas.height = rect.height * dpr;
  ctx.scale(dpr, dpr);
  const w = rect.width;
  const h = rect.height;

  const css = getComputedStyle(document.documentElement);
  const accent = (css.getPropertyValue("--accent") || "#c2734e").trim();
  const grid = (css.getPropertyValue("--edge-dark") || "#6b6459").trim();
  const muted = (css.getPropertyValue("--muted") || "#6f6a60").trim();

  ctx.clearRect(0, 0, w, h);

  const vals = points.map((p) => p.ms);
  const max = Math.max(...vals, 1);
  const min = Math.min(...vals);
  const range = max - min || 1;
  const padL = 38, padR = 8, padT = 10, padB = 20;
  const gw = w - padL - padR;
  const gh = h - padT - padB;

  const toX = (i) => padL + (i / (MAX_POINTS - 1)) * gw;
  const toY = (v) => padT + (1 - (v - min) / range) * gh;

  ctx.lineWidth = 1;
  ctx.setLineDash([]);
  [0, 0.25, 0.5, 0.75, 1].forEach((t) => {
    const v = min + t * range;
    const y = padT + (1 - t) * gh;
    ctx.save();
    ctx.globalAlpha = 0.2;
    ctx.strokeStyle = grid;
    ctx.beginPath();
    ctx.moveTo(padL, y);
    ctx.lineTo(w - padR, y);
    ctx.stroke();
    ctx.restore();
    ctx.fillStyle = muted;
    ctx.font = `9px monospace`;
    ctx.textAlign = "right";
    ctx.fillText(Math.round(v) + "ms", padL - 4, y + 3);
  });

  ctx.save();
  ctx.globalAlpha = 0.16;
  ctx.beginPath();
  points.forEach((p, i) => {
    const x = toX(MAX_POINTS - points.length + i);
    const y = toY(p.ms);
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
  });
  ctx.lineTo(toX(MAX_POINTS - 1), padT + gh);
  ctx.lineTo(toX(MAX_POINTS - points.length), padT + gh);
  ctx.closePath();
  ctx.fillStyle = accent;
  ctx.fill();
  ctx.restore();

  ctx.strokeStyle = accent;
  ctx.lineWidth = 1.5;
  ctx.lineJoin = "round";
  ctx.setLineDash([]);
  ctx.beginPath();
  points.forEach((p, i) => {
    const x = toX(MAX_POINTS - points.length + i);
    const y = toY(p.ms);
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
  });
  ctx.stroke();

  const last = points[points.length - 1];
  const lx = toX(MAX_POINTS - 1);
  const ly = toY(last.ms);
  ctx.fillStyle = accent;
  ctx.beginPath();
  ctx.arc(lx, ly, 3, 0, Math.PI * 2);
  ctx.fill();
}

export default function Latency({ agent }) {
  const [points, setPoints] = useState([]);
  const canvasRef = useRef(null);
  const agentId = String(agent.id);

  useEffect(() => {
    send("latency_open", { agent_id: agentId });
    return () => send("latency_close", { agent_id: agentId });
  }, [agentId]);

  useEffect(() => {
    const onHistory = (p) => {
      if (!p || String(p.agent_id) !== agentId) return;
      setPoints(p.points || []);
    };
    const onPoint = (p) => {
      if (!p || String(p.id) !== agentId) return;
      setPoints((prev) => [...prev.slice(-(MAX_POINTS - 1)), { ms: p.ms, t: p.t }]);
    };
    const u1 = on("latency_history", onHistory);
    const u2 = on("agent_ping", onPoint);
    return () => { off("latency_history", onHistory); off("agent_ping", onPoint); };
  }, [agentId]);

  useEffect(() => {
    drawGraph(canvasRef.current, points);
  }, [points]);

  const vals = points.map((p) => p.ms);
  const current = vals.length ? vals[vals.length - 1] : null;
  const avg = vals.length ? Math.round(vals.reduce((a, b) => a + b, 0) / vals.length) : null;
  const min = vals.length ? Math.min(...vals) : null;
  const max = vals.length ? Math.max(...vals) : null;

  const stat = (label, val) => (
    <div className="bevel-in px-3 py-1.5 flex flex-col items-center gap-0.5" style={{ minWidth: 70 }}>
      <span className="text-xs opacity-50">{label}</span>
      <span className="font-mono font-bold text-sm">{val !== null ? `${val}ms` : "—"}</span>
    </div>
  );

  return (
    <div className="h-full flex flex-col gap-2 p-2" style={{ background: "var(--window-face)" }}>
      <div className="flex gap-2">
        {stat("Now", current)}
        {stat("Avg", avg)}
        {stat("Min", min)}
        {stat("Max", max)}
      </div>
      <div className="flex-1 bevel-in overflow-hidden" style={{ background: "var(--input-bg)" }}>
        <canvas ref={canvasRef} style={{ width: "100%", height: "100%", display: "block" }} />
      </div>
    </div>
  );
}