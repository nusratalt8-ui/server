import { useEffect, useState } from "react";
import { formatSize, isText, category } from "../../utils/attachments";
import {
  DocumentIcon, CodeIcon, SheetIcon, ArchiveIcon, ImageIcon, ExeIcon,
  DownloadIcon, ExpandIcon, ChevronDownIcon, ChevronUpIcon, CloseIcon,
} from "../icons";

const COLLAPSED = 4;
const EXPANDED = 27;

export function CatIcon({ cat, className = "w-5 h-5" }) {
  switch (cat) {
    case "code": return <CodeIcon className={className} />;
    case "sheet": return <SheetIcon className={className} />;
    case "archive": return <ArchiveIcon className={className} />;
    case "image": return <ImageIcon className={className} />;
    case "exe": return <ExeIcon className={className} />;
    default: return <DocumentIcon className={className} />;
  }
}

function download(url, filename) {
  fetch(url, { credentials: "same-origin" })
    .then((r) => r.blob())
    .then((b) => {
      const u = URL.createObjectURL(b);
      const a = document.createElement("a");
      a.href = u;
      a.download = filename || "file";
      a.click();
      URL.revokeObjectURL(u);
    })
    .catch(() => {});
}

export function CodeLines({ lines, limit }) {
  const shown = typeof limit === "number" ? lines.slice(0, limit) : lines;
  return (
    <pre className="file-text vista-scroll">
      {shown.map((ln, i) => (
        <div key={i}><span className="file-ln">{i + 1}</span>{ln || "\u00A0"}</div>
      ))}
    </pre>
  );
}

function Fullscreen({ filename, size, lines, url, onClose }) {
  return (
    <div className="file-modal-backdrop" onClick={onClose}>
      <div className="file-modal y2k-window" onClick={(e) => e.stopPropagation()}>
        <div className="y2k-titlebar flex items-center gap-2">
          <span className="flex-1 truncate">{filename || "file"}</span>
          <span className="file-card-size">{formatSize(size)} · {lines.length} lines</span>
          <button className="file-foot-btn" title="save" onClick={() => download(url, filename)}><DownloadIcon /></button>
          <button className="file-foot-btn" title="close" onClick={onClose}><CloseIcon /></button>
        </div>
        <div className="file-modal-body vista-scroll">
          <CodeLines lines={lines} />
        </div>
      </div>
    </div>
  );
}

export default function FilePreview({ att, url }) {
  const text = isText(att);
  const [body, setBody] = useState(null);
  const [expanded, setExpanded] = useState(false);
  const [full, setFull] = useState(false);

  useEffect(() => {
    if (!text) return;
    let live = true;
    fetch(url, { credentials: "same-origin" })
      .then((r) => (r.ok ? r.text() : ""))
      .then((t) => { if (live) setBody(t); })
      .catch(() => {});
    return () => { live = false; };
  }, [url, text]);

  if (!text) {
    return (
      <div className="file-pill mt-1">
        <span className="file-icon"><CatIcon cat={category(att)} /></span>
        <span className="file-pill-meta min-w-0">
          <span className="block truncate">{att.filename || "attachment"}</span>
          <span className="file-card-size">{formatSize(att.size)}</span>
        </span>
        <button className="file-foot-btn" title="save" onClick={() => download(url, att.filename)}><DownloadIcon /></button>
      </div>
    );
  }

  const lines = (body || "").replace(/\n$/, "").split("\n");
  const limit = expanded ? EXPANDED : COLLAPSED;
  const hasMore = lines.length > limit;

  return (
    <div className="file-card mt-1">
      <div className="file-text-wrap">
        {body === null ? <div className="file-loading">loading…</div> : <CodeLines lines={lines} limit={limit} />}
      </div>
      <div className="file-foot">
        {hasMore && (
          <button className="file-foot-btn" title={expanded ? "collapse" : "expand"} onClick={() => setExpanded((v) => !v)}>
            {expanded ? <ChevronUpIcon /> : <ChevronDownIcon />}
          </button>
        )}
        <span className="file-foot-meta min-w-0">
          <span className="block truncate">{att.filename || "file"}</span>
          <span className="file-card-size">{formatSize(att.size)}{lines.length ? ` · ${lines.length} lines` : ""}</span>
        </span>
        {body !== null && (
          <button className="file-foot-btn" title="view whole" onClick={() => setFull(true)}><ExpandIcon /></button>
        )}
        <button className="file-foot-btn" title="save" onClick={() => download(url, att.filename)}><DownloadIcon /></button>
      </div>
      {full && body !== null && (
        <Fullscreen filename={att.filename} size={att.size} lines={lines} url={url} onClose={() => setFull(false)} />
      )}
    </div>
  );
}