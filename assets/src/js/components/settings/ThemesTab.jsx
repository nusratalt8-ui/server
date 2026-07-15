import { THEMES } from "../../themes";
import { useUIPrefs } from "../../hooks/useUIPrefs";
import ThemePicker from "../theme/ThemePicker";

export default function ThemesTab() {
  const { get, set } = useUIPrefs();
  const active = get("theme", "dark");
  return <ThemePicker themes={THEMES} activeId={active} onSelect={(id) => set("theme", id)} />;
}