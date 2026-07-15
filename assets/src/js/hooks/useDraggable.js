import { useRef, useState, useCallback, useEffect } from "react";

function clampPos(p) {
  return {
    x: Math.min(Math.max(p.x, 0), Math.max(0, window.innerWidth - 100)),
    y: Math.min(Math.max(p.y, 0), Math.max(0, window.innerHeight - 40)),
  };
}

export default function useDraggable(initial, onChange) {
  const [pos, setPos] = useState(() =>
    clampPos(
      initial || {
        x: Math.max(8, Math.round((window.innerWidth - 660) / 2)),
        y: Math.max(8, Math.round(window.innerHeight * 0.1)),
      }
    )
  );
  const start = useRef(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    const onResize = () => setPos((p) => clampPos(p));
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  const onPointerDown = useCallback((e) => {
    if (e.button !== 0) return;
    e.preventDefault();
    start.current = { px: e.clientX, py: e.clientY, x: pos.x, y: pos.y };
    const move = (ev) => {
      if (!start.current) return;
      const next = clampPos({
        x: start.current.x + ev.clientX - start.current.px,
        y: start.current.y + ev.clientY - start.current.py,
      });
      setPos(next);
      onChangeRef.current?.(next);
    };
    const up = () => {
      start.current = null;
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
      window.removeEventListener("pointercancel", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
    window.addEventListener("pointercancel", up);
  }, [pos.x, pos.y]);

  return { pos, onPointerDown };
}