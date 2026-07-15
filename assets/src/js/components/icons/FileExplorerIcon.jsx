export function FileExplorerIcon({ className = "w-5 h-5" }) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none">
      {/* tab */}
      <path d="M2 8 Q2 6 4 6 L9 6 L11 8 L22 8 Q23 8 23 9 L23 19 Q23 20 22 20 L2 20 Q1 20 1 19 L1 9 Q1 8 2 8Z"
        fill="#e5c07b" />
      {/* shadow/depth bottom */}
      <path d="M1 14 L1 19 Q1 20 2 20 L22 20 Q23 20 23 19 L23 14Z"
        fill="#c9a84c" opacity="0.5" />
      {/* shine top */}
      <path d="M2 8 L22 8 L22 11 Q12 12 2 11Z"
        fill="white" opacity="0.15" />
    </svg>
  );
}