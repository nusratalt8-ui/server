import { CheckIcon } from "../icons";

function Swatch({ theme, active, onSelect }) {
  const v = theme.vars || {};
  return (
    <button type="button" onClick={onSelect} title={theme.name} className="flex flex-col items-center gap-2">
      <div
        className="relative w-[120px] h-[80px] overflow-hidden bevel-out"
        style={{ outline: active ? "2px solid var(--accent)" : "none", outlineOffset: 2 }}
      >
        <div className="absolute inset-0" style={{ background: v["window-face"] }} />
        <div className="absolute left-0 top-0 bottom-0 w-[30px]" style={{ background: v["sidebar"] }} />
        <div className="absolute left-[30px] right-0 top-0 h-[14px]" style={{ background: v["titlebar"] }} />
        <div className="absolute left-[38px] top-[24px] h-[4px] w-[46px] rounded" style={{ background: v["text"], opacity: 0.5 }} />
        <div className="absolute left-[38px] top-[34px] h-[4px] w-[60px] rounded" style={{ background: v["text"], opacity: 0.3 }} />
        <div className="absolute left-[38px] top-[44px] h-[4px] w-[34px] rounded" style={{ background: v["text"], opacity: 0.3 }} />
        <div className="absolute right-2 bottom-2 h-[10px] w-[24px] rounded" style={{ background: v["accent"] }} />
        {active && (
          <div
            className="absolute top-1 right-1 w-5 h-5 rounded-full flex items-center justify-center"
            style={{ background: "var(--accent)" }}
          >
            <CheckIcon className="w-3 h-3" />
          </div>
        )}
      </div>
      <span className="text-xs" style={{ opacity: active ? 1 : 0.6 }}>{theme.name}</span>
    </button>
  );
}

export default function ThemePicker({ themes, activeId, onSelect }) {
  return (
    <div>
      <p className="font-bold mb-1">Theme</p>
      <p className="text-xs opacity-60 mb-4">Saved to your account.</p>
      <div className="flex flex-wrap gap-4">
        {themes.map((t) => (
          <Swatch key={t.id} theme={t} active={activeId === t.id} onSelect={() => onSelect(t.id)} />
        ))}
      </div>
    </div>
  );
}