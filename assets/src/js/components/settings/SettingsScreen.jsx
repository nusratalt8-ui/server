import { useMemo } from "react";
import { navigate } from "../../router";
import { useAuth } from "../../hooks/useAuth";
import FullscreenLayout from "../ui/FullscreenLayout";
import AgentKeyTab from "./AgentKeyTab";
import BuildTab from "./BuildTab";
import ThemesTab from "./ThemesTab";
import PlansTab from "./PlansTab";
import GuideTab from "./GuideTab";

export default function SettingsScreen({ defaultTab }) {
  const { user } = useAuth();

  const tabs = useMemo(() => [
    { id: "agentkey", label: "Agent Key", group: "panel", groupLabel: "Panel" },
    { id: "build", label: "Build", group: "panel" },
    { id: "plans", label: "Plan", group: "panel" },
    { id: "themes", label: "Themes", group: "panel" },
    { id: "guide", label: "Guide", group: "help", groupLabel: "Help" },
  ], []);

  return (
    <FullscreenLayout title="Settings" tabs={tabs} defaultTab={defaultTab} onClose={() => navigate("/")}>
      {(tab) => (
        <>
          {tab === "agentkey" && <AgentKeyTab />}
          {tab === "build" && <BuildTab />}
          {tab === "plans" && <PlansTab />}
          {tab === "themes" && <ThemesTab />}
          {tab === "guide" && <GuideTab />}
        </>
      )}
    </FullscreenLayout>
  );
}