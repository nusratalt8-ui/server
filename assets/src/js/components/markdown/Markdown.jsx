import { useState } from "react";
import { CopyIcon, CheckIcon } from "../icons";
import { tokenize } from "./tokens";

function Spoiler({ children }) {
  const [revealed, setRevealed] = useState(false);
  if (revealed) {
    return <span className="px-0.5" style={{ background: "var(--edge-light)" }}>{children}</span>;
  }
  return (
    <span
      onClick={(e) => { e.stopPropagation(); setRevealed(true); }}
      className="px-0.5 cursor-pointer select-none text-transparent"
      style={{ background: "var(--text)" }}
      title="Click to reveal"
    >
      {children}
    </span>
  );
}

function CodeBlock({ children, inline }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard?.writeText(children).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {});
  };
  if (inline) {
    return (
      <code
        className="px-1 text-[12px] font-mono bevel-in"
        style={{ background: "var(--input-bg)" }}
      >
        {children}
      </code>
    );
  }
  return (
    <div className="relative group/code my-1">
      <code
        className="block px-2 py-1 text-[12px] font-mono whitespace-pre-wrap bevel-in"
        style={{ background: "var(--input-bg)" }}
      >
        {children}
      </code>
      <button
        onClick={copy}
        className="absolute top-1 right-1 p-0.5 bevel-out opacity-0 group-hover/code:opacity-100"
        style={{ background: "var(--window-face)" }}
        title="Copy"
      >
        {copied ? <CheckIcon className="w-3.5 h-3.5" /> : <CopyIcon className="w-3.5 h-3.5" />}
      </button>
    </div>
  );
}

