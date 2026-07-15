export function SystemIcon({ className }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="2" y="3" width="20" height="14" rx="2" stroke="currentColor" strokeWidth="1.5" fill="none"/>
      <path d="M8 21h8M12 17v4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
      <path d="M6 7h4M6 10h8M6 13h5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" opacity="0.7"/>
      <circle cx="17" cy="10" r="2.5" stroke="currentColor" strokeWidth="1.2" fill="none"/>
      <path d="M17 8.5V10l1 1" stroke="currentColor" strokeWidth="1" strokeLinecap="round"/>
    </svg>
  );
}