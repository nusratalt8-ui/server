const KEY = "agent_app_state_v1";

let cache = {};
try {
  cache = JSON.parse(localStorage.getItem(KEY)) || {};
} catch {
  cache = {};
}

let timer = null;
function persist() {
  clearTimeout(timer);
  timer = setTimeout(() => {
    try {
      localStorage.setItem(KEY, JSON.stringify(cache));
    } catch {}
  }, 300);
}

export function getAppState(agentId, appId) {
  return cache[`${agentId}:${appId}`] || {};
}

export function setAppState(agentId, appId, patch) {
  const k = `${agentId}:${appId}`;
  cache[k] = { ...cache[k], ...patch };
  persist();
}