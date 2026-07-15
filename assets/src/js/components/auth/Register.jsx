import { useState } from "react";
import AuthLayout from "./AuthLayout";
import Field from "../ui/Field";
import Spinner from "../ui/Spinner";

export default function Register({ onRegister, onShowLogin }) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setError("");
    if (password !== confirm) {
      setError("passwords do not match");
      return;
    }
    setLoading(true);
    try {
      await onRegister(username.trim(), password);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout title="Register">
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
          Create Account
        </h2>
        <Field label="Username" value={username} onChange={setUsername} autoFocus autoComplete="username" />
        <Field label="Password" type="password" value={password} onChange={setPassword} autoComplete="new-password" />
        <Field label="Confirm Password" type="password" value={confirm} onChange={setConfirm} autoComplete="new-password" />
        {error && <div className="auth-error mb-3">{error}</div>}
        <button onClick={submit} disabled={loading} className="auth-btn">
          {loading ? <><Spinner size={16} />Creating…</> : "Create Account"}
        </button>
        <div style={{ textAlign: "center", marginTop: 16 }}>
          <button type="button" onClick={onShowLogin} className="auth-link">
            Already have an account? Log in
          </button>
        </div>
      </div>
    </AuthLayout>
  );
}