import { useState } from "react";
import { CodeLines } from "../messages/FilePreview";
import { DownloadIcon, ExpandIcon, ChevronDownIcon, ChevronUpIcon, CloseIcon } from "../icons";

const COLLAPSED = 6;
const EXPANDED = 30;

function downloadText(data, filename) {
  const blob = new Blob([data], { type: "application/octet-stream" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename || "export.json";
  a.click();
  URL.revokeObjectURL(url);
}

function Fullscreen({ data, filename, lines, onClose }) {
  return (
    <div className="file-modal-backdrop" onClick={onClose}>
      <div className="file-modal y2k-window" onClick={(e) => e.stopPropagation()}>
        <div className="y2k-titlebar flex items-center gap-2">
          <span className="flex-1 truncate">{filename}</span>
          <span className="file-card-size">{lines.length} lines</span>
          <button className="file-foot-btn" title="save" onClick={() => downloadText(data, filename)}><DownloadIcon /></button>
          <button className="file-foot-btn" title="close" onClick={onClose}><CloseIcon /></button>
        </div>
        <div className="file-modal-body vista-scroll">
          <CodeLines lines={lines} />
        </div>
      </div>
    </div>
  );
}

export default function DataPreview({ data, filename }) {
  const [expanded, setExpanded] = useState(false);
  const [full, setFull] = useState(false);

  let display = data;
  try {
    const parsed = JSON.parse(data);
    display = JSON.stringify(parsed, null, 2);
  } catch {}

  const lines = display.replace(/\n$/, "").split("\n");
  const limit = expanded ? EXPANDED : COLLAPSED;
  const hasMore = lines.length > limit;

  return (
    <div className="file-card mt-1">
      <div className="file-text-wrap">
        <CodeLines lines={lines} limit={limit} />
      </div>
      <div className="file-foot">
        {hasMore && (
          <button className="file-foot-btn" title={expanded ? "collapse" : "expand"} onClick={() => setExpanded(v => !v)}>
            {expanded ? <ChevronUpIcon /> : <ChevronDownIcon />}
          </button>
        )}
        <span className="file-foot-meta min-w-0">
          <span className="block truncate">{filename}</span>
          <span className="file-card-size">{lines.length} lines</span>
        </span>
        <button className="file-foot-btn" title="view whole" onClick={() => setFull(true)}><ExpandIcon /></button>
        <button className="file-foot-btn" title="save" onClick={() => downloadText(data, filename)}><DownloadIcon /></button>
      </div>
      {full && (
        <Fullscreen data={data} filename={filename} lines={lines} onClose={() => setFull(false)} />
      )}
    </div>
  );
}