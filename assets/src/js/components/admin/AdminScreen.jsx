import { useState, useEffect, useCallback } from "react";
import { apiListUsers, apiSetUserPlan } from "../../api/admin";
import { on, off } from "../../socket/events";
import Button from "../ui/Button";

function UserModal({ user, onClose, onUpgrade, onDemote }) {
  if (!user) return null;
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,0.7)", zIndex: 100, display: "flex", alignItems: "center", justifyContent: "center" }} onClick={onClose}>
      <div style={{ background: "var(--panel-bg)", border: "1px solid var(--border)", padding: 24, minWidth: 320, maxWidth: 480 }} onClick={e => e.stopPropagation()}>
        <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16 }}>{user.username}</h2>
        <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 20, fontSize: 13 }}>
          <div>ID: <span style={{ color: "var(--muted)", fontFamily: "monospace", fontSize: 11 }}>{user.id}</span></div>
          <div>Plan: <span style={{ color: user.plan === 1 ? "#22c55e" : "#ef4444", fontWeight: 700 }}>{user.plan === 1 ? "PAID" : "FREE"}</span></div>
          <div>Joined: {new Date(user.created_at * 1000).toLocaleDateString()}</div>
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          {user.plan === 0 ? (
            <Button onClick={() => onUpgrade(user.id)}>Upgrade to Paid</Button>
          ) : (
            <Button onClick={() => onDemote(user.id)} color="red">Demote to Free</Button>
          )}
          <Button onClick={onClose}>Close</Button>
        </div>
      </div>
    </div>
  );
}

export default function AdminScreen() {
  const [users, setUsers] = useState([]);
  const [selected, setSelected] = useState(null);

  const fetchUsers = useCallback(async () => {
    try {
      const data = await apiListUsers();
      setUsers(data);
    } catch (e) {
      console.error("Failed to fetch users:", e);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
    const un = on("plan_updated", () => fetchUsers());
    return () => off("plan_updated", un);
  }, [fetchUsers]);

  const handleUpgrade = async (userID) => {
    try {
      await apiSetUserPlan(userID, 1);
      setSelected(null);
      fetchUsers();
    } catch (e) {
      console.error("Upgrade failed:", e);
    }
  };

  const handleDemote = async (userID) => {
    try {
      await apiSetUserPlan(userID, 0);
      setSelected(null);
      fetchUsers();
    } catch (e) {
      console.error("Demote failed:", e);
    }
  };

  return (
    <div style={{ padding: 24, maxWidth: 800 }}>
      <h1 style={{ fontSize: 18, fontWeight: 700, marginBottom: 20 }}>Admin Panel</h1>
      <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
        {users.map(u => (
          <div
            key={u.id}
            onClick={() => setSelected(u)}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              padding: "10px 14px",
              background: "var(--input-bg)",
              border: "1px solid var(--border)",
              cursor: "pointer",
              fontSize: 13,
            }}
            onMouseEnter={e => e.currentTarget.style.background = "var(--hover-bg)"}
            onMouseLeave={e => e.currentTarget.style.background = "var(--input-bg)"}
          >
            <span style={{ fontWeight: 600 }}>{u.username}</span>
            <span style={{
              fontSize: 10,
              fontWeight: 700,
              padding: "2px 8px",
              borderRadius: 4,
              color: u.plan === 1 ? "#22c55e" : "#ef4444",
              border: `1px solid ${u.plan === 1 ? "#22c55e" : "#ef4444"}`,
            }}>
              {u.plan === 1 ? "PAID" : "FREE"}
            </span>
          </div>
        ))}
      </div>
      {selected && (
        <UserModal
          user={selected}
          onClose={() => setSelected(null)}
          onUpgrade={handleUpgrade}
          onDemote={handleDemote}
        />
      )}
    </div>
  );
}