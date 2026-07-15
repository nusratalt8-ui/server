import { request } from "./client";
import { r } from "../routes";

export function apiListAgents(offset = 0, limit = 50) {
  return request("GET", `${r.agents()}?offset=${offset}&limit=${limit}`);
}

export function apiListAllAgents(offset = 0, limit = 50) {
  return request("GET", `${r.agentsAll()}?offset=${offset}&limit=${limit}`);
}

export function apiListUserAgents(uid, offset = 0, limit = 50) {
  return request("GET", `${r.agentsByUser(uid)}?offset=${offset}&limit=${limit}`);
}

export function apiGetAgent(id) {
  return request("GET", r.agent(id));
}

export function apiDeleteAgent(id) {
  return request("DELETE", r.agent(id));
}