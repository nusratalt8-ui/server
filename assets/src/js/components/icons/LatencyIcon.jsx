export function LatencyIcon({ className }) {
  return (
    <svg className={className} viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="0" y="0" width="16" height="16" rx="2" fill="white"/>
      <rect x="2" y="10" width="2" height="4" rx="0.5" fill="#1a1a1a"/>
      <rect x="5" y="7" width="2" height="7" rx="0.5" fill="#1a1a1a"/>
      <rect x="8" y="4" width="2" height="10" rx="0.5" fill="#1a1a1a"/>
      <rect x="11" y="2" width="2" height="12" rx="0.5" fill="#1a1a1a"/>
    </svg>
  );
}