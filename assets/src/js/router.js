import { useState, useEffect } from "react";

export function navigate(to, replace = false) {
  if (replace) window.history.replaceState({}, "", to);
  else window.history.pushState({}, "", to);
  window.dispatchEvent(new Event("am:navigate"));
}

export function useRouter() {
  const [path, setPath] = useState(window.location.pathname);

  useEffect(() => {
    const onChange = () => setPath(window.location.pathname);
    window.addEventListener("popstate", onChange);
    window.addEventListener("am:navigate", onChange);
    return () => {
      window.removeEventListener("popstate", onChange);
      window.removeEventListener("am:navigate", onChange);
    };
  }, []);

  const match = (pattern) => {
    const pp = pattern.split("/");
    const ap = path.split("/");
    if (pp.length !== ap.length) return null;
    const params = {};
    for (let i = 0; i < pp.length; i++) {
      if (pp[i].startsWith(":")) params[pp[i].slice(1)] = ap[i];
      else if (pp[i] !== ap[i]) return null;
    }
    return params;
  };

  return { path, navigate, match };
}