export function ControlPanelIcon({ className = "w-5 h-5" }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none">
      <rect x="2" y="2" width="9" height="9" rx="1.5" fill="rgba(100,160,255,0.3)" stroke="rgba(100,160,255,0.7)" strokeWidth="1.2"/>
      <rect x="13" y="2" width="9" height="9" rx="1.5" fill="rgba(255,120,80,0.3)" stroke="rgba(255,120,80,0.7)" strokeWidth="1.2"/>
      <rect x="2" y="13" width="9" height="9" rx="1.5" fill="rgba(80,200,100,0.3)" stroke="rgba(80,200,100,0.7)" strokeWidth="1.2"/>
      <rect x="13" y="13" width="9" height="9" rx="1.5" fill="rgba(255,180,50,0.3)" stroke="rgba(255,180,50,0.7)" strokeWidth="1.2"/>
    </svg>
  );
}