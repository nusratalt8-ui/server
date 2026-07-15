import { request } from "./client";
import { r } from "../routes";

export function apiListUsers() {
  return request("GET", r.plans.users());
}

export function apiSetUserPlan(userID, plan) {
  return request("PUT", r.plans.user(userID), { plan });
}