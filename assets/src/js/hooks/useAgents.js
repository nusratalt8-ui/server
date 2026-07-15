import { useSyncExternalStore } from "react";
import { on } from "../socket";
import { apiListAgents, apiListAllAgents, apiListUserAgents } from "../api/agents";

const connectAudio = new Audio("/media/sound.mp3");
connectAudio.volume = 0.4;

function playConnectSound() {
  try { connectAudio.currentTime = 0; connectAudio.play().catch(() => {}); } catch (_) {}
}

let state = {
  agents: [],
  agentsTotal: 0,
  agentsOnline: 0,
  adminTotal: 0,
  adminOnline: 0,
  perUserCounts: {},
  userAgents: {},
  ready: false,
};
const subs = new Set();

function set(next) {
  state = { ...state, ...next };
  subs.forEach((fn) => fn());
}

function applyAgents(p) {
  if (!p || !Array.isArray(p.items)) return;
  const newOnline = p.items.filter((a) => a.online && !state.agents.find((q) => q.id === a.id && q.online));
  if (state.ready && newOnline.length > 0) {
    playConnectSound();
    const orig = document.title.replace(/^\(\d+\) /, "");
    document.title = `(${newOnline.length}) ${orig}`;
    setTimeout(() => { document.title = orig; }, 4000);
  }
  if (state.ready && state.agents.length > p.items.length) {
    const incoming = new Map(p.items.map((a) => [a.id, a]));
    const brandNew = p.items.filter((a) => !state.agents.find((q) => q.id === a.id));
    const merged = state.agents.map((a) => incoming.has(a.id) ? incoming.get(a.id) : a);
    set({ agents: [...brandNew, ...merged], agentsTotal: p.total ?? p.items.length, agentsOnline: p.online_total ?? 0, ready: true });
  } else {
    set({ agents: p.items, agentsTotal: p.total ?? p.items.length, agentsOnline: p.online_total ?? 0, ready: true });
  }
}

function applyAdminAgents(p) {
  if (!p) return;
  const counts = p.per_user ?? state.perUserCounts;
  const userAgents = { ...state.userAgents };
  if (Array.isArray(p.items)) {
    const byUser = {};
    for (const a of p.items) {
      if (!byUser[a.user_id]) byUser[a.user_id] = [];
      byUser[a.user_id].push(a);
    }
    for (const [uid, incoming] of Object.entries(byUser)) {
      const existing = userAgents[uid];
      if (!existing) continue;
      const incomingMap = new Map(incoming.map((a) => [a.id, a]));
      const brandNew = incoming.filter((a) => !existing.find((q) => q.id === a.id));
      const merged = existing.map((a) => incomingMap.has(a.id) ? incomingMap.get(a.id) : a);
      userAgents[uid] = [...brandNew, ...merged];
    }
  }
  set({ adminTotal: p.total ?? state.adminTotal, adminOnline: p.online_total ?? state.adminOnline, perUserCounts: counts, userAgents, ready: true });
}

on("agents", applyAgents);
on("admin_agents", applyAdminAgents);
on("@open", () => { fetchAgents(); });

export async function fetchAgents() {
  try { applyAgents(await apiListAgents(0, 50)); } catch {}
}

export async function fetchAllAgents() {
  try {
    const p = await apiListAllAgents(0, 50);
    if (!p) return;
    applyAdminAgents(p);
  } catch {}
}

export async function loadMoreAgents(offset) {
  try {
    const p = await apiListAgents(offset, 50);
    if (!p || !Array.isArray(p.items)) return;
    set({ agents: [...state.agents.slice(0, offset), ...p.items], agentsTotal: p.total ?? state.agentsTotal, agentsOnline: p.online_total ?? state.agentsOnline });
  } catch {}
}

export function setUserAgents(uid, agents) {
  set({ userAgents: { ...state.userAgents, [uid]: agents } });
}

export async function loadMoreUserAgents(uid, offset) {
  try {
    const p = await apiListUserAgents(uid, offset, 50);
    if (!p || !Array.isArray(p.items)) return;
    const existing = state.userAgents[uid] || [];
    set({ userAgents: { ...state.userAgents, [uid]: [...existing.slice(0, offset), ...p.items] } });
  } catch {}
}

function subscribe(fn) { subs.add(fn); return () => subs.delete(fn); }
function snapshot() { return state; }

export function useAgents() {
  return useSyncExternalStore(subscribe, snapshot, snapshot);
}