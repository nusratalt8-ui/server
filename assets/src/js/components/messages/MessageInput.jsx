import { useRef, useState, useEffect } from "react";
import { uploadFile } from "../../api/upload";
import ProgressBar from "../ui/ProgressBar";
import useDragDrop from "../../hooks/useDragDrop";

export default function MessageInput({ onSend, placeholder, disableWindowDrop = false }) {
  const [text, setText] = useState("");
  const [atts, setAtts] = useState([]);
  const [uploads, setUploads] = useState([]);
  const taRef = useRef(null);
  const fileRef = useRef(null);

  useEffect(() => {
    const el = taRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 140) + "px";
  }, [text]);

  const upload = (file) => {
    if (!file) return;
    const id = Math.random().toString(36).slice(2);
    setUploads((p) => [...p, { id, name: file.name, progress: 0, error: null }]);
    uploadFile(file, (p) => {
      setUploads((prev) => prev.map((u) => u.id === id ? { ...u, progress: p * 100 } : u));
    }).then((a) => {
      setAtts((p) => [...p, { id: a.id, filename: a.filename }]);
      setUploads((prev) => prev.filter((u) => u.id !== id));
    }).catch((e) => {
      setUploads((prev) => prev.map((u) => u.id === id ? { ...u, error: e.message } : u));
      setTimeout(() => setUploads((prev) => prev.filter((u) => u.id !== id)), 3000);
    });
  };

  const uploadMany = (files) => { for (const f of files) upload(f); };

  const { dragging } = useDragDrop(uploadMany, { bindWindow: true, enabled: !disableWindowDrop });

  const onPaste = (e) => {
    const items = e.clipboardData?.items;
    if (!items) return;
    const files = [];
    for (const item of items) {
      if (item.type.startsWith("image/")) {
        const file = item.getAsFile();
        if (file) files.push(file);
      }
    }
    if (files.length > 0) { e.preventDefault(); uploadMany(files); }
  };

  const submit = () => {
    const t = text.trim();
    if (!t && atts.length === 0) return;
    onSend(t, atts.map((a) => a.id));
    setText("");
    setAtts([]);
  };

  const busy = uploads.length > 0;
  const canSend = !busy && (text.trim() || atts.length > 0);

  return (
    <div className="composer">
      {dragging && (
        <div className="fixed inset-0 z-50 flex items-center justify-center pointer-events-none" style={{ background: "rgba(0,0,0,0.35)" }}>
          <div className="bevel-out px-8 py-6 text-center font-bold" style={{ background: "var(--window-face)", border: "2px dashed var(--accent)" }}>
            Drop files to attach
          </div>
        </div>
      )}
      {uploads.map((u) => (
        <div key={u.id} style={{ marginBottom: 6 }}>
          <ProgressBar progress={u.progress} label={u.name} error={u.error} />
        </div>
      ))}
      {atts.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {atts.map((a, i) => (
            <span key={a.id} className="chip">
              <span className="truncate max-w-[140px]">{a.filename}</span>
              <button className="chip-x" onClick={() => setAtts((p) => p.filter((_, j) => j !== i))}>×</button>
            </span>
          ))}
        </div>
      )}
      <div className="composer-well">
        <input ref={fileRef} type="file" multiple className="hidden" onChange={(e) => { uploadMany(Array.from(e.target.files || [])); e.target.value = ""; }} />
        <button className="composer-attach" onClick={() => fileRef.current?.click()} disabled={busy} title="Attach file">＋</button>
        <textarea
          ref={taRef}
          rows={1}
          value={text}
          maxLength={4000}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); submit(); } }}
          onPaste={onPaste}
          placeholder={placeholder || "Type a message…"}
          className="composer-text vista-scroll"
        />
        <button className="composer-send" onClick={submit} disabled={!canSend}>Send</button>
      </div>
    </div>
  );
}