import { useState, useEffect } from "react";
import { apiMe, apiRegister, apiLogin, apiLogout } from "../api/auth";
import { on, off } from "../socket/events";

export function useAuth() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    apiMe()
      .then((u) => { if (alive) setUser(u); })
      .catch(() => {})
      .finally(() => { if (alive) setLoading(false); });
    const unPlan = on("plan_updated", (p) => {
      setUser((prev) => {
        if (!prev || prev.user_id !== p?.user_id) return prev;
        return { ...prev, plan: p.plan };
      });
    });
    return () => { alive = false; off("plan_updated", unPlan); };
  }, []);

  const login = async (username, password) => {
    await apiLogin(username, password);
    const me = await apiMe();
    setUser(me);
    return me;
  };

  const register = async (username, password) => {
    await apiRegister(username, password);
    const me = await apiMe();
    setUser(me);
    return me;
  };

  const logout = async () => {
    try { await apiLogout(); } catch {}
    setUser(null);
  };

  return { user, loading, ready: !loading, login, register, logout };
}