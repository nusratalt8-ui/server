import { useState, useEffect } from "react";
import { useAgents } from "../../hooks/useAgents";
import { useChat } from "../../hooks/useChat";
import { MessageList, MessageInput } from "../messages";
import Toolbar from "./Toolbar";
import { apiGetAgent } from "../../api/agents";

export default function AgentView({ id }) {
  const { agents, userAgents, ready } = useAgents();
  const [remoteAgent, setRemoteAgent] = useState(null);
  const [fetched, setFetched] = useState(false);

  const allCachedAgents = Object.values(userAgents).flat();
  const localAgent = agents.find((a) => String(a.id) === String(id))
    || allCachedAgents.find((a) => String(a.id) === String(id));

  useEffect(() => {
    setRemoteAgent(null);
    setFetched(false);
  }, [id]);

  useEffect(() => {
    if (localAgent || fetched) return;
    apiGetAgent(id).then((a) => { setRemoteAgent(a); }).catch(() => {}).finally(() => setFetched(true));
  }, [id, localAgent, fetched]);

  const agent = localAgent || remoteAgent;
  const { messages, sendMessage, loadMore, hasMore } = useChat(id);

  if (!ready && !agent) return <Centered>Loading…</Centered>;
  if (ready && !agent && fetched) return <Centered>Agent {id} isn't connected.</Centered>;
  if (!agent) return <Centered>Loading…</Centered>;

  return (
    <div className="h-full flex flex-col">
      <div className="y2k-titlebar flex items-center gap-2">
        <span style={{
          display: "inline-block",
          width: 8,
          height: 8,
          borderRadius: "50%",
          flexShrink: 0,
          background: agent.online ? "#4ade80" : "var(--muted)",
          boxShadow: agent.online ? "0 0 6px #4ade80" : "none",
          transition: "background 0.3s, box-shadow 0.3s",
        }} />
        <span className="flex-1">{agent.name}</span>
      </div>
      <Toolbar key={"toolbar-" + id} agent={agent} />
      <MessageList key={"chat-" + id} messages={messages} onResend={(t) => sendMessage(t)} loadMore={loadMore} hasMore={hasMore} />
      <MessageInput key={"input-" + id} onSend={sendMessage} placeholder={`Message ${agent.name}…`} />
    </div>
  );
}

function Centered({ children }) {
  return <div className="h-full flex items-center justify-center p-4"><p className="opacity-70">{children}</p></div>;
}