export function TaskMgrIcon({ className = "" }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none">
      <rect x="1" y="2" width="22" height="20" rx="1.5" fill="#ece9d8" stroke="#5a7edc" strokeWidth="1" />
      <rect x="1" y="2" width="22" height="4.5" rx="1.5" fill="#2a5fce" />
      <rect x="1" y="5" width="22" height="1.5" fill="#2a5fce" />
      <circle cx="20.5" cy="4.2" r="1" fill="#d44" />
      <rect x="3" y="8.5" width="18" height="11" fill="#000000" stroke="#5a8a5a" strokeWidth="0.5" />
      <path d="M3 12h18M3 15.5h18M7.5 8.5v11M12 8.5v11M16.5 8.5v11" stroke="#1d3d1d" strokeWidth="0.5" />
      <polyline points="3.5,17.5 6.5,13.5 9,15.5 12,10.5 14.5,14.5 17.5,10 20.5,13" stroke="#00d400" strokeWidth="1.4" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  );
}