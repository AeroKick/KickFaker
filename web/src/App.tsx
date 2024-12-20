import { useEffect, useState } from "react";
import { useWebSocket } from "./hooks/useWebSocket";
import { useSearchParams } from "react-router-dom";

const eventTypes = [
	"chat",
	"chat_celebration",
	"subscription",
	"gifted_subscriptions",
	"raid",
	"live",
	"stop_broadcast",
] as const;

function App() {
	const [searchParams, setSearchParams] = useSearchParams();
	const sessionId = searchParams.get("session");
	const {
		isConnected,
		messages,
		subscribe,
		triggerEvent,
		setMessageRate,
		toggleChatInterval,
		toggleAllEventsInterval,
		sessionId: currentSessionId,
	} = useWebSocket(sessionId);

	const [messageRate, setLocalMessageRate] = useState(1);
	const [isChatActive, setIsChatActive] = useState(false);
	const [isAllEventsActive, setIsAllEventsActive] = useState(false);

	useEffect(() => {
		if (isConnected) {
			subscribe("chatroom-1");
			subscribe("channel-1");
		}
	}, [isConnected]);

	useEffect(() => {
		if (currentSessionId && !sessionId) {
			setSearchParams({ session: currentSessionId });
		}
	}, [currentSessionId]);

	const handleMessageRateChange = (rate: number) => {
		setLocalMessageRate(rate);
		setMessageRate(rate);
	};

	const handleToggleChat = () => {
		setIsChatActive(!isChatActive);
		toggleChatInterval();
	};

	const handleToggleAllEvents = () => {
		setIsAllEventsActive(!isAllEventsActive);
		toggleAllEventsInterval();
	};

	const getWebSocketUrl = () => {
		const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
		return `${protocol}//${window.location.host}/ws${currentSessionId ? `?session=${currentSessionId}` : ""}`;
	};

	return (
		<div className="min-h-screen bg-gray-900 text-white p-6">
			<div className="container mx-auto">
				<header className="mb-8">
					<h1 className="text-3xl font-bold mb-2">KickFaker Demo</h1>
					<div className="flex items-center space-x-4">
						<div
							className={`h-3 w-3 rounded-full ${isConnected ? "bg-green-500" : "bg-red-500"}`}
						/>
						<p className="text-sm text-gray-400">
							{isConnected ? "Connected" : "Disconnected"}
						</p>
					</div>
					{currentSessionId && (
						<div className="mt-2 p-4 bg-gray-800 rounded-lg">
							<p className="text-sm text-gray-400">WebSocket URL:</p>
							<code className="text-green-400 break-all">
								{getWebSocketUrl()}
							</code>
						</div>
					)}
				</header>

				<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<div className="space-y-6">
						<div className="bg-gray-800 p-6 rounded-lg">
							<h2 className="text-xl font-semibold mb-4">Controls</h2>

							<div className="space-y-4">
								<div>
									<label className="block text-sm text-gray-400 mb-2">
										Message Rate (per second)
									</label>
									<input
										type="range"
										min="1"
										max="10"
										value={messageRate}
										onChange={(e) =>
											handleMessageRateChange(Number(e.target.value))
										}
										className="w-full"
									/>
									<span className="text-sm text-gray-400">{messageRate}/s</span>
								</div>

								<div className="flex space-x-4">
									<button
										onClick={handleToggleChat}
										className={`px-4 py-2 rounded-lg ${
											isChatActive
												? "bg-red-600 hover:bg-red-700"
												: "bg-green-600 hover:bg-green-700"
										} transition-colors`}
									>
										{isChatActive ? "Stop Chat" : "Start Chat"}
									</button>

									<button
										onClick={handleToggleAllEvents}
										className={`px-4 py-2 rounded-lg ${
											isAllEventsActive
												? "bg-red-600 hover:bg-red-700"
												: "bg-green-600 hover:bg-green-700"
										} transition-colors`}
									>
										{isAllEventsActive ? "Stop All Events" : "Start All Events"}
									</button>
								</div>
							</div>
						</div>

						<div className="bg-gray-800 p-6 rounded-lg">
							<h2 className="text-xl font-semibold mb-4">Trigger Events</h2>
							<div className="grid grid-cols-2 gap-3">
								{eventTypes.map((type) => (
									<button
										key={type}
										onClick={() => triggerEvent(type)}
										className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
									>
										{type
											.split("_")
											.map(
												(word) => word.charAt(0).toUpperCase() + word.slice(1),
											)
											.join(" ")}
									</button>
								))}
							</div>
						</div>
					</div>

					<div className="bg-gray-800 p-6 rounded-lg">
						<h2 className="text-xl font-semibold mb-4">Event Log</h2>
						<div className="h-[600px] overflow-y-auto space-y-2">
							{messages.map((message, index) => (
								<div key={index} className="p-3 bg-gray-700 rounded">
									<div className="text-sm text-gray-400">{message.event}</div>
									{message.channel && (
										<div className="text-xs text-gray-500">
											Channel: {message.channel}
										</div>
									)}
									<pre className="mt-1 text-sm overflow-x-auto">
										{JSON.stringify(message.data, null, 2)}
									</pre>
								</div>
							))}
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}

export default App;
