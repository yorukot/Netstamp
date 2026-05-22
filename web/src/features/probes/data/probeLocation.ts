export type CoordinateInputMode = "search" | "manual";
export type GeocodeStatus = "idle" | "searching" | "resolved" | "error";

export interface ProbeCoordinates {
	latitude: number;
	longitude: number;
}

interface NominatimSearchResult {
	lat?: string;
	lon?: string;
}

const nominatimSearchEndpoint = "https://nominatim.openstreetmap.org/search";

export function parseCoordinateInput(value: string) {
	const trimmed = value.trim();

	if (!trimmed) {
		return null;
	}

	const coordinate = Number(trimmed);
	return Number.isFinite(coordinate) ? coordinate : null;
}

export function coordinateInputError(label: string, value: string, min: number, max: number) {
	if (!value.trim()) {
		return `${label} is required.`;
	}

	const coordinate = parseCoordinateInput(value);

	if (coordinate === null) {
		return `${label} must be a number.`;
	}

	if (coordinate < min || coordinate > max) {
		return `${label} must be between ${min} and ${max}.`;
	}

	return "";
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
	const language = navigator.language || navigator.languages?.[0];

	if (language) {
		searchParams.set("accept-language", language);
	}

	const response = await fetch(`${nominatimSearchEndpoint}?${searchParams.toString()}`, {
		headers: { Accept: "application/json" },
		signal
	});

	if (!response.ok) {
		throw new Error("Location search failed. Try again later.");
	}

	const results = (await response.json()) as NominatimSearchResult[];
	const result = results[0];
	const latitude = result ? Number(result.lat) : Number.NaN;
	const longitude = result ? Number(result.lon) : Number.NaN;

	if (!result || !Number.isFinite(latitude) || !Number.isFinite(longitude)) {
		throw new Error("No usable location result was found.");
	}

	if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
		throw new Error("The returned coordinates are outside valid ranges.");
	}

	return {
		coordinates: { latitude, longitude },
		locationName: normalizedQuery
	};
}
