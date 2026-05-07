import { type Probe } from "@/features/probes/data/probes";
import type { Map as MapLibreMap, Marker as MapLibreMarker, StyleSpecification } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { useEffect, useRef, useState } from "react";
import styles from "./NetworkMap.module.css";

interface NetworkMapProps {
	probes: Probe[];
	selectedId: string;
	onSelect?: (probeId: string) => void;
	mode?: "fleet" | "detail";
	className?: string;
}

const defaultCenter: [number, number] = [74, 29];
type MapLibreModule = typeof import("maplibre-gl");

function createCartoDarkStyle(): StyleSpecification {
	return {
		version: 8,
		sources: {
			"carto-dark": {
				type: "raster",
				tiles: [
					"https://a.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png",
					"https://b.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png",
					"https://c.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png",
					"https://d.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png"
				],
				tileSize: 256,
				attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
			}
		},
		layers: [
			{
				id: "carto-dark-base",
				type: "raster",
				source: "carto-dark",
				paint: {
					"raster-opacity": 1,
					"raster-brightness-min": 0.08,
					"raster-brightness-max": 1,
					"raster-contrast": 0.14,
					"raster-saturation": 0
				}
			}
		]
	};
}

function createMarkerElement(probe: Probe, active: boolean, mode: "fleet" | "detail", onSelect?: (probeId: string) => void) {
	const markerEl = document.createElement("button");
	markerEl.type = "button";
	markerEl.setAttribute("aria-label", `Select probe ${probe.name}`);

	Object.assign(markerEl.style, {
		display: "flex",
		flexDirection: "column",
		alignItems: "center",
		border: "0",
		padding: "0",
		color: "#fff",
		background: "transparent",
		cursor: onSelect ? "pointer" : "default",
		pointerEvents: "auto",
		transform: "translateY(-8px)"
	});

	markerEl.addEventListener("click", event => {
		event.stopPropagation();
		onSelect?.(probe.id);
	});

	const labelEl = document.createElement("div");
	labelEl.textContent = probe.name;

	Object.assign(labelEl.style, {
		display: active ? "block" : "none",
		marginBottom: mode === "detail" ? "8px" : "6px",
		padding: mode === "detail" ? "5px 8px" : "4px 7px",
		color: "#ffffff",
		background: "rgba(0, 0, 0, 0.94)",
		border: "2px solid rgba(255, 255, 255, 0.96)",
		borderRadius: "4px",
		fontFamily: "JetBrains Mono, monospace",
		fontSize: mode === "detail" ? "12px" : "10px",
		fontWeight: "900",
		lineHeight: "1",
		letterSpacing: "0.06em",
		textShadow: "0 0 4px #000, 0 0 10px #000",
		textTransform: "uppercase",
		boxShadow: "0 0 18px rgba(255, 106, 0, 0.55)",
		whiteSpace: "nowrap"
	});

	const squareSize = active ? (mode === "detail" ? 18 : 14) : mode === "detail" ? 14 : 11;
	const squareEl = document.createElement("div");

	Object.assign(squareEl.style, {
		width: `${squareSize}px`,
		height: `${squareSize}px`,
		background: "#ff6a00",
		border: active ? "3px solid #ffffff" : "1px solid rgba(255, 255, 255, 0.62)",
		outline: active ? "2px solid #000000" : "1px solid rgba(0, 0, 0, 0.82)",
		boxShadow: active ? "0 0 0 2px rgba(255, 106, 0, 1), 0 0 24px 8px rgba(255, 106, 0, 0.72)" : "0 0 14px 4px rgba(255, 106, 0, 0.5)"
	});

	markerEl.appendChild(labelEl);
	markerEl.appendChild(squareEl);

	return markerEl;
}

function clearMarkers(markers: MapLibreMarker[]) {
	for (const marker of markers) {
		marker.remove();
	}
}

