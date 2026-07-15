import AgentEntry from "./AgentEntry";
import AdminAgentSidebar from "./AdminAgentSidebar";
import { useAgents, loadMoreAgents } from "../../hooks/useAgents";
import { usePagination } from "../../hooks/usePagination";
import { useRouter } from "../../router";
import { navigate } from "../../router";
import { ROUTES } from "../../routes";
import Button from "../ui/Button";

export default function AgentSidebar({ user, onSelect, onLogout }) {
  const { agents, agentsTotal, agentsOnline, adminTotal, adminOnline, perUserCounts, ready } = useAgents();
  const { match } = useRouter();
  const { hasMore, sentinelRef } = usePagination({ list: agents, total: agentsTotal, fetchPage: loadMoreAgents });
  const activeId = match("/agents/:id")?.id ?? null;
  const go = (to) => { navigate(to); onSelect?.(); };

  const isEmpty = ready && Object.keys(perUserCounts).length === 0 && agents.length === 0;

  return (
    <div className="y2k-sidebar h-full flex flex-col">
      <div style={{ padding: "14px 16px 10px", fontSize: 11, fontWeight: 700, letterSpacing: "0.08em", textTransform: "uppercase", color: "var(--muted)", flexShrink: 0 }}>
        {user?.is_admin
          ? <span>Agents <span style={{ fontWeight: 400, opacity: 0.5 }}>{adminOnline}/{adminTotal}</span></span>
          : <span>Agents <span style={{ fontWeight: 400, opacity: 0.5 }}>{agentsOnline}/{agentsTotal}</span></span>
        }
      </div>

      <div className="flex-1 overflow-y-auto vista-scroll" style={{ paddingBottom: 4 }}>
        {!ready ? (
          <p style={{ padding: "24px 16px", textAlign: "center", fontSize: 12, color: "var(--muted)" }}>Loading…</p>
        ) : isEmpty ? (
          <p style={{ padding: "24px 16px", textAlign: "center", fontSize: 12, color: "var(--muted)" }}>No agents connected</p>
        ) : user?.is_admin ? (
          <AdminAgentSidebar activeId={activeId} onSelect={onSelect} />
        ) : (
          <>
            {agents.map((a) => (
              <AgentEntry key={a.id} agent={a} active={String(a.id) === activeId} onSelect={onSelect} />
            ))}
            {hasMore && <div ref={sentinelRef} style={{ height: 1 }} />}
          </>
        )}
      </div>

      <div style={{ padding: "8px", borderTop: "1px solid var(--border)", display: "flex", flexDirection: "column", gap: 4 }}>
        {user?.is_admin && (
          <Button className="w-full" onClick={() => go(ROUTES.admin)}>Admin</Button>
        )}
        <Button className="w-full" onClick={() => go(ROUTES.settings)}>Settings</Button>
        <Button className="w-full" onClick={onLogout}>Log Out</Button>
      </div>
    </div>
  );
}