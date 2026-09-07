const DEFAULT_UMAMI_SCRIPT_URL = "https://cloud.umami.is/script.js";
const DEFAULT_STORAGE_KEY = "netstamp.tracking-consent.v1";

export const DEFAULT_TRACKING_CONSENT_COUNTRIES = [
	"AT",
	"BE",
	"BG",
	"HR",
	"CY",
	"CZ",
	"DK",
	"EE",
	"FI",
	"FR",
	"DE",
	"GR",
	"HU",
	"IS",
	"IE",
	"IT",
	"LV",
	"LI",
	"LT",
	"LU",
	"MT",
	"NL",
	"NO",
	"PL",
	"PT",
	"RO",
	"SK",
	"SI",
	"ES",
	"SE",
	"CH",
	"GB"
];

export type TrackingConsentMode = "regional" | "always" | "immediate";
export type TrackingConsentState = "accepted" | "declined";

export interface RawTrackerConfig {
	umamiWebsiteId?: string;
	umamiScriptUrl?: string;
	consentMode?: string;
	consentCountries?: string;
	storageKey?: string;
}

export interface TrackerConfig {
	umamiWebsiteId?: string;
	umamiScriptUrl: string;
	consentMode: TrackingConsentMode;
	consentCountries: string[];
	storageKey: string;
}

export interface TrackingPageView {
	location: string;
	path: string;
	title: string;
}

interface ScriptAttributes {
	async?: boolean;
	defer?: boolean;
	crossOrigin?: string;
	dataset?: Record<string, string>;
}

type UmamiPayload = Record<string, unknown>;

interface UmamiTracker {
	track?: (payload?: string | UmamiPayload | ((props: UmamiPayload) => UmamiPayload), data?: UmamiPayload) => void;
}

declare global {
	interface Window {
		NETSTAMP_VISITOR_COUNTRY?: string;
		umami?: UmamiTracker;
	}
}

const scriptPromises = new Map<string, Promise<void>>();
const loadedTrackers = new Set<string>();
const boundPageViewEvents = new Set<string>();
const runtimeAllowedStorageKeys = new Set<string>();
let lastTrackedPageLocation = "";

export const normalizeTrackerConfig = (raw: RawTrackerConfig): TrackerConfig => {
	return {
		umamiWebsiteId: cleanValue(raw.umamiWebsiteId),
		umamiScriptUrl: cleanValue(raw.umamiScriptUrl) ?? DEFAULT_UMAMI_SCRIPT_URL,
		consentMode: normalizeConsentMode(raw.consentMode),
		consentCountries: parseCountryList(raw.consentCountries),
		storageKey: cleanValue(raw.storageKey) ?? DEFAULT_STORAGE_KEY
	};
};

export function hasEnabledTrackers(config: TrackerConfig): boolean {
	return Boolean(config.umamiWebsiteId);
}

export function readTrackingConsent(config: TrackerConfig): TrackingConsentState | null {
	if (!isBrowser()) {
		return null;
	}

	try {
		const stored = window.localStorage.getItem(config.storageKey);
		return stored === "accepted" || stored === "declined" ? stored : null;
	} catch {
		return null;
	}
}

export function writeTrackingConsent(config: TrackerConfig, state: TrackingConsentState): void {
	if (!isBrowser()) {
		return;
	}

	try {
		window.localStorage.setItem(config.storageKey, state);
	} catch {
		// Storage can be unavailable in privacy modes. The runtime still honors the in-page action.
	}
}

export function isTrackingAllowed(config: TrackerConfig): boolean {
	const stored = readTrackingConsent(config);

	if (stored === "accepted") {
		return true;
	}

	if (stored === "declined") {
		return false;
	}

	return runtimeAllowedStorageKeys.has(config.storageKey) || config.consentMode === "immediate";
}

export async function shouldRequestTrackingConsent(config: TrackerConfig): Promise<boolean> {
	if (!hasEnabledTrackers(config) || config.consentMode === "immediate") {
		return false;
	}

	if (config.consentMode === "always") {
		return true;
	}

	const country = await resolveVisitorCountry();
	return country ? config.consentCountries.includes(country) : true;
}

export async function resolveVisitorCountry(): Promise<string | null> {
	if (!isBrowser()) {
		return null;
	}

	const configuredCountry =
		normalizeCountry(window.NETSTAMP_VISITOR_COUNTRY) ??
		normalizeCountry(document.documentElement.dataset.netstampCountry) ??
		normalizeCountry(document.querySelector<HTMLMetaElement>('meta[name="netstamp-visitor-country"]')?.content);

	if (configuredCountry) {
		return configuredCountry;
	}

	if (typeof window.fetch !== "function") {
		return null;
	}

	try {
		const response = await window.fetch("/cdn-cgi/trace", {
			cache: "no-store",
			credentials: "omit"
		});

		if (!response.ok) {
			return null;
		}

		const body = await response.text();
		const country = /^loc=([A-Za-z]{2})$/m.exec(body)?.[1];
		return normalizeCountry(country);
	} catch {
		return null;
	}
}

