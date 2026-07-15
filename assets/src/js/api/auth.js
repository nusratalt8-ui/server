import { request } from "./client";
import { r } from "../routes";

export function apiRegister(username, password) {
  return request("POST", r.auth.register(), { username, password });
}

export function apiICEServers() {
  return request("GET", r.auth.iceServers());
}

export function apiGetUIPrefs() {
  return request("GET", r.auth.uiPrefs());
}

export function apiSetUIPrefs(prefs) {
  return request("PUT", r.auth.uiPrefs(), prefs);
}

export function apiLogin(username, password) {
  return request("POST", r.auth.login(), { username, password });
}

export function apiLogout() {
  return request("POST", r.auth.logout());
}

export async function apiMe() {
  try {
    return await request("GET", r.me());
  } catch {
    return null;
  }
}