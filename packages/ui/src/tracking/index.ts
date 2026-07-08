const DEFAULT_POSTHOG_HOST = "https://us.i.posthog.com";
const DEFAULT_PLAUSIBLE_SCRIPT_URL = "https://plausible.io/js/script.js";
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

export type TrackingConsentMode = "regional" | "always" | "never";
export type TrackingConsentState = "accepted" | "declined";

export interface RawTrackerConfig {
	googleTagId?: string;
	gaMeasurementId?: string;
	clarityProjectId?: string;
	metaPixelId?: string;
	facebookPixelId?: string;
	posthogKey?: string;
	posthogHost?: string;
	plausibleDomain?: string;
	plausibleScriptUrl?: string;
	umamiWebsiteId?: string;
	umamiScriptUrl?: string;
	consentMode?: string;
	consentCountries?: string;
	storageKey?: string;
}

export interface TrackerConfig {
	googleTagId?: string;
	clarityProjectId?: string;
	metaPixelId?: string;
	posthogKey?: string;
	posthogHost: string;
	plausibleDomain?: string;
	plausibleScriptUrl: string;
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

interface ClarityQueue {
	(...args: unknown[]): void;
	q?: unknown[][];
}

interface PostHogQueue extends Array<unknown> {
	[key: string]: unknown;
	_i?: unknown[];
	__SV?: number;
	init?: (token: string, options: Record<string, unknown>, name?: string) => void;
	people?: PostHogQueue;
	capture?: (...args: unknown[]) => void;
	toString: (stub?: unknown) => string;
}

interface MetaPixelQueue {
	(...args: unknown[]): void;
	callMethod?: (...args: unknown[]) => void;
	queue?: unknown[][];
	push?: MetaPixelQueue;
	loaded?: boolean;
	version?: string;
	disablePushState?: boolean;
}

interface PlausibleTracker {
	(...args: unknown[]): void;
	init?: (options: Record<string, unknown>) => void;
	o?: Record<string, unknown>;
	q?: unknown[][];
}

type UmamiPayload = Record<string, unknown>;

interface UmamiTracker {
	track?: (payload?: string | UmamiPayload | ((props: UmamiPayload) => UmamiPayload), data?: UmamiPayload) => void;
}

declare global {
	interface Window {
		NETSTAMP_VISITOR_COUNTRY?: string;
		_fbq?: MetaPixelQueue;
		clarity?: ClarityQueue;
		dataLayer?: unknown[][];
		fbq?: MetaPixelQueue;
		gtag?: (...args: unknown[]) => void;
		plausible?: PlausibleTracker;
		posthog?: PostHogQueue;
		umami?: UmamiTracker;
	}
}

const scriptPromises = new Map<string, Promise<void>>();
const loadedTrackers = new Set<string>();
const boundPageViewEvents = new Set<string>();
const runtimeAllowedStorageKeys = new Set<string>();
let lastTrackedPageLocation = "";

export function normalizeTrackerConfig(raw: RawTrackerConfig): TrackerConfig {
	return {
		googleTagId: cleanValue(raw.googleTagId) ?? cleanValue(raw.gaMeasurementId),
		clarityProjectId: cleanValue(raw.clarityProjectId),
		metaPixelId: cleanValue(raw.metaPixelId) ?? cleanValue(raw.facebookPixelId),
		posthogKey: cleanValue(raw.posthogKey),
		posthogHost: cleanValue(raw.posthogHost) ?? DEFAULT_POSTHOG_HOST,
		plausibleDomain: cleanValue(raw.plausibleDomain),
		plausibleScriptUrl: cleanValue(raw.plausibleScriptUrl) ?? DEFAULT_PLAUSIBLE_SCRIPT_URL,
		umamiWebsiteId: cleanValue(raw.umamiWebsiteId),
		umamiScriptUrl: cleanValue(raw.umamiScriptUrl) ?? DEFAULT_UMAMI_SCRIPT_URL,
		consentMode: normalizeConsentMode(raw.consentMode),
		consentCountries: parseCountryList(raw.consentCountries),
		storageKey: cleanValue(raw.storageKey) ?? DEFAULT_STORAGE_KEY
	};
}

export function hasEnabledTrackers(config: TrackerConfig): boolean {
	return Boolean(config.googleTagId || config.clarityProjectId || config.metaPixelId || config.posthogKey || config.plausibleDomain || config.umamiWebsiteId);
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

	return runtimeAllowedStorageKeys.has(config.storageKey) || config.consentMode === "never";
}

export async function shouldRequestTrackingConsent(config: TrackerConfig): Promise<boolean> {
	if (!hasEnabledTrackers(config) || config.consentMode === "never") {
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

	if (config.googleTagId) {
		loads.push(loadGoogleTag(config.googleTagId));
	}

	if (config.clarityProjectId) {
		loads.push(loadClarity(config.clarityProjectId));
	}

	if (config.metaPixelId) {
		loads.push(loadMetaPixel(config.metaPixelId));
	}

	if (config.posthogKey) {
		loads.push(loadPostHog(config.posthogKey, config.posthogHost));
	}

	if (config.plausibleDomain) {
		loads.push(loadPlausible(config.plausibleDomain, config.plausibleScriptUrl));
	}

	if (config.umamiWebsiteId) {
		loads.push(loadUmami(config.umamiWebsiteId, config.umamiScriptUrl));
	}

	await Promise.allSettled(loads);
}

export function trackConfiguredPageView(config: TrackerConfig, page: TrackingPageView = currentPageView()): void {
	if (!isBrowser() || !isTrackingAllowed(config)) {
		return;
	}

	if (config.googleTagId && window.gtag) {
		window.gtag("event", "page_view", {
			page_location: page.location,
			page_path: page.path,
			page_title: page.title
		});
	}

	if (config.posthogKey && window.posthog?.capture) {
		window.posthog.capture("$pageview", {
			$current_url: page.location,
			page_path: page.path,
			page_title: page.title
		});
	}

	if (config.metaPixelId && window.fbq) {
		window.fbq("track", "PageView", {
			page_location: page.location,
			page_path: page.path,
			page_title: page.title
		});
	}

	if (config.plausibleDomain && window.plausible) {
		window.plausible("pageview", {
			url: page.location
		});
	}

	if (config.umamiWebsiteId && window.umami?.track) {
		window.umami.track(props => ({
			...props,
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

function loadGoogleTag(tagId: string): Promise<void> {
	const trackerKey = `google:${tagId}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	window.dataLayer = window.dataLayer ?? [];
	window.gtag =
		window.gtag ??
		function gtag(...args: unknown[]) {
			window.dataLayer?.push(args);
		};
	window.gtag("js", new Date());
	window.gtag("config", tagId, { send_page_view: false });

	const scriptUrl = new URL("https://www.googletagmanager.com/gtag/js");
	scriptUrl.searchParams.set("id", tagId);
	return loadScript(scriptId("google-tag", tagId), scriptUrl.href, { async: true });
}

function loadClarity(projectId: string): Promise<void> {
	const trackerKey = `clarity:${projectId}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	window.clarity =
		window.clarity ??
		function clarity(...args: unknown[]) {
			window.clarity = window.clarity ?? clarity;
			window.clarity.q = window.clarity.q ?? [];
			window.clarity.q.push(args);
		};

	return loadScript(scriptId("clarity", projectId), `https://www.clarity.ms/tag/${encodeURIComponent(projectId)}`, { async: true });
}

function loadMetaPixel(pixelId: string): Promise<void> {
	const trackerKey = `meta-pixel:${pixelId}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	installMetaPixelStub()?.("init", pixelId);
	return loadScript(scriptId("meta-pixel", pixelId), "https://connect.facebook.net/en_US/fbevents.js", { async: true });
}

function loadPostHog(projectKey: string, apiHost: string): Promise<void> {
	const safeApiHost = resolveHttpUrl(apiHost);

	if (!safeApiHost) {
		return Promise.resolve();
	}

	const trackerKey = `posthog:${projectKey}:${safeApiHost}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	installPostHogStub()?.init?.(projectKey, {
		api_host: safeApiHost.replace(/\/$/, ""),
		capture_pageview: false,
		person_profiles: "identified_only"
	});

	return Promise.resolve();
}

function loadPlausible(domain: string, scriptUrl: string): Promise<void> {
	const trackerKey = `plausible:${domain}:${scriptUrl}`;

	if (loadedTrackers.has(trackerKey)) {
		return Promise.resolve();
	}

	loadedTrackers.add(trackerKey);
	installPlausibleStub()?.init?.({
		autoCapturePageviews: false
	});

	return loadScript(scriptId("plausible", domain), scriptUrl, {
		defer: true,
		dataset: {
			domain
		}
	});
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

function installMetaPixelStub(): MetaPixelQueue | null {
	if (!isBrowser()) {
		return null;
	}

	if (window.fbq?.loaded) {
		return window.fbq;
	}

	const fbq =
		window.fbq ??
		(function fbq(...args: unknown[]) {
			const target = window.fbq;

			if (target?.callMethod) {
				target.callMethod(...args);
				return;
			}

			if (target) {
				target.queue = target.queue ?? [];
				target.queue.push(args);
			}
		} as MetaPixelQueue);

	window.fbq = fbq;
	window._fbq = fbq;
	fbq.push = fbq;
	fbq.loaded = true;
	fbq.version = "2.0";
	fbq.disablePushState = true;
	fbq.queue = fbq.queue ?? [];

	return fbq;
}

function installPlausibleStub(): PlausibleTracker | null {
	if (!isBrowser()) {
		return null;
	}

	const plausible =
		window.plausible ??
		(function plausible(...args: unknown[]) {
			const target = window.plausible;

			if (target) {
				target.q = target.q ?? [];
				target.q.push(args);
			}
		} as PlausibleTracker);

	window.plausible = plausible;
	plausible.init =
		plausible.init ??
		((options: Record<string, unknown>) => {
			plausible.o = options || {};
		});

	return plausible;
}

function installPostHogStub(): PostHogQueue | null {
	if (!isBrowser()) {
		return null;
	}

	if (window.posthog?.__SV) {
		return window.posthog;
	}

	const posthog = window.posthog ?? ([] as unknown as PostHogQueue);
	window.posthog = posthog;
	posthog._i = [];
	posthog.init = (token, options, name) => {
		const apiHost = typeof options.api_host === "string" ? options.api_host : DEFAULT_POSTHOG_HOST;
		const script = document.createElement("script");
		const scriptSrc = resolveHttpUrl(`${apiHost.replace(".i.posthog.com", "-assets.i.posthog.com").replace(/\/$/, "")}/static/array.js`);

		if (scriptSrc) {
			script.type = "text/javascript";
			script.crossOrigin = "anonymous";
			script.async = true;
			script.src = scriptSrc;
			insertBeforeFirstScript(script);
		}

		let target = posthog;

		if (name) {
			target = [] as unknown as PostHogQueue;
			posthog[name] = target;
		}

		target.people = target.people ?? ([] as unknown as PostHogQueue);
		target.toString = stub => {
			const instanceName = name ?? "posthog";
			return instanceName === "posthog" ? `posthog${stub ? "" : " (stub)"}` : `posthog.${instanceName}${stub ? "" : " (stub)"}`;
		};
		target.people.toString = () => `${target.toString(true)}.people (stub)`;

		for (const method of POSTHOG_STUB_METHODS) {
			assignPostHogMethod(target, method);
		}

		posthog._i?.push([token, options, name]);
	};
	posthog.__SV = 1;

	return posthog;
}

const POSTHOG_STUB_METHODS = [
	"init",
	"capture",
	"register",
	"register_once",
	"register_for_session",
	"unregister",
	"unregister_for_session",
	"getFeatureFlag",
	"getFeatureFlagPayload",
	"isFeatureEnabled",
	"reloadFeatureFlags",
	"updateEarlyAccessFeatureEnrollment",
	"getEarlyAccessFeatures",
	"on",
	"onFeatureFlags",
	"onSessionId",
	"getSurveys",
	"getActiveMatchingSurveys",
	"renderSurvey",
	"canRenderSurvey",
	"getNextSurveyStep",
	"identify",
	"setPersonProperties",
	"group",
	"resetGroups",
	"setPersonPropertiesForFlags",
	"resetPersonPropertiesForFlags",
	"setGroupPropertiesForFlags",
	"resetGroupPropertiesForFlags",
	"reset",
	"get_distinct_id",
	"getGroups",
	"get_session_id",
	"get_session_replay_url",
	"alias",
	"set_config",
	"startSessionRecording",
	"stopSessionRecording",
	"sessionRecordingStarted",
	"captureException",
	"loadToolbar",
	"get_property",
	"getSessionProperty",
	"createPersonProfile",
	"opt_in_capturing",
	"opt_out_capturing",
	"has_opted_in_capturing",
	"has_opted_out_capturing",
	"clear_opt_in_out_capturing",
	"debug"
];

function assignPostHogMethod(target: PostHogQueue, method: string): void {
	target[method] = (...args: unknown[]) => {
		target.push([method, ...args]);
	};
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

function insertBeforeFirstScript(script: HTMLScriptElement): void {
	const firstScript = document.getElementsByTagName("script")[0];
	firstScript?.parentNode?.insertBefore(script, firstScript) ?? document.head.append(script);
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

	if (mode === "always" || mode === "never") {
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
