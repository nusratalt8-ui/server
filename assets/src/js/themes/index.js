import dark from "./dark";
import coal from "./coal";
import light from "./light";
import midnight from "./midnight";

export const THEMES = [dark, coal, light, midnight];

const byId = {};
for (const t of THEMES) byId[t.id] = t;

export const defaultTheme = dark;

export function getThemeById(id) {
  return byId[id] || defaultTheme;
}

let _activeId = defaultTheme.id;
export function activeThemeId() {
  return _activeId;
}

export function applyTheme(theme = defaultTheme) {
  const root = document.documentElement;
  const vars = (theme && theme.vars) || {};
  for (const key in vars) {
    root.style.setProperty(`--${key}`, vars[key]);
  }
  if (theme && theme.id) _activeId = theme.id;
}