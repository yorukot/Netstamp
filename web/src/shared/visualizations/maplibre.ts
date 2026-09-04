import maplibreWorkerUrl from "maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url";

export type MapLibreModule = typeof import("maplibre-gl");

let modulePromise: Promise<MapLibreModule> | null = null;

// MapLibre v6 ships its worker as a separate module and resolves it relative to its own URL.
// Vite relocates that module (dev pre-bundling, production chunking), so the default lookup 404s
// and vector tiles never parse. Hand MapLibre the worker bundle Vite emits instead.
export function loadMapLibre(): Promise<MapLibreModule> {
	modulePromise ??= import("maplibre-gl").then(maplibregl => {
		maplibregl.setWorkerUrl(maplibreWorkerUrl);
		return maplibregl;
	});

	return modulePromise;
}
