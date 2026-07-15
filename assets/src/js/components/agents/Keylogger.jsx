import { useState, useEffect, useRef } from "react";
import { send } from "../../socket";
import { on, off } from "../../socket/events";

export default function Keylogger({ agent }) {
	const agentId = String(agent.id);
	const [text, setText] = useState("");
	const scrollRef = useRef(null);

	useEffect(() => {
		send("keylog_open", { agent_id: agentId });

		const onResult = ({ t, d } = {}) => {
			if (t === "data" && d?.text) {
				setText((prev) => prev + d.text);
			}
		};
		on("keylog_result", onResult);

		return () => {
			off("keylog_result", onResult);
			send("keylog_close", { agent_id: agentId });
		};
	}, [agentId]);

	useEffect(() => {
		if (scrollRef.current) {
			scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
		}
	}, [text]);

	return (
		<div className="h-full" style={{ padding: 8 }}>
			<div
				ref={scrollRef}
				style={{
					height: "100%",
					overflow: "auto",
					background: "var(--edge-dark)",
					borderRadius: 4,
					padding: 8,
					fontFamily: "monospace",
					fontSize: 11,
					whiteSpace: "pre-wrap",
					wordBreak: "break-all",
					color: "var(--text)",
				}}
			>
				{text || <span style={{ color: "var(--muted)" }}>listening...</span>}
			</div>
		</div>
	);
}