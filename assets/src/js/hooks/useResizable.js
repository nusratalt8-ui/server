import { useRef, useState, useCallback, useEffect } from "react";

function clampSize(s) {
  return {
    w: Math.min(s.w, Math.max(0, window.innerWidth - 16)),
    h: Math.min(s.h, Math.max(0, window.innerHeight - 60)),
  };
}

export default function useResizable(initial = { w: 480, h: 360 }, min = { w: 280, h: 180 }, onChange) {
  const [size, setSize] = useState(() => clampSize(initial));
  const start = useRef(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    const onResize = () => setSize((s) => clampSize(s));
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  const onResizeDown = useCallback((e) => {
    if (e.button !== 0) return;
    e.preventDefault();
    e.stopPropagation();
    start.current = { px: e.clientX, py: e.clientY, w: size.w, h: size.h };
    const move = (ev) => {
      if (!start.current) return;
      const next = {
        w: Math.max(min.w, start.current.w + ev.clientX - start.current.px),
        h: Math.max(min.h, start.current.h + ev.clientY - start.current.py),
      };
      setSize(next);
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
  }, [size.w, size.h, min.w, min.h]);

  return { size, onResizeDown };
}