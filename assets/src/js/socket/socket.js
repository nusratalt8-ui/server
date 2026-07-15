export function send(type, payload) {
  const frame = JSON.stringify({ type, payload });
  if (typeof window !== "undefined" && typeof window.__wsSend === "function") {
    window.__wsSend(frame);
  }
}