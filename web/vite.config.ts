import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, new URL(".", import.meta.url).pathname, "");
	const apiProxyTarget = env.VITE_NETSTAMP_API_PROXY_TARGET || "http://localhost:8080";

	return {
		plugins: [react()],
		resolve: {
			alias: {
				"@": new URL("./src", import.meta.url).pathname
			}
		},
		server: {
			proxy: {
				"/api": {
					target: apiProxyTarget,
					changeOrigin: true,
					secure: false
				}
			}
		}
	};
});
