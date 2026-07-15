import { useEffect, useLayoutEffect, useRef } from "react";
import Message from "./Message";

const NEAR_BOTTOM_PX = 150;

export default function MessageList({ messages, onResend, loadMore, hasMore }) {
  const containerRef = useRef(null);
  const contentRef = useRef(null);
  const didInit = useRef(false);
  const pinned = useRef(true);

  const onScroll = () => {
    const el = containerRef.current;
    if (!el) return;
    pinned.current = el.scrollHeight - el.scrollTop - el.clientHeight < NEAR_BOTTOM_PX;
  };

  useLayoutEffect(() => {
    const el = containerRef.current;
    if (!el || messages.length === 0) return;
    if (!didInit.current) {
      didInit.current = true;
      el.scrollTop = el.scrollHeight;
      return;
    }
    if (pinned.current) {
      el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
    }
  }, [messages]);

  useEffect(() => {
    const content = contentRef.current;
    const el = containerRef.current;
    if (!content || !el || typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => {
      if (pinned.current) el.scrollTop = el.scrollHeight;
    });
    ro.observe(content);
    return () => ro.disconnect();
  }, []);

  return (
    <div ref={containerRef} onScroll={onScroll} className="flex-1 overflow-y-auto p-3 space-y-2 vista-scroll">
      <div ref={contentRef} className="space-y-2">
        {hasMore && (
          <div style={{ textAlign: "center", paddingBottom: 8 }}>
            <button onClick={loadMore} className="bevel-out" style={{ fontSize: 11, padding: "3px 12px", background: "var(--window-face)", color: "var(--muted)", border: "none", cursor: "pointer" }}>
              Load older messages
            </button>
          </div>
        )}
        {messages.length === 0 ? (
          <p className="opacity-60 text-sm">No messages yet. Try <b>.help</b></p>
        ) : (
          messages.map((m) => <Message key={m.id} msg={m} onResend={onResend} />)
        )}
      </div>
    </div>
  );
}