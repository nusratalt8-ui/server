import { useSyncExternalStore, useEffect, useCallback, useRef } from "react";
import { on } from "../socket/events";
import { send } from "../socket";
import { r } from "../routes";

const store = new Map();
const loaded = new Set();
const hasMoreMap = new Map();
const subs = new Set();
const EMPTY = [];

function notify() { subs.forEach((f) => f()); }

function setList(agentId, list) { store.set(agentId, list); notify(); }

function prepend(agentId, msgs) {
  const existing = new Set((store.get(agentId) || []).map((x) => x.id));
  const fresh = msgs.filter((m) => !existing.has(m.id));
  if (fresh.length > 0) setList(agentId, [...fresh, ...(store.get(agentId) || [])]);
}

function append(agentId, msg) {
  const list = store.get(agentId) || [];
  if (msg.id && list.some((x) => x.id === msg.id)) return;
  setList(agentId, [...list, msg]);
}

on("message", (m) => { if (m?.agent_id) append(m.agent_id, m); });

function subscribe(fn) { subs.add(fn); return () => subs.delete(fn); }

async function fetchPage(agentId, before) {
  const url = before
    ? `${r.messages(agentId)}?before=${encodeURIComponent(before)}`
    : r.messages(agentId);
  const res = await fetch(url, { credentials: "same-origin" });
  if (!res.ok) return { messages: [], has_more: false };
  return res.json();
}

export function useChat(agentId) {
  const messages = useSyncExternalStore(
    subscribe,
    () => store.get(agentId) || EMPTY,
    () => store.get(agentId) || EMPTY,
  );

  const loadingRef = useRef(false);

  useEffect(() => {
    if (!agentId || loaded.has(agentId)) return;
    loaded.add(agentId);
    fetchPage(agentId, null).then(({ messages: msgs, has_more }) => {
      hasMoreMap.set(agentId, !!has_more);
      if (Array.isArray(msgs)) setList(agentId, msgs);
    }).catch(() => {});
  }, [agentId]);

  const loadMore = useCallback(async () => {
    if (!agentId || loadingRef.current || !hasMoreMap.get(agentId)) return;
    const list = store.get(agentId) || [];
    if (list.length === 0) return;
    loadingRef.current = true;
    try {
      const { messages: msgs, has_more } = await fetchPage(agentId, list[0].id);
      hasMoreMap.set(agentId, !!has_more);
      if (Array.isArray(msgs) && msgs.length > 0) prepend(agentId, msgs);
    } catch {} finally {
      loadingRef.current = false;
    }
  }, [agentId]);

  const sendMessage = useCallback((text, attachments = []) => {
    const t = (text || "").trim();
    if (!t && attachments.length === 0) return;
    send("chat_send", { agent_id: agentId, text: t, attachments });
  }, [agentId]);

  return { messages, sendMessage, loadMore, hasMore: hasMoreMap.get(agentId) ?? false };
}