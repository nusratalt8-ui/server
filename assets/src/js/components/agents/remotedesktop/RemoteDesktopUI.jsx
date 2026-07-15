import { useState, useRef, useEffect, useCallback } from "react";
import StreamView from "./StreamView";
import ControlPanel from "./ControlPanel";
import LedRow from "./LedRow";
import useRemoteDesktop from "./useRemoteDesktop";
import { FullscreenIcon, ExitFullscreenIcon } from "../../../components/icons/FullscreenIcon";
import { LED_ON, LED_OFF, LED_GLOW, MM, MT, MHI, M, KEY_HI, KEY_BG, MONO } from "./constants";
import VolumeSlider from "../../ui/VolumeSlider";
import Dropdown, { DropdownItem } from "../../ui/Dropdown";

export default function RemoteDesktopUI({ agent }) {
  const containerRef = useRef(null);
  const [panelOpen, setPanelOpen] = useState(true);
  const [isFullscreen, setIsFullscreen] = useState(false);

  const {
    canvasRef, wrapRef, fps, stale, res, status, errMsg, vol,
    action, onMouseMove, onMouseDown, onMouseUp, onWheel,
    onKeyDown, onKeyUp,
    micOn, micMuted, micDevices, micDevice, setMicDevice,
    requestMicDevices, startMic, stopMic, toggleMute,
  } = useRemoteDesktop(String(agent.id));

  const handleRequestDevices = () => {
    requestMicDevices();
  };

  const toggleFullscreen = useCallback(() => {
    const el = containerRef.current;
    if (!el) return;
    if (!document.fullscreenElement) {
      el.requestFullscreen?.() || el.webkitRequestFullscreen?.();
    } else {
      document.exitFullscreen?.() || document.webkitExitFullscreen?.();
    }
  }, []);

  useEffect(() => {
    const onChange = () => setIsFullscreen(!!document.fullscreenElement);
    document.addEventListener("fullscreenchange", onChange);
    document.addEventListener("webkitfullscreenchange", onChange);
    return () => {
      document.removeEventListener("fullscreenchange", onChange);
      document.removeEventListener("webkitfullscreenchange", onChange);
    };
  }, []);

  return (
    <div
      ref={containerRef}
      className="h-full flex flex-col outline-none"
      style={{ background: "#0a0a0a", position: "relative" }}
      tabIndex={0}
      onKeyDown={onKeyDown}
      onKeyUp={onKeyUp}
    >
      {/* Top bar */}
      {!isFullscreen && (
        <div style={{
          display: "flex", alignItems: "center", gap: 6,
          padding: "0 8px", height: 32, flexShrink: 0,
          background: M, borderBottom: `1px solid ${MHI}`,
        }}>
          <span style={{
            fontFamily: MONO, fontSize: 11,
            color: stale ? "#e87070" : LED_ON,
            background: stale ? "rgba(220,50,50,0.12)" : "rgba(0,0,0,0.3)",
            border: `1px solid ${stale ? "rgba(220,50,50,0.3)" : "rgba(255,255,255,0.06)"}`,
            padding: "1px 8px", borderRadius: 3,
          }}>
            {stale ? "signal lost" : `${fps} fps`}
          </span>
          {res.w > 0 && (
            <span style={{
              fontFamily: MONO, fontSize: 11, color: MM,
              background: "rgba(0,0,0,0.3)", border: "1px solid rgba(255,255,255,0.06)",
              padding: "1px 8px", borderRadius: 3,
            }}>
              {res.w}×{res.h}
            </span>
          )}
          {errMsg && (
            <span style={{ fontFamily: MONO, fontSize: 11, color: "#e87070", marginLeft: 4 }}>
              ⚠ {errMsg}
            </span>
          )}
          <div style={{ flex: 1 }} />
          <button
            onClick={toggleFullscreen}
            title="Fullscreen"
            style={{
              background: "transparent", border: "none", cursor: "pointer",
              color: MM, padding: "4px", borderRadius: 3,
              display: "flex", alignItems: "center", transition: "color 0.15s",
            }}
            onMouseEnter={e => e.currentTarget.style.color = MT}
            onMouseLeave={e => e.currentTarget.style.color = MM}
          >
            <FullscreenIcon className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Middle: left mic sidebar + stream + right control panel */}
      <div style={{ flex: 1, minHeight: 0, display: "flex" }}>

        {/* Left mic sidebar */}
        {!isFullscreen && (
          <MicSidebar
            onRequestDevices={handleRequestDevices}
            devices={micDevices} device={micDevice} setDevice={setMicDevice}
            micOn={micOn} startMic={startMic} stopMic={stopMic}
            micMuted={micMuted} toggleMute={toggleMute}
          />
        )}

        <StreamView
          canvasRef={canvasRef} wrapRef={wrapRef} stale={stale}
          onMouseMove={onMouseMove} onMouseDown={onMouseDown}
          onMouseUp={onMouseUp} onWheel={onWheel}
        />

        {!isFullscreen && panelOpen && (
          <ControlPanel agent={agent} status={status} action={action} onHide={() => setPanelOpen(false)} />
        )}
      </div>

      {/* Bottom bar */}
      {!isFullscreen && (
        <div style={{
          display: "flex", alignItems: "center", gap: 10,
          padding: "0 12px", height: 40, flexShrink: 0,
          background: M, borderTop: `1px solid ${MHI}`,
        }}>
          <VolumeSlider value={vol} onChange={(v) => action("volume", v)} />
          <div style={{ flex: 1 }} />
          {!panelOpen && (
            <button
              onClick={() => setPanelOpen(true)}
              style={{
                fontFamily: MONO, fontSize: 11, padding: "3px 10px",
                background: "transparent", border: `1px solid ${MHI}`,
                borderRadius: 3, cursor: "pointer", color: MM,
                transition: "color 0.15s, border-color 0.15s",
              }}
              onMouseEnter={e => { e.currentTarget.style.color = MT; e.currentTarget.style.borderColor = MT; }}
              onMouseLeave={e => { e.currentTarget.style.color = MM; e.currentTarget.style.borderColor = MHI; }}
            >
              ◀ Controls
            </button>
          )}
        </div>
      )}

      {/* Fullscreen exit */}
      {isFullscreen && (
        <button
          onClick={toggleFullscreen}
          title="Exit fullscreen"
          style={{
            position: "absolute", top: 8, right: 8, zIndex: 30,
            background: "rgba(0,0,0,0.6)", border: "1px solid rgba(255,255,255,0.1)",
            cursor: "pointer", color: "#fff", padding: "6px",
            borderRadius: 4, display: "flex", alignItems: "center",
            opacity: 0.5, transition: "opacity 0.2s",
          }}
          onMouseEnter={e => e.currentTarget.style.opacity = "1"}
          onMouseLeave={e => e.currentTarget.style.opacity = "0.5"}
        >
          <ExitFullscreenIcon className="w-4 h-4" />
        </button>
      )}
    </div>
  );
}

function MicSidebar({ onRequestDevices, devices, device, setDevice, micOn, startMic, stopMic, micMuted, toggleMute }) {
  const selectedDevice = devices.find((d) => d.id === device);
  const [requested, setRequested] = useState(false);
  const pendingStart = useRef(false);
  const loading = requested && devices.length === 0;

  useEffect(() => {
    if (pendingStart.current && devices.length > 0 && !micOn) {
      pendingStart.current = false;
      startMic(devices[0].id);
    }
  }, [devices]);

  const micToggle = () => {
    if (micOn) { stopMic(); return; }
    if (devices.length > 0) { startMic(device); return; }
    setRequested(true);
    pendingStart.current = true;
    onRequestDevices();
  };

  return (
    <div style={{
      width: 160, flexShrink: 0,
      background: M,
      backgroundImage: "repeating-linear-gradient(180deg, transparent 0px, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 3px)",
      borderRight: `1px solid ${MHI}`,
      boxShadow: `inset -1px 0 0 ${KEY_HI}`,
      display: "flex", flexDirection: "column",
      padding: "10px 10px",
      gap: 8,
    }} className="vista-scroll">

      <div style={{ paddingBottom: 6, borderBottom: `1px solid ${MHI}`, display: "flex", alignItems: "center", gap: 6 }}>
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke={MM} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="9" y="2" width="6" height="11" rx="3" />
          <path d="M5 10a7 7 0 0014 0M12 19v3M8 22h8" />
        </svg>
        <span style={{ fontFamily: MONO, fontSize: 10, color: MM, letterSpacing: "0.06em", textTransform: "uppercase", fontWeight: 600 }}>
          Mic
        </span>
      </div>

      {/* Single capture toggle */}
      <LedRow
        label={loading ? "Loading…" : micOn ? "Capturing" : "Capture"}
        active={micOn}
        onClick={micToggle}
      />

      {/* Device picker — only when devices loaded */}
      {devices.length > 0 && (
        <Dropdown
          align="left"
          trigger={
            <button style={{
              display: "flex", alignItems: "center", justifyContent: "space-between",
              width: "100%", padding: "5px 8px",
              background: KEY_BG, border: `1px solid ${MHI}`,
              borderRadius: 3, cursor: "pointer",
              fontFamily: MONO, fontSize: 11, color: MT,
              textAlign: "left", gap: 4,
              opacity: micOn ? 0.6 : 1,
            }}>
              <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                {selectedDevice?.name ?? "Select mic"}
              </span>
              <span style={{ color: MM, fontSize: 9, flexShrink: 0 }}>▾</span>
            </button>
          }
        >
          {devices.map((d) => (
            <DropdownItem
              key={d.id}
              checked={d.id === device}
              onClick={() => {
                setDevice(d.id);
                if (micOn) {
                  stopMic();
                  setTimeout(() => startMic(d.id), 100);
                }
              }}
            >
              {d.name}
            </DropdownItem>
          ))}
        </Dropdown>
      )}
    </div>
  );
}