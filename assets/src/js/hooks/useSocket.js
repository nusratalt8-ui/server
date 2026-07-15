import { useEffect, useRef, useCallback } from "react";
import { emit } from "../socket/events";
import { r } from "../routes";

export function useSocket(user) {
  const wsRef = useRef(null);
  const reconnectRef = useRef(null);
  const attemptsRef = useRef(0);
  const userRef = useRef(user);
  userRef.current = user;
  const queueRef = useRef([]);

  const connect = useCallback(() => {
    if (wsRef.current && (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING)) return;
    if (!userRef.current) return;
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(`${proto}//${window.location.host}${r.ws()}`);
    ws.binaryType = "arraybuffer";
    wsRef.current = ws;

    ws.onopen = () => {
      attemptsRef.current = 0;
      const queued = queueRef.current.splice(0);
      queued.forEach((frame) => ws.send(frame));
      emit("@open", {});
    };

    ws.onmessage = (e) => {
      if (e.data instanceof ArrayBuffer) {
        const tag = new Uint8Array(e.data, 0, 1)[0];
        if (tag === 0x01) {
          if (window.__wsFrameHandler) window.__wsFrameHandler(new Blob([new Uint8Array(e.data, 1)]));
        } else if (tag === 0x02) {
          if (window.__wsMicHandler) window.__wsMicHandler(e.data, 1);
        } else if (tag === 0x03) {
          if (window.__wsCamHandler) window.__wsCamHandler(new Blob([new Uint8Array(e.data, 1)]));
        }
        return;
      }
      let msg;
      try { msg = JSON.parse(e.data); } catch { return; }
      if (!msg || !msg.type) return;
      if (msg.type === "pong") {
        if (typeof window.__wsHandlePong === "function") window.__wsHandlePong(msg.payload?.id);
        return;
      }
      if (msg.type === "socks5_active") {
        window.__socks5Modal?.(msg.payload);
        emit(msg.type, msg.payload);
        return;
      }
      if (msg.type === "socks5_inactive") {
        window.__socks5Modal?.(null);
        emit(msg.type, msg.payload);
        return;
      }
      emit(msg.type, msg.payload);
    };

    ws.onclose = () => {
      wsRef.current = null;
      emit("@close", {});
      if (!userRef.current) return;
      attemptsRef.current++;
      const delay = Math.min(500 * 2 ** attemptsRef.current, 8000);
      reconnectRef.current = setTimeout(connect, delay);
    };

    ws.onerror = () => {
    };
  }, []);

  useEffect(() => {
    connect();
    const onUnload = () => {
      if (reconnectRef.current) clearTimeout(reconnectRef.current);
      if (wsRef.current) { wsRef.current.onclose = null; wsRef.current.close(); wsRef.current = null; }
    };
    window.addEventListener("beforeunload", onUnload);
    window.addEventListener("pagehide", onUnload);
    return () => {
      window.removeEventListener("beforeunload", onUnload);
      window.removeEventListener("pagehide", onUnload);
      onUnload();
    };
  }, [connect]);

  useEffect(() => {
    if (user && (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN)) {
      attemptsRef.current = 0;
      connect();
    }
    if (!user && wsRef.current) { wsRef.current.close(); }
  }, [user, connect]);

  const pingTimers = useRef({});

  window.__wsPing = () => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return;
    const id = Date.now().toString();
    pingTimers.current[id] = performance.now();
    wsRef.current.send(JSON.stringify({ type: "ping", payload: { id } }));
  };

  window.__wsHandlePong = (id) => {
    const start = pingTimers.current[id];
    if (!start) return null;
    delete pingTimers.current[id];
    return Math.round(performance.now() - start);
  };

  window.__wsSend = (frame) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(frame);
      return true;
    }
    queueRef.current.push(frame);
    return false;
  };

  return wsRef;
}