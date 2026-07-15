// Inline markdown tokenizer (ported from banterchat, stripped of
// emoji/timestamp tokens which this app has no infrastructure for).
// Groups: 1 fence, 2 inline code, 3 bold+italic, 4 bold, 5 italic,
// 6 strikethrough, 7 link text, 8 link url, 9 raw url, 10 spoiler.
export const INLINE_RE = /```([\s\S]*?)```|`([^`\n]+)`|\*\*\*(.+?)\*\*\*|\*\*(.+?)\*\*|\*(.+?)\*|~~(.+?)~~|\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)|(https?:\/\/[^\s<]+)|\|\|([\s\S]+?)\|\|/g;

// Visitor-style tokenizer: `visit(match)` per regex match, `plain(text)`
// for the runs of plain text between matches.
export function tokenize(text, visit, plain) {
  if (!text) return;
  INLINE_RE.lastIndex = 0;
  let last = 0;
  let match;
  while ((match = INLINE_RE.exec(text)) !== null) {
    if (match.index > last) plain(text.slice(last, match.index));
    visit(match);
    last = INLINE_RE.lastIndex;
  }
  if (last < text.length) plain(text.slice(last));
}