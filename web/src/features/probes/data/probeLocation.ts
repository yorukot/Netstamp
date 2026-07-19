export type CoordinateInputMode = "search" | "manual";
export type GeocodeStatus = "idle" | "searching" | "resolved" | "error";
export type CoordinateInputErrorCode = "required" | "number" | "range";
export type LocationSearchErrorCode = "searchFailed" | "noResult" | "outsideRange";

export interface ProbeCoordinates {
	latitude: number;
	longitude: number;
}

interface NominatimSearchResult {
	lat?: string;
	lon?: string;
}

const nominatimSearchEndpoint = "https://nominatim.openstreetmap.org/search";

export class LocationSearchError extends Error {
	constructor(public readonly code: LocationSearchErrorCode) {
		super(code);
		this.name = "LocationSearchError";
	}
}

export function parseCoordinateInput(value: string) {
	const trimmed = value.trim();

	if (!trimmed) {
		return null;
	}

	const coordinate = Number(trimmed);
	return Number.isFinite(coordinate) ? coordinate : null;
}

export function coordinateInputError(value: string, min: number, max: number): CoordinateInputErrorCode | null {
	if (!value.trim()) {
		return "required";
	}

	const coordinate = parseCoordinateInput(value);

	if (coordinate === null) {
		return "number";
	}

	if (coordinate < min || coordinate > max) {
		return "range";
	}

	return null;
}

export function formatCoordinate(value: number) {
	return value.toFixed(6);
}

export function coordinateSummary(latitude: number | null | undefined, longitude: number | null | undefined) {
	return typeof latitude === "number" && typeof longitude === "number" ? `${formatCoordinate(latitude)}, ${formatCoordinate(longitude)}` : "";
}

export async function searchNominatimLocation(query: string, signal?: AbortSignal) {
	const normalizedQuery = query.trim();
	const searchParams = new URLSearchParams({
		q: normalizedQuery,
		format: "jsonv2",
		limit: "1"
	});
	const language = document.documentElement.lang || navigator.language || navigator.languages?.[0];

	if (language) {
		searchParams.set("accept-language", language);
	}

	const response = await fetch(`${nominatimSearchEndpoint}?${searchParams.toString()}`, {
		headers: { Accept: "application/json" },
		signal
	});

	if (!response.ok) {
		throw new LocationSearchError("searchFailed");
	}

	const results = (await response.json()) as NominatimSearchResult[];
	const result = results[0];
	const latitude = result ? Number(result.lat) : Number.NaN;
	const longitude = result ? Number(result.lon) : Number.NaN;

	if (!result || !Number.isFinite(latitude) || !Number.isFinite(longitude)) {
		throw new LocationSearchError("noResult");
	}

	if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
		throw new LocationSearchError("outsideRange");
	}

	return {
		coordinates: { latitude, longitude },
		locationName: normalizedQuery
	};
}
