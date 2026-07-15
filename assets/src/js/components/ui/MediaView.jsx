import { useState } from "react";

export default function MediaView({ url, alt = "", maxWidth = 320, className = "" }) {
  const [full, setFull] = useState(false);
  if (!url) return null;

  return (
    <>
      <div className={className} style={{ maxWidth }}>
        <img
          src={url}
          alt={alt}
          className="bevel-out w-full cursor-pointer"
          loading="lazy"
          onClick={() => setFull(true)}
        />
      </div>
      {full && (
        <div
          className="file-modal-backdrop"
          onClick={() => setFull(false)}
          style={{
            position: "fixed", inset: 0, zIndex: 9999,
            background: "rgba(0,0,0,0.85)",
            display: "flex", alignItems: "center", justifyContent: "center",
            cursor: "pointer",
          }}
        >
          <img
            src={url}
            alt={alt}
            onClick={(e) => e.stopPropagation()}
            style={{ maxWidth: "90vw", maxHeight: "90vh", objectFit: "contain" }}
          />
        </div>
      )}
    </>
  );
}