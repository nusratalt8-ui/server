import { useState, useEffect, useCallback } from "react";
import { on } from "../../socket/events";
import { send } from "../../socket";
import { useContextMenu } from "../contextmenu/ContextMenu";
import Button from "../ui/Button";

export default function ProcessManager({ agent }) {
  const [procs, setProcs] = useState([]);
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [killing, setKilling] = useState(null);
  const [expanded, setExpanded] = useState({});
  const [selected, setSelected] = useState(null);
  const [payloads, setPayloads] = useState([]);
  const openMenu = useContextMenu();

  const parseProcs = useCallback((raw) => {
    const lines = (raw || "").trim().split("\n").filter(Boolean);
    return lines.map((l) => {
      const [pid, name, mem] = l.split("|");
      return { pid: parseInt(pid), name, mem: parseInt(mem) };
    }).filter((p) => p.pid && p.name);
  }, []);

  useEffect(() => {
    send("procwatch_start", { agent_id: String(agent.id) });
    return () => send("procwatch_stop", { agent_id: String(agent.id) });
  }, [agent.id]);

  useEffect(() => {
    return on("task_update", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      setProcs(parseProcs(p.procs));
      setLoading(false);
    });
  }, [agent.id, parseProcs]);

  useEffect(() => {
    return on("proclist_result", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      setProcs(parseProcs(p.procs));
      setLoading(false);
    });
  }, [agent.id, parseProcs]);

  useEffect(() => {
    return on("proc_kill_result", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      setKilling(null);
    });
  }, [agent.id]);

  useEffect(() => {
    send("payload_list_get", { agent_id: String(agent.id) });
    return on("payload_list_result", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      try { setPayloads(JSON.parse(p.payloads || "[]")); }
      catch { setPayloads([]); }
    });
  }, [agent.id]);

  useEffect(() => {
    return on("payload_inject_result", (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
    });
  }, [agent.id]);

  const kill = (pid) => {
    setKilling(pid);
    send("proc_kill", { agent_id: String(agent.id), data: { pid } });
  };

  const killGroup = (name) => {
    grouped[name]?.forEach((p) => kill(p.pid));
  };

  const killSelected = () => {
    if (!selected) return;
    if (selected.isGroup) killGroup(selected.name);
    else kill(selected.pid);
  };

  const injectPayload = (pid, name) => {
    send("payload_inject", { agent_id: String(agent.id), pid, name });
  };

  const grouped = procs
    .filter((p) => !filter || p.name.toLowerCase().includes(filter.toLowerCase()))
    .reduce((acc, p) => {
      if (!acc[p.name]) acc[p.name] = [];
      acc[p.name].push(p);
      return acc;
    }, {});

  const groupNames = Object.keys(grouped).sort((a, b) => {
    const memA = grouped[a].reduce((s, p) => s + p.mem, 0);
    const memB = grouped[b].reduce((s, p) => s + p.mem, 0);
    return memB - memA;
  });

  const onRowCtx = (e, pid, name, isGroup) => {
    setSelected({ pid, name, isGroup });
    const payloadSub = payloads.length > 0 ? [{
      label: "Inject Payload",
      submenu: payloads.map((pl) => ({
        label: `${pl.name} — ${pl.description}`,
        action: () => injectPayload(pid || grouped[name]?.[0]?.pid, pl.name),
      })),
    }] : [];
    openMenu(e, isGroup ? [
      { label: `End all ${name} (${grouped[name]?.length})`, danger: true, action: () => killGroup(name) },
      ...payloadSub,
    ] : [
      { label: `End Task (PID ${pid})`, danger: true, action: () => kill(pid) },
      "—",
      { label: `End all ${name}`, danger: true, action: () => killGroup(name) },
      ...payloadSub,
    ]);
  };

  const canKill = selected && killing === null;

  return (
    <div className="h-full flex flex-col text-xs font-mono" style={{ background: "var(--window-face)" }}>
      <div className="flex items-center gap-2 px-2 py-1 border-b" style={{ borderColor: "var(--edge-dark)" }}>
        <input
          className="flex-1 bevel-in px-1 py-0.5 text-xs"
          style={{ background: "var(--input-bg)", color: "var(--text)", outline: "none" }}
          placeholder="Filter processes…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <Button
          color="red"
          disabled={!canKill}
          onClick={killSelected}
        >
          {selected?.isGroup ? `End Group` : "End Task"}
        </Button>
        <span style={{ color: "var(--muted)", fontSize: 10, paddingLeft: 4 }}>
          {loading ? "connecting…" : "● live"}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto vista-scroll">
        <table className="w-full border-collapse">
          <thead className="sticky top-0" style={{ background: "var(--window-face)" }}>
            <tr style={{ borderBottom: "1px solid var(--edge-dark)" }}>
              <th className="text-left px-2 py-1 font-bold" style={{ color: "var(--muted)" }}>Name</th>
              <th className="text-right px-2 py-1 font-bold" style={{ color: "var(--muted)" }}>Count</th>
              <th className="text-right px-2 py-1 font-bold" style={{ color: "var(--muted)" }}>Mem (KB)</th>
            </tr>
          </thead>
          <tbody>
            {groupNames.map((name) => {
              const group = grouped[name];
              const totalMem = group.reduce((s, p) => s + p.mem, 0);
              const isExpanded = expanded[name];
              const isGroupSelected = selected?.name === name && selected?.isGroup;
              return [
                <tr
                  key={name}
                  onContextMenu={(e) => onRowCtx(e, null, name, true)}
                  onClick={(e) => { e.stopPropagation(); setSelected({ name, isGroup: true }); setExpanded(x => ({ ...x, [name]: !x[name] })); }}
                  style={{
                    borderBottom: "1px solid var(--edge-dark)",
                    background: isGroupSelected ? "var(--titlebar)" : undefined,
                    color: isGroupSelected ? "var(--titlebar-text)" : "var(--text)",
                    cursor: "pointer",
                  }}
                >
                  <td className="px-2 py-0.5 font-bold">
                    <span style={{ opacity: 0.5, marginRight: 4 }}>{isExpanded ? "▾" : "▸"}</span>
                    {name}
                  </td>
                  <td className="px-2 py-0.5 text-right" style={{ color: "var(--muted)" }}>{group.length}</td>
                  <td className="px-2 py-0.5 text-right" style={{ color: "var(--muted)" }}>{totalMem.toLocaleString()}</td>
                </tr>,
                ...(isExpanded ? group.map((p) => {
                  const isSelected = selected?.pid === p.pid && !selected?.isGroup;
                  return (
                    <tr
                      key={p.pid}
                      onContextMenu={(e) => onRowCtx(e, p.pid, p.name, false)}
                      onClick={(e) => { e.stopPropagation(); setSelected({ pid: p.pid, name: p.name, isGroup: false }); }}
                      style={{
                        borderBottom: "1px solid var(--edge-dark)",
                        opacity: killing === p.pid ? 0.4 : 1,
                        background: isSelected ? "var(--titlebar)" : "var(--input-bg)",
                        color: isSelected ? "var(--titlebar-text)" : "var(--text)",
                        cursor: "pointer",
                      }}
                    >
                      <td className="px-4 py-0.5" style={{ color: "var(--muted)" }}>↳ PID {p.pid}</td>
                      <td className="px-2 py-0.5 text-right" style={{ color: "var(--muted)" }}>—</td>
                      <td className="px-2 py-0.5 text-right" style={{ color: "var(--muted)" }}>{p.mem.toLocaleString()}</td>
                    </tr>
                  );
                }) : []),
              ];
            })}
          </tbody>
        </table>
        {!loading && groupNames.length === 0 && (
          <div className="p-4 text-center" style={{ color: "var(--muted)" }}>No processes found</div>
        )}
      </div>

      <div className="px-2 py-1 border-t flex gap-3" style={{ borderColor: "var(--edge-dark)", color: "var(--muted)" }}>
        <span>{procs.length} processes</span>
        <span>{groupNames.length} unique</span>
        {selected && (
          <span style={{ color: "var(--text)" }}>
            {selected.isGroup ? `${selected.name} (${grouped[selected.name]?.length ?? 0})` : `PID ${selected.pid}`}
          </span>
        )}
      </div>
    </div>
  );
}