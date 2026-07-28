/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_NETSTAMP_API_BASE_URL?: string;
	readonly VITE_NETSTAMP_API_PROXY_TARGET?: string;
	readonly VITE_NETSTAMP_DEMO_EMAIL?: string;
	readonly VITE_NETSTAMP_DEMO_PASSWORD?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
