import { useState, useEffect, useRef } from "react";

export function useAltTab(openIds) {
  const [visible, setVisible] = useState(false);
  const [index, setIndex] = useState(0);
  const committedIndex = useRef(0);

  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.ctrlKey && (e.key === "`" || e.key === "~")) {
        e.preventDefault();
        if (openIds.length === 0) return;
        setVisible(true);
        setIndex((prev) => {
          const next = e.shiftKey
            ? (prev - 1 + openIds.length) % openIds.length
            : (prev + 1) % openIds.length;
          committedIndex.current = next;
          return next;
        });
      }
    };
    const handleKeyUp = (e) => {
      if (e.key === "Control") {
        setVisible(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
    };
  }, [openIds]);

  // clamp index if open windows change
  useEffect(() => {
    if (index >= openIds.length && openIds.length > 0) {
      setIndex(openIds.length - 1);
    }
  }, [openIds]);

  return { visible, index };
}   