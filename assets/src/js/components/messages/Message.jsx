import { useContextMenu } from "../contextmenu/ContextMenu";
import Embed from "./Embed";
import AttachmentView from "./AttachmentView";
import Markdown from "../markdown/Markdown";

export default function Message({ msg, onResend }) {
  const openMenu = useContextMenu();
  const mine = msg.sender === "panel";
  const items = [
    { label: "Copy", action: () => navigator.clipboard?.writeText(msg.body || "") },
    { label: "Resend", action: () => onResend(msg.body || "") },
  ];
  return (
    <div className={"flex " + (mine ? "justify-end" : "justify-start")}>
      <div className="max-w-[80%]" onContextMenu={(e) => openMenu(e, items)}>
        {msg.body && (
          <div
            className="bevel-out px-2 py-1 break-words"
            style={{ background: mine ? "var(--titlebar)" : "var(--window-face)", color: mine ? "var(--titlebar-text)" : "var(--text)" }}
          >
            <Markdown text={msg.body} />
          </div>
        )}
        {msg.embed && <Embed embed={msg.embed} />}
        <AttachmentView attachments={msg.attachments} />
      </div>
    </div>
  );
}