export default function AuthLayout({ title, children }) {
  return (
    <div className="auth-bg min-h-screen flex items-center justify-center p-4">
      <div className="auth-card w-full max-w-[400px] p-8">
        <div style={{ textAlign: "center", marginBottom: 28 }}>
          <h1 style={{
            fontSize: 22,
            fontWeight: 800,
            letterSpacing: "-0.02em",
            color: "var(--text)",
            margin: 0,
            lineHeight: 1.2,
          }}>
            Agent64
          </h1>
          <div style={{
            width: 40,
            height: 3,
            borderRadius: 2,
            background: "linear-gradient(90deg, var(--accent), transparent)",
            margin: "10px auto 0",
          }} />
        </div>
        {children}
      </div>
    </div>
  );
}