export function NetworkMap({ probes, selectedId, onSelect, mode = "fleet", className }: NetworkMapProps) {
	const mapContainerRef = useRef<HTMLDivElement | null>(null);
	const maplibreglRef = useRef<MapLibreModule | null>(null);
	const mapRef = useRef<MapLibreMap | null>(null);
	const markersRef = useRef<MapLibreMarker[]>([]);
	const [mapReady, setMapReady] = useState(false);
	const classes = ["ns-cut-frame", styles.map, className].filter(Boolean).join(" ");

	useEffect(() => {
		let cancelled = false;

		async function initializeMap() {
			const maplibregl = await import("maplibre-gl");

			if (cancelled || !mapContainerRef.current || mapRef.current) {
				return;
			}

			const map = new maplibregl.Map({
				container: mapContainerRef.current,
				style: createCartoDarkStyle(),
				center: defaultCenter,
				zoom: 2.15,
				attributionControl: { compact: true }
			});

			maplibreglRef.current = maplibregl;
			mapRef.current = map;
			map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "bottom-right");
			setMapReady(true);
		}

		initializeMap();

		return () => {
			cancelled = true;
			clearMarkers(markersRef.current);
			markersRef.current = [];
			mapRef.current?.remove();
			maplibreglRef.current = null;
			mapRef.current = null;
		};
	}, []);

	useEffect(() => {
		if (!mapContainerRef.current) {
			return undefined;
		}

		const resizeObserver = new ResizeObserver(() => {
			mapRef.current?.resize();
		});

		resizeObserver.observe(mapContainerRef.current);

		return () => resizeObserver.disconnect();
	}, []);

	useEffect(() => {
		const map = mapRef.current;
		const maplibregl = maplibreglRef.current;

		if (!map || !maplibregl || !mapReady) {
			return undefined;
		}

		const activeMap = map;
		const activeMaplibregl = maplibregl;

		function renderMarkers() {
			clearMarkers(markersRef.current);
			markersRef.current = probes.map(probe => {
				const marker = new activeMaplibregl.Marker({
					element: createMarkerElement(probe, probe.id === selectedId, mode, onSelect),
					anchor: "bottom"
				})
					.setLngLat(probe.coordinates)
					.addTo(activeMap);

				return marker;
			});
		}

		if (activeMap.loaded()) {
			renderMarkers();
		} else {
			activeMap.once("load", renderMarkers);
		}

		return () => {
			activeMap.off("load", renderMarkers);
			clearMarkers(markersRef.current);
			markersRef.current = [];
		};
	}, [mapReady, mode, onSelect, probes, selectedId]);

	useEffect(() => {
		const map = mapRef.current;

		if (!map || !mapReady || mode !== "detail") {
			return undefined;
		}

		const selectedProbe = probes.find(probe => probe.id === selectedId) || probes[0];

		if (!selectedProbe) {
			return undefined;
		}

		const activeMap = map;

		function focusSelectedProbe() {
			activeMap.easeTo({
				center: selectedProbe.coordinates,
				zoom: 12.35,
				pitch: 35,
				bearing: -20,
				duration: 420
			});
		}

		if (activeMap.loaded()) {
			focusSelectedProbe();
		} else {
			activeMap.once("load", focusSelectedProbe);
		}

		return () => {
			activeMap.off("load", focusSelectedProbe);
		};
	}, [mapReady, mode, probes, selectedId]);

	useEffect(() => {
		const map = mapRef.current;
		const maplibregl = maplibreglRef.current;

		if (!map || !maplibregl || !mapReady || mode !== "fleet" || probes.length === 0) {
			return undefined;
		}

		const activeMap = map;
		const activeMaplibregl = maplibregl;

		function fitFleetBounds() {
			const bounds = new activeMaplibregl.LngLatBounds(probes[0].coordinates, probes[0].coordinates);

			for (const probe of probes.slice(1)) {
				bounds.extend(probe.coordinates);
			}

			activeMap.fitBounds(bounds, {
				padding: { top: 128, right: 96, bottom: 180, left: 96 },
				maxZoom: 4.2,
				duration: 520
			});
		}

		if (activeMap.loaded()) {
			fitFleetBounds();
		} else {
			activeMap.once("load", fitFleetBounds);
		}

		return () => {
			activeMap.off("load", fitFleetBounds);
		};
	}, [mapReady, mode, probes]);

	return (
		<div className={classes}>
			<div ref={mapContainerRef} className={styles.canvas} />
		</div>
	);
}
