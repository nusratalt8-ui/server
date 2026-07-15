import { useState, useEffect, useRef } from "react";
import { on } from "../../socket/events";
import { send } from "../../socket";
import { getAppState, setAppState } from "../../appstate";

function ShellPane({ agent, shellKey, active }) {
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [history, setHistory] = useState([]);
  const [histIdx, setHistIdx] = useState(-1);
  const [cwd, setCwd] = useState("C:\\");
  const cwdRef = useRef("C:\\");
  const outputRef = useRef(null);
  const bottomRef = useRef(null);
  const inputRef = useRef(null);
  const busyRef = useRef(false);

  const prompt = `${agent.username || "user"}@${agent.hostname || agent.name}:${cwd}${shellKey === "ps" ? " PS>" : ">"}`;

  const appendLine = (text, color) => {
    const el = document.createElement("div");
    el.style.cssText = `color:${color};white-space:pre-wrap;word-break:break-all`;
    el.textContent = text;
    outputRef.current?.appendChild(el);
    bottomRef.current?.scrollIntoView({ behavior: "auto" });
  };

  useEffect(() => {
    if (active) setTimeout(() => inputRef.current?.focus(), 0);
  }, [active]);

  useEffect(() => {
    const event = shellKey === "ps" ? "terminal_ps_output" : "terminal_output";
    return on(event, (p) => {
      if (!p || String(p.agent_id) !== String(agent.id)) return;
      let out = "(no output)";
      try { out = atob(p.output); } catch {}
      if (p.cwd) { setCwd(p.cwd); cwdRef.current = p.cwd; }
      appendLine(out, p.exit === 0 ? "#cccccc" : "#f47c7c");
      busyRef.current = false;
      setBusy(false);
      setTimeout(() => inputRef.current?.focus(), 0);
    });
  }, [agent.id, shellKey]);

  const submit = () => {
    const cmd = input.trim();
    if (!cmd || busyRef.current) return;
    appendLine(`${agent.username || "user"}@${agent.hostname || agent.name}:${cwdRef.current}${shellKey === "ps" ? " PS>" : ">"} ${cmd}`, "#ffffff");
    setHistory(h => [cmd, ...h.slice(0, 99)]);
    setHistIdx(-1);
    setInput("");
    busyRef.current = true;
    setBusy(true);
    const event = shellKey === "ps" ? "terminal_ps_exec" : "terminal_exec";
    send(event, { agent_id: String(agent.id), cmd });
  };

  const onKeyDown = (e) => {
    if (e.key === "Enter") { e.preventDefault(); submit(); return; }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      const idx = Math.min(histIdx + 1, history.length - 1);
      setHistIdx(idx);
      setInput(history[idx] ?? "");
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      const idx = Math.max(histIdx - 1, -1);
      setHistIdx(idx);
      setInput(idx === -1 ? "" : (history[idx] ?? ""));
    }
  };

  return (
    <div style={{ display: active ? "flex" : "none", flexDirection: "column", flex: 1, minHeight: 0 }}>
      <div
        ref={outputRef}
        className="flex-1 overflow-y-auto p-3 vista-scroll"
        style={{ scrollbarColor: "#444 #0c0c0c", flex: 1, minHeight: 0 }}
        onClick={() => inputRef.current?.focus()}
      >
        {busy && <div style={{ color: "#888" }}>running…</div>}
        <div ref={bottomRef} />
      </div>
      <div className="flex items-center gap-1 px-3 py-2 border-t" style={{ borderColor: "#333", background: "#0c0c0c", flexShrink: 0 }}>
        <span style={{ color: shellKey === "ps" ? "#569cd6" : "#4ec9b0", flexShrink: 0, fontFamily: "monospace", fontSize: 12 }}>{prompt}</span>
        <input
          ref={inputRef}
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={onKeyDown}
          className="flex-1 bg-transparent outline-none text-xs font-mono"
          style={{ color: "#ffffff", caretColor: "#ffffff" }}
          spellCheck={false}
          autoComplete="off"
        />
      </div>
    </div>
  );
}

export default function Terminal({ agent, activeTab }) {
  const shell = activeTab || "cmd";

  return (
    <div className="h-full flex flex-col font-mono text-xs" style={{ background: "#0c0c0c", color: "#cccccc" }}>
      <ShellPane key="cmd" agent={agent} shellKey="cmd" active={shell === "cmd"} />
      <ShellPane key="ps" agent={agent} shellKey="ps" active={shell === "ps"} />
    </div>
  );
}