export async function loadConfiguredTrackers(config: TrackerConfig): Promise<void> {
	if (!isBrowser() || !hasEnabledTrackers(config)) {
		return;
	}

	runtimeAllowedStorageKeys.add(config.storageKey);
	const loads: Promise<void>[] = [];

	if (config.umamiWebsiteId) {
		loads.push(loadUmami(config.umamiWebsiteId, config.umamiScriptUrl));
	}

	await Promise.allSettled(loads);
}

export function trackConfiguredPageView(config: TrackerConfig, page: TrackingPageView = currentPageView()): void {
	if (!isBrowser() || !isTrackingAllowed(config)) {
		return;
	}

	if (config.umamiWebsiteId && window.umami?.track) {
		window.umami.track(() => ({
			website: config.umamiWebsiteId,
			url: page.path,
			title: page.title
		}));
	}
}

export function bindTrackingPageViews(config: TrackerConfig, events: string[] = []): void {
	if (!isBrowser() || !hasEnabledTrackers(config)) {
		return;
	}

	const emit = () => {
		window.setTimeout(() => {
			const page = currentPageView();

			if (page.location === lastTrackedPageLocation) {
				return;
			}

			lastTrackedPageLocation = page.location;
			trackConfiguredPageView(config, page);
		}, 0);
	};

	emit();

	for (const event of events) {
		if (boundPageViewEvents.has(event)) {
			continue;
		}

		boundPageViewEvents.add(event);
		document.addEventListener(event, emit);
	}
}

function loadUmami(websiteId: string, scriptUrl: string): Promise<void> {
	const trackerKey = `umami:${websiteId}:${scriptUrl}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	return loadScript(scriptId("umami", websiteId), scriptUrl, {
		defer: true,
		dataset: {
			autoTrack: "false",
			websiteId
		}
	});
}

function loadScript(id: string, src: string, attributes: ScriptAttributes = {}): Promise<void> {
	const safeSrc = resolveHttpUrl(src);

	if (!safeSrc) {
		return Promise.resolve();
	}

	const existing = document.getElementById(id);

	if (existing) {
		return scriptPromises.get(id) ?? Promise.resolve();
	}

	const promise = new Promise<void>((resolve, reject) => {
		const script = document.createElement("script");
		script.id = id;
		script.src = safeSrc;
		script.async = Boolean(attributes.async);
		script.defer = Boolean(attributes.defer);

		if (attributes.crossOrigin) {
			script.crossOrigin = attributes.crossOrigin;
		}

		for (const [key, value] of Object.entries(attributes.dataset ?? {})) {
			script.dataset[key] = value;
		}

		script.addEventListener("load", () => resolve(), { once: true });
		script.addEventListener("error", () => reject(new Error(`Failed to load tracker script: ${safeSrc}`)), { once: true });
		document.head.append(script);
	});

	scriptPromises.set(id, promise);
	return promise;
}

function currentPageView(): TrackingPageView {
	return {
		location: window.location.href,
		path: `${window.location.pathname}${window.location.search}${window.location.hash}`,
		title: document.title
	};
}

function resolveHttpUrl(value: string): string | null {
	if (!isBrowser()) {
		return null;
	}

	try {
		const url = new URL(value, window.location.origin);

		if (url.protocol === "https:" || url.protocol === "http:") {
			return url.href;
		}
	} catch {
		return null;
	}

	return null;
}

function scriptId(prefix: string, value: string): string {
	return `netstamp-${prefix}-${value.replace(/[^a-z0-9_-]/gi, "-")}`;
}

function cleanValue(value: string | null | undefined): string | undefined {
	const cleaned = value?.trim();
	return cleaned ? cleaned : undefined;
}

function normalizeConsentMode(value: string | null | undefined): TrackingConsentMode {
	const mode = cleanValue(value)?.toLowerCase();

	if (mode === "always" || mode === "immediate") {
		return mode;
	}

	return "regional";
}

function parseCountryList(value: string | null | undefined): string[] {
	const countries = cleanValue(value)
		?.split(/[\s,]+/)
		.map(normalizeCountry)
		.filter((country): country is string => Boolean(country));

	return countries?.length ? countries : [...DEFAULT_TRACKING_CONSENT_COUNTRIES];
}

function normalizeCountry(value: string | null | undefined): string | null {
	const country = cleanValue(value)?.toUpperCase();
	return country && /^[A-Z]{2}$/.test(country) ? country : null;
}

function isBrowser(): boolean {
	return typeof window !== "undefined" && typeof document !== "undefined";
}
