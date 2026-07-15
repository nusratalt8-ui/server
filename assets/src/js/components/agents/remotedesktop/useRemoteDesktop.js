import { useCallback, useEffect, useRef, useState } from "react";
import { send } from "../../../socket";
import { on, off } from "../../../socket/events";
import { CODE_TO_VK } from "./constants";

export default function useRemoteDesktop(agentId) {
  const canvasRef = useRef(null);
  const wrapRef = useRef(null);
  const sessId = useRef(Math.random().toString(36).slice(2)).current;
  const [fps, setFps] = useState(0);
  const [stale, setStale] = useState(false);
  const [res, setRes] = useState({ w: 0, h: 0 });
  const resRef = useRef({ w: 0, h: 0 });
  const [status, setStatus] = useState({});
  const [errMsg, setErrMsg] = useState(null);
  const lastFrame = useRef(0);
  const fc = useRef(0);
  const staleRef = useRef(false);
  const imgRef = useRef(null);
  const rectRef = useRef(null);
  const pendingMove = useRef(null);
  const moveActive = useRef(false);
  const keysDown = useRef(new Set());

  const [micOn, setMicOn] = useState(false);
  const micOnRef = useRef(false);
  const [micMuted, setMicMuted] = useState(false);
  const micMutedRef = useRef(false);
  const [micDevices, setMicDevices] = useState([]);
  const [micDevice, setMicDevice] = useState(0);
  const audioCtxRef = useRef(null);
  const gainNodeRef = useRef(null);
  const nextTimeRef = useRef(0);

  const updateRect = useCallback(() => {
    const c = canvasRef.current;
    const wrap = wrapRef.current;
    const dim = resRef.current;
    if (!c || !wrap || dim.w === 0) { rectRef.current = null; return; }
    const r = wrap.getBoundingClientRect();
    const ar = r.width / r.height;
    const vr = dim.w / dim.h;
    let rw, rh;
    if (vr > ar) { rw = r.width; rh = r.width / vr; }
    else { rh = r.height; rw = r.height * vr; }
    const left = (r.width - rw) / 2;
    const top = (r.height - rh) / 2;
    c.style.left = left + "px";
    c.style.top = top + "px";
    c.style.width = rw + "px";
    c.style.height = rh + "px";
    rectRef.current = { left: r.left + left, top: r.top + top, width: rw, height: rh, vw: dim.w, vh: dim.h };
  }, []);

  const toScreen = useCallback((e) => {
    const r = rectRef.current;
    if (!r) return null;
    return {
      x: Math.round(((e.clientX - r.left) / r.width) * r.vw),
      y: Math.round(((e.clientY - r.top) / r.height) * r.vh),
    };
  }, []);

  const action = useCallback((id, level) => {
    send("panel_action", { agent_id: agentId, id, ...(level !== undefined && { level }) });
    if (level === undefined) {
      setStatus(prev => ({ ...prev, [id]: !prev[id] }));
    }
  }, [agentId]);

  const releaseAllKeys = useCallback(() => {
    keysDown.current.forEach((vk) =>
      send("liveview_input", { agent_id: agentId, input: { type: "key", vk, down: false } })
    );
    keysDown.current.clear();
  }, [agentId]);

  const requestMicDevices = useCallback(() => {
    send("mic_list", { agent_id: agentId });
  }, [agentId]);

  const startMic = useCallback((deviceID) => {
    if (micOnRef.current) {
      send("mic_stop", { agent_id: agentId });
      micOnRef.current = false;
    }
    send("mic_start", { agent_id: agentId, device_id: deviceID });
    micOnRef.current = true;
    nextTimeRef.current = audioCtxRef.current ? audioCtxRef.current.currentTime + 0.05 : 0;
    setMicOn(true);
    setMicMuted(false);
  }, [agentId]);

  const stopMic = useCallback(() => {
    send("mic_stop", { agent_id: agentId });
    micOnRef.current = false;
    setMicOn(false);
  }, [agentId]);

  const toggleMute = useCallback(() => {
    setMicMuted(m => {
      const next = !m;
      micMutedRef.current = next;
      if (gainNodeRef.current) gainNodeRef.current.gain.value = next ? 0.0 : 1.0;
      return next;
    });
  }, []);

  useEffect(() => {
    send("panel_open", { agent_id: agentId });
    send("liveview_start", { agent_id: agentId, sess_id: sessId });

    const unButtons = on("panel_buttons", (p) => {
      if (!p || String(p.agent_id) !== agentId) return;
      setStatus(p.result ?? {});
    });

    const unResult = on("panel_result", (p) => {
      if (!p || String(p.agent_id) !== agentId) return;
      const r = p.result;
      if (r?.err) { setErrMsg(`${r.id}: ${r.err}`); setTimeout(() => setErrMsg(null), 4000); }
    });

    const unMicDevices = on("mic_devices", (p) => {
      if (!p || String(p.agent_id) !== agentId) return;
      const devices = Array.isArray(p.devices) ? p.devices : [];
      if (devices.length === 0) return;
      setMicDevices(devices);
      setMicDevice(devices[0].id);
    });

    const t1 = setInterval(() => {
      if (lastFrame.current > 0 && Date.now() - lastFrame.current > 3000)
        if (!staleRef.current) { staleRef.current = true; setStale(true); setFps(0); }
    }, 500);
    const t2 = setInterval(() => { setFps(fc.current); fc.current = 0; }, 1000);

    const drawFrame = async (blob) => {
      const c = canvasRef.current;
      if (!c) return;
      try {
        const bmp = await createImageBitmap(blob);
        if (imgRef.current) { imgRef.current.close(); }
        imgRef.current = bmp;
        lastFrame.current = Date.now();
        fc.current++;
        if (staleRef.current) { staleRef.current = false; setStale(false); }
        const resChanged = c.width !== bmp.width || c.height !== bmp.height;
        if (resChanged) {
          c.width = bmp.width;
          c.height = bmp.height;
          resRef.current = { w: bmp.width, h: bmp.height };
          setRes({ w: bmp.width, h: bmp.height });
        }
        updateRect();
        const ctx = c.getContext("2d");
        ctx.imageSmoothingEnabled = true;
        ctx.imageSmoothingQuality = "high";
        ctx.drawImage(bmp, 0, 0);
      } catch (_) {}
    };
    window.__wsFrameHandler = drawFrame;

    window.addEventListener("blur", releaseAllKeys);

    return () => {
      releaseAllKeys();
      window.removeEventListener("blur", releaseAllKeys);
      clearInterval(t1);
      clearInterval(t2);
      send("liveview_stop", { agent_id: agentId, sess_id: sessId });
      if (micOnRef.current) send("mic_stop", { agent_id: agentId });
      send("panel_close", { agent_id: agentId });
      if (audioCtxRef.current) { audioCtxRef.current.close(); audioCtxRef.current = null; }
      window.__wsFrameHandler = null;
      if (imgRef.current) { imgRef.current.close(); imgRef.current = null; }
      off("panel_buttons", unButtons);
      off("panel_result", unResult);
      off("mic_devices", unMicDevices);
    };
  }, [agentId, sessId, releaseAllKeys]);

  useEffect(() => {
    const w = wrapRef.current;
    if (!w) return;
    const obs = new ResizeObserver(updateRect);
    obs.observe(w);
    window.addEventListener("resize", updateRect);
    return () => { obs.disconnect(); window.removeEventListener("resize", updateRect); };
  }, [updateRect]);

  const onMouseMove = useCallback((e) => {
    const p = toScreen(e);
    if (!p) return;
    pendingMove.current = p;
    if (!moveActive.current) {
      moveActive.current = true;
      requestAnimationFrame(() => {
        moveActive.current = false;
        const m = pendingMove.current;
        if (m) send("liveview_input", { agent_id: agentId, input: { type: "move", x: m.x, y: m.y } });
      });
    }
  }, [agentId, toScreen]);

  const onMouseDown = useCallback((e) => {
    const p = toScreen(e);
    if (!p) return;
    const b = e.button === 2 ? 2 : 1;
    send("liveview_input", { agent_id: agentId, input: { type: "click", x: p.x, y: p.y, button: b, down: true } });
  }, [agentId, toScreen]);

  const onMouseUp = useCallback((e) => {
    const p = toScreen(e);
    if (!p) return;
    const b = e.button === 2 ? 2 : 1;
    send("liveview_input", { agent_id: agentId, input: { type: "click", x: p.x, y: p.y, button: b, down: false } });
  }, [agentId, toScreen]);

  const onWheel = useCallback((e) => {
    const p = toScreen(e);
    if (!p) return;
    send("liveview_input", { agent_id: agentId, input: { type: "scroll", x: p.x, y: p.y, delta: -Math.round(e.deltaY / 40) } });
  }, [agentId, toScreen]);

  const onKeyDown = useCallback((e) => {
    const vk = CODE_TO_VK[e.code];
    if (!vk) return;
    keysDown.current.add(vk);
    send("liveview_input", { agent_id: agentId, input: { type: "key", vk, down: true } });
    e.preventDefault();
  }, [agentId]);

  const onKeyUp = useCallback((e) => {
    const vk = CODE_TO_VK[e.code];
    if (!vk) return;
    keysDown.current.delete(vk);
    send("liveview_input", { agent_id: agentId, input: { type: "key", vk, down: false } });
    e.preventDefault();
  }, [agentId]);

  useEffect(() => {
    const handler = (data, offset = 0) => {
      if (!micOnRef.current || micMutedRef.current) return;
      if (!(data instanceof ArrayBuffer)) return;
      const pcmBytes = new Uint8Array(data, offset);
      if (pcmBytes.length < 2) return;

      if (!audioCtxRef.current) {
        audioCtxRef.current = new (window.AudioContext || window.webkitAudioContext)();
        gainNodeRef.current = audioCtxRef.current.createGain();
        gainNodeRef.current.connect(audioCtxRef.current.destination);
        gainNodeRef.current.gain.value = 1.0;
        nextTimeRef.current = audioCtxRef.current.currentTime;
      }
      const ctx = audioCtxRef.current;
      // Cap lookahead to 200ms to prevent runaway scheduling lag
      const now = ctx.currentTime;
      if (nextTimeRef.current > now + 0.2) nextTimeRef.current = now;

      const aligned = new Uint8Array(pcmBytes.length);
      aligned.set(pcmBytes);
      const pcm = new Int16Array(aligned.buffer, 0, aligned.length >> 1);
      const buf = ctx.createBuffer(1, pcm.length, 8000);
      const ch = buf.getChannelData(0);
      for (let i = 0; i < pcm.length; i++) ch[i] = pcm[i] / 32768.0;
      const src = ctx.createBufferSource();
      src.buffer = buf;
      src.connect(gainNodeRef.current);
      if (nextTimeRef.current < now) nextTimeRef.current = now;
      src.start(nextTimeRef.current);
      nextTimeRef.current += buf.duration;
    };
    window.__wsMicHandler = handler;
    return () => { window.__wsMicHandler = null; };
  }, []);

  const vol = typeof status.volume === "number" && status.volume >= 0 ? status.volume : 50;

  return {
    canvasRef, wrapRef, fps, stale, res, status, errMsg, vol,
    action, updateRect,
    onMouseMove, onMouseDown, onMouseUp, onWheel,
    onKeyDown, onKeyUp, releaseAllKeys,
    micOn, micMuted, micDevices, micDevice, setMicDevice,
    requestMicDevices, startMic, stopMic, toggleMute,
  };
}