import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [react()],
	server: {
		proxy: {
			"/ws": {
				target: "ws://localhost:4400",
				ws: true,
			},
		},
	},
});
