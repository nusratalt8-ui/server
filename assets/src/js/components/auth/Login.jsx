import { useState } from "react";
import AuthLayout from "./AuthLayout";
import Field from "../ui/Field";
import Spinner from "../ui/Spinner";

export default function Login({ onLogin, onShowRegister }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setError("");
    setLoading(true);
    try {
      await onLogin(username.trim(), password);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Log In">
      <div onKeyDown={(e) => e.key === "Enter" && submit()}>
        <h2 style={{
          fontSize: 13,
          fontWeight: 600,
          textTransform: "uppercase",
          letterSpacing: "0.08em",
          color: "var(--muted)",
          marginBottom: 20,
          textAlign: "center",
        }}>
          Member Login
        </h2>
        <Field label="Username" value={username} onChange={setUsername} autoFocus autoComplete="username" />
        <Field label="Password" type="password" value={password} onChange={setPassword} autoComplete="current-password" />
        {error && <div className="auth-error mb-3">{error}</div>}
        <button onClick={submit} disabled={loading} className="auth-btn">
          {loading ? <><Spinner size={16} />Signing in…</> : "Sign In"}
        </button>
        <div style={{ textAlign: "center", marginTop: 16 }}>
          <button type="button" onClick={onShowRegister} className="auth-link">
            Need an account? Register
          </button>
        </div>
      </div>
    </AuthLayout>
  );
}