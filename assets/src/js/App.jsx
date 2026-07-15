import { useEffect } from "react";
import { useRouter } from "./router";
import { useAuth } from "./hooks/useAuth";
import { useUIPrefs } from "./hooks/useUIPrefs";
import { ROUTES } from "./routes";
import { useSocket } from "./hooks/useSocket";
import { fetchAgents, fetchAllAgents } from "./hooks/useAgents";
import Login from "./components/auth/Login";
import Register from "./components/auth/Register";
import Welcome from "./components/welcome/Welcome";
import SettingsScreen from "./components/settings/SettingsScreen";
import AgentView from "./components/agents/AgentView";
import AppLayout from "./components/layout/AppLayout";
import Socks5Modal from "./components/agents/Socks5Modal";
import AdminScreen from "./components/admin/AdminScreen";

export default function App() {
  const { path, navigate, match } = useRouter();
  const { user, loading, login, register, logout } = useAuth();

  useEffect(() => {
    if (loading) return;
    if (!user && path !== ROUTES.login && path !== ROUTES.register) navigate(ROUTES.login, true);
    if (user && (path === ROUTES.login || path === ROUTES.register)) navigate(ROUTES.home, true);
  }, [loading, user, path]);

  useSocket(user);
  useUIPrefs();
  useEffect(() => {
    if (!user) return;
    fetchAgents();
    if (user.is_admin) fetchAllAgents();
  }, [user?.id]);


  if (!user && !loading) {
    if (path === ROUTES.register) {
      return <Register onRegister={register} onShowLogin={() => navigate(ROUTES.login)} />;
    }
    return <Login onLogin={login} onShowRegister={() => navigate(ROUTES.register)} />;
  }

  if (loading || !user) {
    return <div className="min-h-screen" style={{ background: "var(--desktop)" }} />;
  }

  const agentMatch = match("/agents/:id");
  const settingsTab = match("/settings/:tab");
  const content = agentMatch ? (
    <AgentView id={agentMatch.id} />
  ) : path === ROUTES.admin ? (
    user?.is_admin ? <AdminScreen /> : <Welcome user={user} />
  ) : path === ROUTES.settings || settingsTab ? (
    <SettingsScreen defaultTab={settingsTab?.tab} />
  ) : (
    <Welcome user={user} />
  );

  return (
    <AppLayout user={user} onLogout={logout}>
      {content}
      <Socks5Modal />
    </AppLayout>
  );
}
