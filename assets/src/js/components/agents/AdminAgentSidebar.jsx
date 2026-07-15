import { useCallback, useEffect, useRef } from "react";
import AgentEntry from "./AgentEntry";
import { useAgents, setUserAgents, loadMoreUserAgents } from "../../hooks/useAgents";
import { usePagination } from "../../hooks/usePagination";
import { useUIPrefs } from "../../hooks/useUIPrefs";
import { apiListUserAgents } from "../../api/agents";

function CollapseBtn({ open }) {
  return (
    <svg viewBox="0 0 12 12" width="10" height="10" style={{ transform: open ? "rotate(90deg)" : "rotate(0deg)", transition: "transform 0.15s ease", opacity: 0.5, flexShrink: 0 }}>
      <path d="M4 2l4 4-4 4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function AccountHeader({ username, open, onClick, online, total }) {
  const highlight = online > 0;
  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-2 px-3 text-left"
      style={{ height: 28, fontSize: 11, fontWeight: highlight ? 700 : 600, letterSpacing: "0.06em", textTransform: "uppercase", color: highlight ? "var(--sidebar-text)" : "var(--muted)", background: "transparent", border: "none", cursor: "pointer", transition: "color 0.2s" }}
      onMouseEnter={(e) => e.currentTarget.style.color = "var(--sidebar-text)"}
      onMouseLeave={(e) => e.currentTarget.style.color = highlight ? "var(--sidebar-text)" : "var(--muted)"}
    >
      <CollapseBtn open={open} />
      <span className="truncate flex-1">{username}</span>
      <span style={{ fontSize: 10, opacity: highlight ? 0.7 : 0.45, color: highlight ? "#4ade80" : "inherit" }}>{online}/{total}</span>
    </button>
  );
}

function AccountGroup({ uid, username, counts, activeId, onSelect }) {
  const { get, set: setPref } = useUIPrefs();
  const key = `sidebar.account.${uid}`;
  const open = get(key, false);
  const toggle = () => setPref(key, !open);

  const { userAgents } = useAgents();
  const agents = userAgents[uid] || [];
  const total = counts.total;
  const loadedRef = useRef(false);
  const prevUid = useRef(uid);

  if (prevUid.current !== uid) {
    prevUid.current = uid;
    loadedRef.current = false;
  }

  useEffect(() => {
    if (!open || loadedRef.current) return;
    loadedRef.current = true;
    apiListUserAgents(uid, 0, 50)
      .then((p) => {
        if (!p || !Array.isArray(p.items)) { loadedRef.current = false; return; }
        setUserAgents(uid, p.items, p.total ?? p.items.length);
      })
      .catch(() => { loadedRef.current = false; });
  }, [open, uid]);

  const loadMore = useCallback((offset) => loadMoreUserAgents(uid, offset), [uid]);
  const { hasMore, sentinelRef } = usePagination({ list: agents, total, fetchPage: loadMore });

  return (
    <div style={{ borderBottom: "1px solid var(--border)" }}>
      <AccountHeader username={username} open={open} onClick={toggle} online={counts.online} total={counts.total} />
      {open && (
        <div style={{ paddingBottom: 4 }}>
          {agents.length === 0 && loadedRef.current && (
            <p style={{ padding: "4px 12px 4px 28px", fontSize: 11, color: "var(--muted)" }}>No agents</p>
          )}
          {agents.map((a) => (
            <AgentEntry key={a.id} agent={a} active={String(a.id) === activeId} onSelect={onSelect} />
          ))}
          {hasMore && <div ref={sentinelRef} style={{ height: 1 }} />}
        </div>
      )}
    </div>
  );
}

export default function AdminAgentSidebar({ activeId, onSelect }) {
  const { perUserCounts } = useAgents();

  const sorted = Object.entries(perUserCounts)
    .sort(([, a], [, b]) => b.online - a.online || b.total - a.total);

  return (
    <>
      {sorted.map(([uid, counts]) => (
        <AccountGroup
          key={uid} uid={uid}
          username={counts.username || uid}
          counts={counts}
          activeId={activeId} onSelect={onSelect}
        />
      ))}
    </>
  );
}