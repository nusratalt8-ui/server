import Markdown from "../markdown/Markdown";
import { ImagePreview } from "./AttachmentView";

function asText(v, key) {
  if (!v) return "";
  if (typeof v === "string") return v;
  if (typeof v === "object") return v[key] || "";
  return "";
}

function asUrl(v) {
  if (!v) return "";
  if (typeof v === "string") return v;
  if (typeof v === "object") return v.url || "";
  return "";
}

function EmbedFields({ fields }) {
  const hasInline = fields.some((f) => f.inline);
  return (
    <div
      className="grid gap-y-2 gap-x-4 mt-2"
      style={{ gridTemplateColumns: hasInline ? "repeat(3, minmax(0, 1fr))" : "minmax(0, 1fr)" }}
    >
      {fields.map((f, i) => (
        <div key={i} className={f.inline ? "min-w-0" : "col-span-full min-w-0"}>
          <div className="text-xs font-bold opacity-80 mb-0.5"><Markdown text={f.name} inline /></div>
          <div className="break-words"><Markdown text={f.value} /></div>
        </div>
      ))}
    </div>
  );
}

export default function Embed({ embed }) {
  if (!embed) return null;

  const authorText = asText(embed.author, "name");
  const footerText = asText(embed.footer, "text");
  const imageUrl = asUrl(embed.image) || asUrl(embed.thumbnail);

  return (
    <div
      className="bevel-out mt-1 p-2 text-sm max-w-[520px]"
      style={{ background: "var(--window-face)", borderLeft: "4px solid " + (embed.color || "var(--titlebar)") }}
    >
      {authorText && <div className="text-xs font-medium opacity-80 mb-1"><Markdown text={authorText} inline /></div>}
      {embed.title && <div className="font-bold mb-1"><Markdown text={embed.title} inline /></div>}
      {embed.description && <div className="break-words"><Markdown text={embed.description} /></div>}
      {embed.fields?.length > 0 && <EmbedFields fields={embed.fields} />}
      {imageUrl && <div className="mt-2"><ImagePreview url={imageUrl} /></div>}
      {footerText && <div className="mt-2 text-xs opacity-60"><Markdown text={footerText} inline /></div>}
    </div>
  );
}