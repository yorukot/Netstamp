/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_NETSTAMP_API_BASE_URL?: string;
	readonly VITE_NETSTAMP_API_PROXY_TARGET?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
