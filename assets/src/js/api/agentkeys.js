import { request } from "./client";
import { r } from "../routes";

export function apiAgentKeyInfo() {
  return request("GET", r.agentKey.info());
}

export function apiRotateAgentKey() {
  return request("POST", r.agentKey.rotate());
}