import { useEffect, useState } from "react";
import { apiAgentKeyInfo, apiRotateAgentKey } from "../../api/agentkeys";
import Button from "../ui/Button";

export default function AgentKeyTab() {
  const [info, setInfo] = useState(null);
  const [newKey, setNewKey] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    try { setInfo(await apiAgentKeyInfo()); } catch (e) { setError(e.message); }
  }
  useEffect(() => { load(); }, []);

  async function rotate() {
    setError("");
    setBusy(true);
    try {
      const res = await apiRotateAgentKey();
      setNewKey(res.key);
      await load();
    } catch (e) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <p className="font-bold mb-2">Agent Key</p>
      {info && (
        <p className="mb-3 opacity-80">
          {info.exists ? "A key is active." : "No key yet — rotate to create one."}
        </p>
      )}
      {newKey && (
        <div className="bevel-in p-2 mb-3" style={{ background: "var(--input-bg)" }}>
          <p className="text-xs opacity-70 mb-1">Copy this now — it won't be shown again:</p>
          <code className="block break-all text-sm">{newKey}</code>
          <Button className="mt-2" onClick={() => navigator.clipboard?.writeText(newKey)}>Copy</Button>
        </div>
      )}
      {error && <p className="y2k-error bevel-in px-2 py-1 mb-3">{error}</p>}
      <Button onClick={rotate} disabled={busy}>{busy ? "..." : "Rotate Key"}</Button>
      <p className="text-xs opacity-60 mt-3">Rotating invalidates the old key and disconnects any agent still using it.</p>
    </div>
  );
}