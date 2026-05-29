import type { Map as MapLibreMap, Marker as MapLibreMarker, StyleSpecification } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { useEffect, useMemo, useRef, useState } from "react";
import styles from "./NetworkMap.module.css";

export interface NetworkMapMarker {
	id: string;
	name: string;
	coordinates?: [number, number];
}

interface NetworkMapProps {
	probes: NetworkMapMarker[];
	selectedId: string;
	onSelect?: (probeId: string) => void;
	mode?: "fleet" | "detail";
	fleetFitPadding?: MapPadding;
	fleetMaxZoom?: number;
	className?: string;
}

const defaultCenter: [number, number] = [74, 29];
const defaultFleetFitPadding = { top: 128, right: 96, bottom: 180, left: 96 };
const defaultFleetMaxZoom = 4.2;
type MapLibreModule = typeof import("maplibre-gl");
type MapPadding = number | { top: number; right: number; bottom: number; left: number };
interface MarkerRecord {
	marker: MapLibreMarker;
	element: HTMLButtonElement;
	probeId: string;
}

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

function setMarkerActive(element: HTMLElement, active: boolean) {
	element.dataset.active = String(active);
}

function createMarkerElement(probe: NetworkMapMarker, mode: "fleet" | "detail", onSelect?: (probeId: string) => void) {
	const markerEl = document.createElement("button");
	markerEl.type = "button";
	markerEl.className = styles.marker;
	markerEl.dataset.mode = mode;
	markerEl.dataset.clickable = String(Boolean(onSelect));
	markerEl.setAttribute("aria-label", `Select probe ${probe.name}`);

	markerEl.addEventListener("click", event => {
		event.stopPropagation();
		onSelect?.(probe.id);
	});

	const labelEl = document.createElement("div");
	labelEl.className = styles.markerLabel;
	labelEl.textContent = probe.name;

	const squareEl = document.createElement("div");
	squareEl.className = styles.markerSquare;

	markerEl.appendChild(labelEl);
	markerEl.appendChild(squareEl);

	return markerEl;
}

function clearMarkers(markers: MarkerRecord[]) {
	for (const { marker } of markers) {
		marker.remove();
	}
}

function hasCoordinates(probe: NetworkMapMarker): probe is NetworkMapMarker & { coordinates: [number, number] } {
	return Array.isArray(probe.coordinates);
}

function fitFleetBounds(map: MapLibreMap, maplibregl: MapLibreModule, probes: Array<NetworkMapMarker & { coordinates: [number, number] }>, padding: MapPadding, maxZoom: number) {
	map.resize();

	const bounds = new maplibregl.LngLatBounds(probes[0].coordinates, probes[0].coordinates);

	for (const probe of probes.slice(1)) {
		bounds.extend(probe.coordinates);
	}

	map.fitBounds(bounds, {
		padding,
		maxZoom,
		duration: 520
	});
}

export function NetworkMap({ probes, selectedId, onSelect, mode = "fleet", fleetFitPadding = defaultFleetFitPadding, fleetMaxZoom = defaultFleetMaxZoom, className }: NetworkMapProps) {
	const mapContainerRef = useRef<HTMLDivElement | null>(null);
	const maplibreglRef = useRef<MapLibreModule | null>(null);
	const mapRef = useRef<MapLibreMap | null>(null);
	const markersRef = useRef<MarkerRecord[]>([]);
	const selectedIdRef = useRef(selectedId);
	const [mapReady, setMapReady] = useState(false);
	const classes = ["ns-cut-frame", styles.map, className].filter(Boolean).join(" ");
	const positionedProbes = useMemo(() => {
		const probesWithCoordinates = probes.filter(hasCoordinates);

		if (mode !== "detail" || !selectedId) {
			return probesWithCoordinates;
		}

		return probesWithCoordinates.filter(probe => probe.id === selectedId);
	}, [mode, probes, selectedId]);

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

		let animationFrame = 0;
		const resizeObserver = new ResizeObserver(() => {
			const map = mapRef.current;
			const maplibregl = maplibreglRef.current;

			map?.resize();

			if (!map || !maplibregl || !mapReady || mode !== "fleet" || positionedProbes.length === 0 || !map.loaded()) {
				return;
			}

			if (animationFrame) {
				window.cancelAnimationFrame(animationFrame);
			}

			animationFrame = window.requestAnimationFrame(() => {
				animationFrame = 0;
				fitFleetBounds(map, maplibregl, positionedProbes, fleetFitPadding, fleetMaxZoom);
			});
		});

		resizeObserver.observe(mapContainerRef.current);

		return () => {
			resizeObserver.disconnect();

			if (animationFrame) {
				window.cancelAnimationFrame(animationFrame);
			}
		};
	}, [fleetFitPadding, fleetMaxZoom, mapReady, mode, positionedProbes]);

	useEffect(() => {
		selectedIdRef.current = selectedId;

		for (const record of markersRef.current) {
			setMarkerActive(record.element, record.probeId === selectedId);
		}
	}, [selectedId]);

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
			markersRef.current = positionedProbes.map(probe => {
				const element = createMarkerElement(probe, mode, onSelect);
				setMarkerActive(element, probe.id === selectedIdRef.current);

				const marker = new activeMaplibregl.Marker({
					element,
					anchor: "bottom"
				})
					.setLngLat(probe.coordinates)
					.addTo(activeMap);

				return { marker, element, probeId: probe.id };
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
	}, [mapReady, mode, onSelect, positionedProbes]);

	useEffect(() => {
		const map = mapRef.current;

		if (!map || !mapReady || mode !== "detail") {
			return undefined;
		}

		const selectedProbe = selectedId ? positionedProbes.find(probe => probe.id === selectedId) : positionedProbes[0];

		if (!selectedProbe) {
			return undefined;
		}

		const activeMap = map;
		const selectedCoordinates = selectedProbe.coordinates;

		function focusSelectedProbe() {
			activeMap.easeTo({
				center: selectedCoordinates,
				zoom: 12.35,
				pitch: 0,
				bearing: 0,
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
	}, [mapReady, mode, positionedProbes, selectedId]);

	useEffect(() => {
		const map = mapRef.current;
		const maplibregl = maplibreglRef.current;

		if (!map || !maplibregl || !mapReady || mode !== "fleet" || positionedProbes.length === 0) {
			return undefined;
		}

		const activeMap = map;
		const activeMaplibregl = maplibregl;

		function focusFleetBounds() {
			fitFleetBounds(activeMap, activeMaplibregl, positionedProbes, fleetFitPadding, fleetMaxZoom);
		}

		if (activeMap.loaded()) {
			focusFleetBounds();
		} else {
			activeMap.once("load", focusFleetBounds);
		}

		return () => {
			activeMap.off("load", focusFleetBounds);
		};
	}, [fleetFitPadding, fleetMaxZoom, mapReady, mode, positionedProbes]);

	return (
		<div className={classes}>
			<div ref={mapContainerRef} className={styles.canvas} />
		</div>
	);
}
