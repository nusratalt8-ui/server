import { useState, useRef, useCallback } from "react";

const EDGE = 24;
const MAX_STEP = 14;

export default function useRubberBand({ rowSelector = "tbody tr[data-name]", onSelect, onClear }) {
  const [band, setBand] = useState(null);
  const ref = useRef(null);
  const onSelectRef = useRef(onSelect);
  const onClearRef = useRef(onClear);
  onSelectRef.current = onSelect;
  onClearRef.current = onClear;

  const onPointerDown = useCallback((e) => {
    if (e.button !== 0) return;
    const onRow = !!e.target.closest("tr");
    const wrap = ref.current;
    if (!wrap) return;
    e.preventDefault();
    const r0 = wrap.getBoundingClientRect();
    const sx = e.clientX - r0.left + wrap.scrollLeft;
    const sy = e.clientY - r0.top + wrap.scrollTop;
    let last = { x: e.clientX, y: e.clientY };
    let raf = 0;

    const update = () => {
      const r = wrap.getBoundingClientRect();
      const cx = Math.min(Math.max(last.x, r.left), r.right) - r.left + wrap.scrollLeft;
      const cy = Math.min(Math.max(last.y, r.top), r.bottom) - r.top + wrap.scrollTop;
      const nb = { x: Math.min(sx, cx), y: Math.min(sy, cy), w: Math.abs(cx - sx), h: Math.abs(cy - sy) };
      setBand(nb);
      const names = new Set();
      wrap.querySelectorAll(rowSelector).forEach((row) => {
        const rr = row.getBoundingClientRect();
        const top = rr.top - r.top + wrap.scrollTop;
        const bottom = rr.bottom - r.top + wrap.scrollTop;
        if (nb.y < bottom && nb.y + nb.h > top) names.add(row.dataset.name);
      });
      onSelectRef.current?.(names);
    };

    const tick = () => {
      const r = wrap.getBoundingClientRect();
      let dy = 0;
      if (last.y < r.top + EDGE) dy = -Math.min(MAX_STEP, (r.top + EDGE - last.y) / 2);
      else if (last.y > r.bottom - EDGE) dy = Math.min(MAX_STEP, (last.y - (r.bottom - EDGE)) / 2);
      if (dy !== 0) {
        wrap.scrollTop += dy;
        update();
      }
      raf = requestAnimationFrame(tick);
    };

    const move = (ev) => {
      last = { x: ev.clientX, y: ev.clientY };
      update();
    };
    let dragging = false;
    const origMove = move;
    const move2 = (ev) => {
      const r = wrap.getBoundingClientRect();
      const ux = ev.clientX - r.left + wrap.scrollLeft;
      const uy = ev.clientY - r.top + wrap.scrollTop;
      if (!dragging && (Math.abs(ux - sx) > 4 || Math.abs(uy - sy) > 4)) {
        dragging = true;
        if (onRow) e.preventDefault();
      }
      if (dragging) origMove(ev);
    };
    const up = (ev) => {
      cancelAnimationFrame(raf);
      const r = wrap.getBoundingClientRect();
      const ux = ev.clientX - r.left + wrap.scrollLeft;
      const uy = ev.clientY - r.top + wrap.scrollTop;
      if (!dragging && Math.abs(ux - sx) < 3 && Math.abs(uy - sy) < 3) onClearRef.current?.();
      setBand(null);
      window.removeEventListener("pointermove", move2);
      window.removeEventListener("pointerup", up);
      window.removeEventListener("pointercancel", up);
    };
    window.addEventListener("pointermove", move2);
    window.addEventListener("pointerup", up);
    window.addEventListener("pointercancel", up);
    raf = requestAnimationFrame(tick);
  }, [rowSelector]);

  return { band, ref, onPointerDown };
}