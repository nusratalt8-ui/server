const IMG = ["image/"];
const TEXT_TYPES = ["text/", "application/json", "application/xml"];
const TEXT_EXT = ["txt", "log", "json", "xml", "csv", "md", "yaml", "yml", "ini", "cfg", "go", "js", "c", "h", "py", "sh"];

const EXT_CAT = {
  doc: ["pdf", "doc", "docx", "rtf", "odt"],
  sheet: ["xls", "xlsx", "csv", "ods"],
  archive: ["zip", "tar", "gz", "rar", "7z", "xz"],
  slide: ["ppt", "pptx"],
  exe: ["exe", "msi", "dll", "bin"],
  image: ["png", "jpg", "jpeg", "gif", "bmp", "webp"],
};

export function extOf(name) {
  return (name || "").split(".").pop()?.toLowerCase() || "";
}

export function isImage(att) {
  if (att.content_type && IMG.some((p) => att.content_type.startsWith(p))) return true;
  return EXT_CAT.image.includes(extOf(att.filename));
}

export function isText(att) {
  if (att.content_type && TEXT_TYPES.some((p) => att.content_type.startsWith(p))) return true;
  return TEXT_EXT.includes(extOf(att.filename));
}

export function category(att) {
  const ext = extOf(att.filename);
  for (const [cat, exts] of Object.entries(EXT_CAT)) {
    if (exts.includes(ext)) return cat;
  }
  return "generic";
}

export function formatSize(bytes) {
  const b = Number(bytes) || 0;
  if (b < 1024) return b + " B";
  if (b < 1024 * 1024) return (b / 1024).toFixed(1) + " KB";
  return (b / (1024 * 1024)).toFixed(1) + " MB";
}