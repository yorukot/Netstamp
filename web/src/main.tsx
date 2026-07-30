import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { HelmetProvider } from "react-helmet-async";
import App from "./App";
import { initializeI18n } from "./i18n";
import "./index.css";
import { loadRuntimeConfig } from "./shared/config/runtime";

const root = document.getElementById("root");

if (!root) {
	throw new Error("Root element not found");
}

await initializeI18n();
await loadRuntimeConfig().catch(() => {
	// Runtime feature defaults are fail-closed so an unavailable API cannot enable gated actions.
});

createRoot(root).render(
	<StrictMode>
		<HelmetProvider>
			<App />
		</HelmetProvider>
	</StrictMode>
);
