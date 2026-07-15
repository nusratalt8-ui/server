import { useState } from "react";
import AgentSidebar from "../agents/AgentSidebar";
import { MenuIcon } from "../icons";

export default function AppLayout({ children, user, onLogout }) {
  const [open, setOpen] = useState(false);

  return (
    <div className="h-screen flex overflow-hidden">
      {open && (
        <div className="fixed inset-0 z-30 bg-black/40 md:hidden" onClick={() => setOpen(false)} />
      )}

      <aside
        className={
          "fixed inset-y-0 left-0 z-40 w-[260px] transition-transform duration-150 md:static md:translate-x-0 " +
          (open ? "translate-x-0" : "-translate-x-full")
        }
      >
        <AgentSidebar user={user} onSelect={() => setOpen(false)} onLogout={onLogout} />
      </aside>

      <div className="flex-1 flex flex-col min-w-0">
        <div className="md:hidden y2k-titlebar flex items-center gap-2">
          <button
            onClick={() => setOpen(true)}
            aria-label="Open agents"
            className="bevel-out active:bevel-in"
            style={{ background: "var(--window-face)", color: "var(--text)", padding: "4px 8px" }}
          >
            <MenuIcon className="w-5 h-5" />
          </button>
          <span>Agents</span>
        </div>
        <main className="flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
}