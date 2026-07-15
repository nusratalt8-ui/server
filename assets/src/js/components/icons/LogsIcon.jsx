export function LogsIcon({ className = "w-5 h-5" }) {
  return (
    <svg className={className} viewBox="0 0 24 24">
      <rect x="2" y="2" width="20" height="20" rx="2" fill="#0c0c0c" stroke="#4a4a4a" strokeWidth="1" />
      <path d="M5 7h4" stroke="#4ec9b0" strokeWidth="1.5" strokeLinecap="round" />
      <path d="M5 11h14" stroke="#888" strokeWidth="1.2" strokeLinecap="round" />
      <path d="M5 14h10" stroke="#888" strokeWidth="1.2" strokeLinecap="round" />
      <path d="M5 17h7" stroke="#888" strokeWidth="1.2" strokeLinecap="round" />
    </svg>
  );
}