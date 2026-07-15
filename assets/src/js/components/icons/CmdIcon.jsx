export function CmdIcon({ className = "w-5 h-5" }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
      <rect x="1" y="2" width="22" height="20" rx="2" fill="#0c0c0c" stroke="#4a4a4a" strokeWidth="1" />
      <rect x="1" y="2" width="22" height="4" rx="2" fill="#2a2a2a" />
      <path d="M4 10l3 2.5L4 15" fill="none" stroke="#ffffff" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M9 15.5h5" stroke="#ffffff" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  );
}