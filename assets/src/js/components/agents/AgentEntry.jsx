import { navigate } from "../../router";
import { agentPath, ROUTES } from "../../routes";
import { useContextMenu } from "../contextmenu/ContextMenu";
import { send } from "../../socket";

export default function AgentEntry({ agent, active, onSelect }) {
  const openMenu = useContextMenu();

  const handleDelete = () => {
    if (!confirm(`Delete agent ${agent.name}?`)) return;
    send("agent_delete", { agent_id: agent.id });
    if (active) navigate(ROUTES.home);
  };

  const handleContext = (e) => {
    e.preventDefault();
    openMenu(e, [
      { label: "Open", action: () => { navigate(agentPath(agent.id)); onSelect?.(); } },
      "—",
      { label: "Copy ID", action: () => navigator.clipboard.writeText(agent.id) },
      { label: "Copy Hostname", action: () => navigator.clipboard.writeText(agent.hostname || "") },
      "—",
      { label: "Delete", danger: true, action: handleDelete },
    ]);
  };

  return (
    <button
      onClick={() => { navigate(agentPath(agent.id)); onSelect?.(); }}
      onContextMenu={handleContext}
      className="w-full text-left flex items-center gap-2"
      style={{
        padding: "5px 12px 5px 28px",
        fontSize: 12.5,
        borderLeft: active ? "2px solid var(--accent)" : "2px solid transparent",
        background: active ? "color-mix(in srgb, var(--accent) 10%, transparent)" : "transparent",
        color: active ? "var(--text)" : "var(--sidebar-text)",
        opacity: agent.online ? 1 : 0.5,
        transition: "background 0.1s, border-color 0.1s",
      }}
      onMouseEnter={(e) => { if (!active) e.currentTarget.style.background = "rgba(255,255,255,0.04)"; }}
      onMouseLeave={(e) => { if (!active) e.currentTarget.style.background = "transparent"; }}
    >
      <span
        style={{
          width: 6,
          height: 6,
          borderRadius: "50%",
          flexShrink: 0,
          background: agent.online ? "#4ade80" : "var(--muted)",
          boxShadow: agent.online ? "0 0 6px #4ade80" : "none",
          transition: "background 0.3s, box-shadow 0.3s",
        }}
      />
      <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{agent.name}</span>
    </button>
  );
}