const HEADER_CLASSES = { 1: "block text-[20px] font-bold my-2", 2: "block text-[17px] font-bold my-1.5", 3: "block text-[15px] font-bold my-1" };
const HEADER_RE = /^(#{1,3}) +(.+)$/;
const UL_RE = /^[-*] +(.+)$/;
const OL_RE = /^(\d+)\. +(.+)$/;

function renderInline(text, keyPrefix) {
  const out = [];
  let i = 0;
  const plain = (s) => { if (s) out.push(s); };
  tokenize(text, (m) => {
    const k = `${keyPrefix}_${i++}`;
    if (m[1] !== undefined)      out.push(<CodeBlock key={k}>{m[1]}</CodeBlock>);
    else if (m[2] !== undefined) out.push(<CodeBlock key={k} inline>{m[2]}</CodeBlock>);
    else if (m[3] !== undefined) out.push(<strong key={k} className="font-bold"><em>{m[3]}</em></strong>);
    else if (m[4] !== undefined) out.push(<strong key={k} className="font-bold">{m[4]}</strong>);
    else if (m[5] !== undefined) out.push(<em key={k}>{m[5]}</em>);
    else if (m[6] !== undefined) out.push(<span key={k} className="line-through">{m[6]}</span>);
    else if (m[7] !== undefined && m[8] !== undefined) out.push(<a key={k} href={m[8]} target="_blank" rel="noopener noreferrer" className="underline" style={{ color: "var(--accent)" }}>{m[7]}</a>);
    else if (m[9] !== undefined) out.push(<a key={k} href={m[9]} target="_blank" rel="noopener noreferrer" className="underline" style={{ color: "var(--accent)" }}>{m[9]}</a>);
    else if (m[10] !== undefined) out.push(<Spoiler key={k}>{m[10]}</Spoiler>);
  }, plain);
  return out;
}

// Classify a raw line into a block kind.
function classify(line) {
  const stripped = line.trimEnd();
  if (stripped.trim() === "") return { kind: "blank" };
  const h = HEADER_RE.exec(stripped);
  if (h) return { kind: "h", level: h[1].length, text: h[2] };
  const ul = UL_RE.exec(stripped);
  if (ul) return { kind: "ul", text: ul[1] };
  const ol = OL_RE.exec(stripped);
  if (ol) return { kind: "ol", text: ol[2], start: parseInt(ol[1], 10) };
  return { kind: "p", text: stripped };
}

// Group contiguous runs so list items coalesce into one <ul>/<ol> and
// adjacent prose lines join into one paragraph with soft breaks.
function groupBlocks(lines) {
  const blocks = [];
  let i = 0;
  while (i < lines.length) {
    const c = classify(lines[i]);
    if (c.kind === "blank") { blocks.push({ kind: "blank" }); i++; continue; }
    if (c.kind === "h")     { blocks.push(c); i++; continue; }
    if (c.kind === "ul" || c.kind === "ol") {
      const items = [c.text];
      const listKind = c.kind;
      const start = c.start;
      i++;
      while (i < lines.length) {
        const n = classify(lines[i]);
        if (n.kind !== listKind) break;
        items.push(n.text);
        i++;
      }
      blocks.push({ kind: listKind, items, start });
      continue;
    }
    const para = [c.text];
    i++;
    while (i < lines.length) {
      const n = classify(lines[i]);
      if (n.kind !== "p") break;
      para.push(n.text);
      i++;
    }
    blocks.push({ kind: "p", lines: para });
  }
  return blocks;
}

function splitFences(text) {
  const FENCE_RE = /```([\s\S]*?)```/g;
  const segs = [];
  let last = 0;
  let m;
  while ((m = FENCE_RE.exec(text)) !== null) {
    if (m.index > last) segs.push({ kind: "text", body: text.slice(last, m.index) });
    segs.push({ kind: "fence", body: m[1].replace(/^\n|\n$/g, "") });
    last = FENCE_RE.lastIndex;
  }
  if (last < text.length) segs.push({ kind: "text", body: text.slice(last) });
  return segs.length ? segs : [{ kind: "text", body: text }];
}

export default function Markdown({ text, inline }) {
  if (!text) return null;
  const segs = splitFences(text);
  const out = [];

  segs.forEach((seg, si) => {
    if (seg.kind === "fence") {
      out.push(<CodeBlock key={`s${si}f`} inline={inline}>{seg.body}</CodeBlock>);
      return;
    }
    if (inline) {
      const lines = seg.body.split("\n");
      lines.forEach((ln, li) => {
        if (li > 0 || (si > 0 && li === 0)) out.push(<br key={`s${si}_ibr${li}`} />);
        out.push(...renderInline(ln, `s${si}_i${li}`));
      });
      return;
    }
    const blocks = groupBlocks(seg.body.split("\n"));
    blocks.forEach((b, bi) => {
      if (b.kind === "blank") return;
      const k = `s${si}_b${bi}`;
      if (b.kind === "h") {
        const Tag = b.level === 1 ? "h1" : b.level === 2 ? "h2" : "h3";
        out.push(<Tag key={k} className={HEADER_CLASSES[b.level]}>{renderInline(b.text, k)}</Tag>);
        return;
      }
      if (b.kind === "ul") {
        out.push(
          <ul key={k} className="list-disc pl-5 my-1 space-y-0.5">
            {b.items.map((t, ii) => <li key={ii}>{renderInline(t, `${k}_${ii}`)}</li>)}
          </ul>
        );
        return;
      }
      if (b.kind === "ol") {
        out.push(
          <ol key={k} start={b.start} className="list-decimal pl-5 my-1 space-y-0.5">
            {b.items.map((t, ii) => <li key={ii}>{renderInline(t, `${k}_${ii}`)}</li>)}
          </ol>
        );
        return;
      }
      const children = [];
      b.lines.forEach((ln, li) => {
        if (li > 0) children.push(<br key={`${k}_br${li}`} />);
        children.push(...renderInline(ln, `${k}_${li}`));
      });
      out.push(<p key={k} className="my-0.5 break-words">{children}</p>);
    });
  });
  return <>{out}</>;
}