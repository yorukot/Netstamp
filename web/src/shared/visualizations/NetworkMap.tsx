import { useTheme } from "@/shared/theme/useTheme";
import { Spinner } from "@netstamp/ui";
import type { Map as MapLibreMap, Marker as MapLibreMarker } from "maplibre-gl";
import "maplibre-gl/dist/maplibre-gl.css";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { loadMapLibre, type MapLibreModule } from "./maplibre";
import styles from "./NetworkMap.module.css";

export interface NetworkMapMarker {
	id: string;
	name: string;
	coordinates?: [number, number];
	status?: string;
}

interface NetworkMapProps {
	probes: NetworkMapMarker[];
	selectedId: string;
	onSelect?: (probeId: string) => void;
	mode?: "fleet" | "detail";
	theme?: MapTheme;
	fleetFitPadding?: MapPadding;
	fleetMaxZoom?: number;
	isLoading?: boolean;
	loadingLabel?: string;
	className?: string;
}

const defaultCenter: [number, number] = [74, 29];
const defaultFleetFitPadding = { top: 128, right: 96, bottom: 180, left: 96 };
const defaultFleetMaxZoom = 4.2;
type MapTheme = "dark" | "light";
type MapPadding = number | { top: number; right: number; bottom: number; left: number };
interface MarkerRecord {
	marker: MapLibreMarker;
	element: HTMLButtonElement;
	probeId: string;
}

// OpenFreeMap serves keyless vector basemaps derived from the CARTO Positron / Dark Matter designs.
// The style JSON carries its own OpenStreetMap / OpenFreeMap attribution.
function getBasemapStyleUrl(theme: MapTheme): string {
	return theme === "light" ? "https://tiles.openfreemap.org/styles/positron" : "https://tiles.openfreemap.org/styles/dark";
}

function setMarkerActive(element: HTMLElement, active: boolean) {
	element.dataset.active = String(active);
}

function updateMarkerElement(element: HTMLButtonElement, probe: NetworkMapMarker, mode: "fleet" | "detail", clickable: boolean, ariaLabel: string) {
	element.dataset.mode = mode;
	element.dataset.clickable = String(clickable);
	element.setAttribute("aria-label", ariaLabel);

	if (probe.status) {
		element.dataset.status = probe.status.toLowerCase();
	} else {
		delete element.dataset.status;
	}

	const label = element.querySelector<HTMLElement>("[data-marker-label]");

	if (label) {
		label.textContent = probe.name;
	}
}

