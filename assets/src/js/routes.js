// SPA page paths — what the router renders.
export const ROUTES = {
  home: "/",
  login: "/login",
  register: "/register",
  settings: "/settings",
  agents: "/agents",
  admin: "/admin",
};

export function agentPath(id) {
  return `/agents/${id}`;
}

// API boundary — every backend URL lives here; nothing else builds /api strings.
export const V1 = "/api/v1";

export const r = {
  me: () => `${V1}/me`,
  ws: () => `${V1}/ws`,
  auth: {
    register: () => `${V1}/register`,
    login: () => `${V1}/login`,
    logout: () => `${V1}/logout`,
    iceServers: () => `${V1}/ice-servers`,
    uiPrefs: () => `${V1}/uiprefs`,
  },
  agentKey: {
    info: () => `${V1}/agent-key`,
    rotate: () => `${V1}/agent-key/rotate`,
  },
  build: {
    start: () => `${V1}/build/start`,
    download: (file) => `${V1}/build/download/${file}`,
    icon: () => `${V1}/build/icon`,
  },
  plans: {
    users: () => `${V1}/users`,
    user: (id) => `${V1}/plans/user/${id}`,
  },
  agents: () => `${V1}/agents`,
  agentsAll: () => `${V1}/agents/all`,
  agentsByUser: (uid) => `${V1}/agents/user/${uid}`,
  agent: (id) => `${V1}/agents/${id}`,
  messages: (id) => `${V1}/messages/${id}`,
  attachments: () => `${V1}/attachments`,
  attachment: (id) => `${V1}/attachments/${id}`,
};