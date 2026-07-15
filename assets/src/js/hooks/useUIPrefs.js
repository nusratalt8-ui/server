import { useState, useEffect, useCallback } from "react";
import { apiGetUIPrefs, apiSetUIPrefs } from "../api/auth";
import { applyTheme, getThemeById } from "../themes";

let cache = null;
const subs = new Set();

function notify() { subs.forEach((fn) => fn()); }

const DEFAULTS = {
  "sidebar.group.online": false,
  "sidebar.group.offline": false,
};

async function load() {
  try {
    cache = { ...DEFAULTS, ...(await apiGetUIPrefs()) };
    if (cache.theme) applyTheme(getThemeById(cache.theme));
    notify();
  } catch (_) {
    cache = { ...DEFAULTS };
    notify();
  }
}

let saveTimer = null;
function scheduleSave() {
  clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    apiSetUIPrefs(cache).catch(() => {});
  }, 600);
}

export function useUIPrefs() {
  const [, rerender] = useState(0);

  useEffect(() => {
    const fn = () => rerender((n) => n + 1);
    subs.add(fn);
    if (cache === null) load();
    return () => subs.delete(fn);
  }, []);

  const get = useCallback((key, def) => {
    const defaults = { ...DEFAULTS };
    const fallback = key in defaults ? defaults[key] : def;
    if (!cache) return fallback;
    return key in cache ? cache[key] : fallback;
  }, []);

  const set = useCallback((key, val) => {
    cache = { ...cache, [key]: val };
    if (key === "theme") applyTheme(getThemeById(val));
    notify();
    scheduleSave();
  }, []);

  return { get, set, ready: cache !== null };
}