function createMarkerElement(probe: NetworkMapMarker, mode: "fleet" | "detail", clickable: boolean, ariaLabel: string, onSelect: (probeId: string) => void) {
	const markerEl = document.createElement("button");
	markerEl.type = "button";
	markerEl.className = styles.marker;

	markerEl.addEventListener("click", event => {
		event.stopPropagation();

		if (markerEl.dataset.clickable === "true") {
			onSelect(probe.id);
		}
	});

	const labelEl = document.createElement("div");
	labelEl.className = styles.markerLabel;
	labelEl.dataset.markerLabel = "";
	labelEl.textContent = probe.name;

	const squareEl = document.createElement("div");
	squareEl.className = styles.markerSquare;

	markerEl.appendChild(labelEl);
	markerEl.appendChild(squareEl);
	updateMarkerElement(markerEl, probe, mode, clickable, ariaLabel);

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

export function NetworkMap({
	probes,
	selectedId,
	onSelect,
	mode = "fleet",
	theme: themeOverride,
	fleetFitPadding = defaultFleetFitPadding,
	fleetMaxZoom = defaultFleetMaxZoom,
	isLoading = false,
	loadingLabel,
	className
}: NetworkMapProps) {
	const { t } = useTranslation("common");
	const { theme: appTheme } = useTheme();
	const mapTheme = themeOverride ?? appTheme;
	const mapContainerRef = useRef<HTMLDivElement | null>(null);
	const maplibreglRef = useRef<MapLibreModule | null>(null);
	const mapRef = useRef<MapLibreMap | null>(null);
	const markersRef = useRef<MarkerRecord[]>([]);
	const selectedIdRef = useRef(selectedId);
	const onSelectRef = useRef(onSelect);
	const themeRef = useRef(mapTheme);
	const appliedThemeRef = useRef<MapTheme | null>(null);
	const [mapReady, setMapReady] = useState(false);
	const classes = [styles.map, `ns-theme-${mapTheme}`, className].filter(Boolean).join(" ");
	const markerClickable = Boolean(onSelect);
	const showLoadingOverlay = isLoading || !mapReady;
	const positionedProbes = useMemo(() => {
		const probesWithCoordinates = probes.filter(hasCoordinates);

		if (mode !== "detail" || !selectedId) {
			return probesWithCoordinates;
		}

		return probesWithCoordinates.filter(probe => probe.id === selectedId);
	}, [mode, probes, selectedId]);

	useEffect(() => {
		themeRef.current = mapTheme;
	}, [mapTheme]);

	useEffect(() => {
		onSelectRef.current = onSelect;
	}, [onSelect]);

	useEffect(() => {
		let cancelled = false;

		async function initializeMap() {
			const maplibregl = await loadMapLibre();

			if (cancelled || !mapContainerRef.current || mapRef.current) {
				return;
			}

			const map = new maplibregl.Map({
				container: mapContainerRef.current,
				style: getBasemapStyleUrl(themeRef.current),
				center: defaultCenter,
				zoom: 2.15,
				attributionControl: { compact: true }
			});

			maplibreglRef.current = maplibregl;
			mapRef.current = map;
			appliedThemeRef.current = themeRef.current;
			map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "bottom-right");
			map.once("load", () => {
				if (cancelled) {
					return;
				}

				map.resize();
				setMapReady(true);
			});
		}

		initializeMap();

		return () => {
			cancelled = true;
			clearMarkers(markersRef.current);
			markersRef.current = [];
			mapRef.current?.remove();
			maplibreglRef.current = null;
			mapRef.current = null;
			appliedThemeRef.current = null;
		};
	}, []);

	useEffect(() => {
		const map = mapRef.current;

		if (!map || !mapReady || appliedThemeRef.current === mapTheme) {
			return;
		}

		appliedThemeRef.current = mapTheme;
		map.setStyle(getBasemapStyleUrl(mapTheme));
	}, [mapReady, mapTheme]);

	useEffect(() => {
		if (!mapContainerRef.current) {
			return undefined;
		}

		let animationFrame = 0;
		const resizeObserver = new ResizeObserver(() => {
			const map = mapRef.current;
			const maplibregl = maplibreglRef.current;

			map?.resize();

			if (!map || !maplibregl || !mapReady || mode !== "fleet" || positionedProbes.length === 0) {
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
			activeMap.resize();

			const existingMarkers = new Map(markersRef.current.map(record => [record.probeId, record]));
			const nextMarkers = positionedProbes.map(probe => {
				const existing = existingMarkers.get(probe.id);

				if (existing) {
					updateMarkerElement(existing.element, probe, mode, markerClickable, t("map.selectProbe", { name: probe.name }));
					setMarkerActive(existing.element, probe.id === selectedIdRef.current);
					existing.marker.setLngLat(probe.coordinates);
					existingMarkers.delete(probe.id);
					return existing;
				}

				const element = createMarkerElement(probe, mode, markerClickable, t("map.selectProbe", { name: probe.name }), probeId => onSelectRef.current?.(probeId));
				setMarkerActive(element, probe.id === selectedIdRef.current);

				const marker = new activeMaplibregl.Marker({
					element,
					anchor: "bottom"
				})
					.setLngLat(probe.coordinates)
					.addTo(activeMap);

				return { marker, element, probeId: probe.id };
			});

			clearMarkers([...existingMarkers.values()]);
			markersRef.current = nextMarkers;
		}

		renderMarkers();
	}, [mapReady, markerClickable, mode, positionedProbes, t]);

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

		focusSelectedProbe();

		return () => {
			activeMap.stop();
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

		focusFleetBounds();

		return () => {
			activeMap.stop();
		};
	}, [fleetFitPadding, fleetMaxZoom, mapReady, mode, positionedProbes]);

	return (
		<div className={classes} data-map-theme={mapTheme}>
			<div ref={mapContainerRef} className={styles.canvas} />
			{showLoadingOverlay ? (
				<div className={styles.loadingOverlay}>
					<div className={styles.loadingFrame}>
						<Spinner label={loadingLabel ?? t("map.loading")} size="lg" />
					</div>
				</div>
			) : null}
		</div>
	);
}
