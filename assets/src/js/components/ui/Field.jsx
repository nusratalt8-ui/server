export default function Field({ label, type = "text", value, onChange, autoFocus, autoComplete, placeholder }) {
  const hasValue = value && value.length > 0;
  return (
    <label className="block mb-3" style={{ position: "relative" }}>
      <span className="auth-label">{label}</span>
      <input
        type={type}
        value={value}
        autoFocus={autoFocus}
        autoComplete={autoComplete}
        placeholder={placeholder || label}
        onChange={(e) => onChange(e.target.value)}
        className="auth-field"
      />
    </label>
  );
}