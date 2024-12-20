import { useEffect, useRef, useState } from "react";

interface WebSocketMessage {
	event: string;
	data: any;
	channel?: string;
}

export const useWebSocket = (sessionId: string | null) => {
	const [isConnected, setIsConnected] = useState(false);
	const [messages, setMessages] = useState<WebSocketMessage[]>([]);
	const wsRef = useRef<WebSocket | null>(null);
	const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);

	useEffect(() => {
		const connectWebSocket = () => {
			const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
			const url = sessionId
				? `${protocol}//${window.location.host}/app/demo?session=${sessionId}`
				: `${protocol}//${window.location.host}/app/demo`;

			const ws = new WebSocket(url);

			ws.onopen = () => {
				console.log("Connected to WebSocket");
				setIsConnected(true);
			};

			ws.onclose = () => {
				console.log("Disconnected from WebSocket");
				setIsConnected(false);
			};

			ws.onmessage = (event) => {
				const message = JSON.parse(event.data);
				if (message.event === "pusher:connection_established") {
					const data = JSON.parse(message.data);
					setCurrentSessionId(data.session_id);
				}
				setMessages((prev) => [...prev, message]);
			};

			wsRef.current = ws;

			return () => {
				ws.close();
			};
		};

		const cleanup = connectWebSocket();
		return cleanup;
	}, [sessionId]);

	const subscribe = (channel: string) => {
		if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
			wsRef.current.send(
				JSON.stringify({
					event: "pusher:subscribe",
					data: { channel },
				}),
			);
		}
	};

	const triggerEvent = (eventType: string) => {
		if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
			wsRef.current.send(
				JSON.stringify({
					event: "trigger_event",
					data: { type: eventType },
				}),
			);
		}
	};

	const setMessageRate = (rate: number) => {
		if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
			wsRef.current.send(
				JSON.stringify({
					event: "set_message_rate",
					data: { rate },
				}),
			);
		}
	};

	const toggleChatInterval = () => {
		if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
			wsRef.current.send(
				JSON.stringify({
					event: "toggle_chat_interval",
				}),
			);
		}
	};

	const toggleAllEventsInterval = () => {
		if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
			wsRef.current.send(
				JSON.stringify({
					event: "toggle_all_events_interval",
				}),
			);
		}
	};

	return {
		isConnected,
		messages,
		subscribe,
		triggerEvent,
		setMessageRate,
		toggleChatInterval,
		toggleAllEventsInterval,
		sessionId: currentSessionId,
	};
};
