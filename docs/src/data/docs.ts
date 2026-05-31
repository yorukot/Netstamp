import { defaultLocale, docsBasePath, getSectionLabel, localeFromSegment, locales, type Locale } from "@/i18n/config";
import { getCollection, type CollectionEntry } from "astro:content";

export type DocIcon = "activity" | "api" | "bolt" | "book" | "code" | "compass" | "cube" | "database" | "deployment" | "key" | "map" | "route" | "server" | "shield" | "terminal" | "users" | "wrench";

export interface DocNavItem {
	title: string;
	href: string;
	icon: DocIcon;
	description: string;
}

export interface DocPage extends DocNavItem {
	locale: Locale;
	logicalPath: string;
	sectionKey: string;
	section: string;
	order: number;
	editPath: string;
	translated: boolean;
}

export interface DocNavGroup {
	key: string;
	title: string;
	items: DocNavItem[];
}

export interface SearchEntry extends DocNavItem {
	keywords: string;
	content: string;
}

export interface LocaleAlternate {
	locale: Locale;
	href: string;
	available: boolean;
}

export interface ResolvedDoc {
	locale: Locale;
	logicalPath: string;
	slug: string;
	href: string;
	entry: DocsEntry;
	translated: boolean;
	title: string;
	description: string;
	icon: DocIcon;
	sectionKey: string;
	editPath: string;
}

export const githubBaseUrl = "https://github.com/yorukot/netstamp/blob/main";

type DocsEntry = CollectionEntry<"docs">;

const docIconNames = new Set<DocIcon>([
	"activity",
	"api",
	"bolt",
	"book",
	"code",
	"compass",
	"cube",
	"database",
	"deployment",
	"key",
	"map",
	"route",
	"server",
	"shield",
	"terminal",
	"users",
	"wrench"
]);

// Section ordering is keyed by the canonical (English) section key so the
// ordering stays identical across every locale; only the display label is
// localized (see `getSectionLabel`).
const sectionOrder = new Map([
	["Start", 0],
	["Guides", 10],
	["Reference", 20]
]);

const allEntries = await getCollection("docs", entry => !entry.data.draft);

function searchContentFromSource(source: string) {
	return source
		.replace(/^---\s*\n[\s\S]*?\n---/, " ")
		.replace(/```[A-Za-z0-9_-]*\s*/g, " ")
		.replace(/import\s+[^;]+;?/g, " ")
		.replace(/<[^>]+>/g, " ")
		.replace(/[{}#[\]()`*_~>|-]/g, " ")
		.replace(/\s+/g, " ")
		.trim()
		.toLowerCase();
}

function toTitleCase(value: string) {
	return value
		.split(/[-_\s]+/)
		.filter(Boolean)
		.map(part => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
		.join(" ");
}

// Reduce an entry id to its locale-independent logical path. The default
// locale lives at the content root; other locales mirror the same paths under
// a `<locale>/` folder, so the leading segment is stripped when it is a locale.
function normalizeLogicalPath(value: string) {
	const withoutExt = value.replace(/\.mdx$/, "");
	if (withoutExt === "index" || withoutExt === "") return "index";
	const withoutIndex = withoutExt.replace(/\/index$/, "");
	return withoutIndex === "" ? "index" : withoutIndex;
}

function splitEntryLocale(entry: DocsEntry): { locale: Locale; logicalPath: string } {
	const id = entry.id;
	const slashIndex = id.indexOf("/");
	const head = slashIndex === -1 ? id.replace(/\.mdx$/, "") : id.slice(0, slashIndex);
	const matched = localeFromSegment(head);

	if (matched && matched !== defaultLocale) {
		const rest = slashIndex === -1 ? "" : id.slice(slashIndex + 1);
		return { locale: matched, logicalPath: normalizeLogicalPath(rest) };
	}

	return { locale: defaultLocale, logicalPath: normalizeLogicalPath(id) };
}

function routeForLogicalPath(locale: Locale, logicalPath: string) {
	const base = docsBasePath(locale);
	return logicalPath === "index" ? `${base}/` : `${base}/${logicalPath}/`;
}

function editPathForEntry(entry: DocsEntry) {
	if (entry.data.editPath) return entry.data.editPath;
	if (entry.filePath?.startsWith("docs/")) return entry.filePath;
	if (entry.filePath) return `docs/${entry.filePath}`;
	return `docs/src/content/docs/${entry.id.endsWith(".mdx") ? entry.id : `${entry.id}.mdx`}`;
}

function fallbackTitleFromLogicalPath(logicalPath: string) {
	if (logicalPath === "index") return "Overview";
	const slug = logicalPath.split("/").filter(Boolean).at(-1) ?? "overview";
	return toTitleCase(slug);
}

function sectionKeyForLogicalPath(logicalPath: string, navSection?: string) {
	if (navSection) return navSection;
	if (logicalPath === "index") return "Start";
	return toTitleCase(logicalPath.split("/")[0]);
}

function iconForEntry(entry: DocsEntry, logicalPath: string): DocIcon {
	const value = entry.data.icon;
	if (value && docIconNames.has(value as DocIcon)) return value as DocIcon;
	if (logicalPath === "index") return "compass";
	if (logicalPath.includes("guides")) return "bolt";
	if (logicalPath.includes("api")) return "api";
	if (logicalPath.includes("config")) return "wrench";
	if (logicalPath.includes("deploy")) return "deployment";
	if (logicalPath.includes("architecture")) return "route";
	if (logicalPath.includes("ui")) return "cube";

	return "book";
}

function navOrderForEntry(entry: DocsEntry, logicalPath: string) {
	const order = entry.data.navOrder ?? entry.data.order;
	if (typeof order === "number" && Number.isFinite(order)) return order;
	return logicalPath === "index" ? 0 : 100;
}

// Group every entry by locale, keyed by logical path. The default locale's
// entries define the canonical set of pages every locale is expected to have.
const entriesByLocale = new Map<Locale, Map<string, DocsEntry>>(locales.map(locale => [locale, new Map<string, DocsEntry>()]));

for (const entry of allEntries) {
	const { locale, logicalPath } = splitEntryLocale(entry);
	entriesByLocale.get(locale)?.set(logicalPath, entry);
}

const defaultEntries = entriesByLocale.get(defaultLocale) ?? new Map<string, DocsEntry>();
const canonicalLogicalPaths = Array.from(defaultEntries.keys());

function resolveDoc(locale: Locale, logicalPath: string): ResolvedDoc | undefined {
	const localized = entriesByLocale.get(locale)?.get(logicalPath);
	const entry = localized ?? defaultEntries.get(logicalPath);
	if (!entry) return undefined;

	return {
		locale,
		logicalPath,
		slug: logicalPath,
		href: routeForLogicalPath(locale, logicalPath),
		entry,
		translated: locale === defaultLocale || Boolean(localized),
		title: entry.data.navTitle ?? entry.data.title ?? fallbackTitleFromLogicalPath(logicalPath),
		description: entry.data.description ?? "Netstamp documentation page.",
		icon: iconForEntry(entry, logicalPath),
		sectionKey: sectionKeyForLogicalPath(logicalPath, entry.data.navSection),
		editPath: editPathForEntry(entry)
	};
}

function comparePages(a: DocPage, b: DocPage) {
	const sectionDelta = (sectionOrder.get(a.sectionKey) ?? 100) - (sectionOrder.get(b.sectionKey) ?? 100) || a.sectionKey.localeCompare(b.sectionKey);
	if (sectionDelta) return sectionDelta;

	return a.order - b.order || a.title.localeCompare(b.title);
}

const pagesCache = new Map<Locale, DocPage[]>();
const navCache = new Map<Locale, DocNavGroup[]>();
const searchCache = new Map<Locale, SearchEntry[]>();

export function getDocsPages(locale: Locale = defaultLocale): DocPage[] {
	const cached = pagesCache.get(locale);
	if (cached) return cached;

	const pages = canonicalLogicalPaths
		.map(logicalPath => resolveDoc(locale, logicalPath))
		.filter((doc): doc is ResolvedDoc => Boolean(doc))
		.map(doc => ({
			title: doc.title,
			href: doc.href,
			icon: doc.icon,
			description: doc.description,
			locale: doc.locale,
			logicalPath: doc.logicalPath,
			sectionKey: doc.sectionKey,
			section: getSectionLabel(locale, doc.sectionKey),
			order: navOrderForEntry(doc.entry, doc.logicalPath),
			editPath: doc.editPath,
			translated: doc.translated
		}))
		.sort(comparePages);

	pagesCache.set(locale, pages);
	return pages;
}

export function getDocsNav(locale: Locale = defaultLocale): DocNavGroup[] {
	const cached = navCache.get(locale);
	if (cached) return cached;

	const groups = new Map<string, DocNavGroup>();

	for (const page of getDocsPages(locale)) {
		const group = groups.get(page.sectionKey) ?? { key: page.sectionKey, title: page.section, items: [] };
		group.items.push({ title: page.title, href: page.href, icon: page.icon, description: page.description });
		groups.set(page.sectionKey, group);
	}

	const nav = Array.from(groups.values());
	navCache.set(locale, nav);
	return nav;
}

export function getDocsSearchIndex(locale: Locale = defaultLocale): SearchEntry[] {
	const cached = searchCache.get(locale);
	if (cached) return cached;

	const bodyByHref = new Map<string, string>();
	for (const logicalPath of canonicalLogicalPaths) {
		const doc = resolveDoc(locale, logicalPath);
		if (doc) bodyByHref.set(doc.href, doc.entry.body ?? "");
	}

	const index = getDocsNav(locale).flatMap(group =>
		group.items.map(item => ({
			...item,
			keywords: `${group.title} ${item.title} ${item.description} ${item.href}`.toLowerCase(),
			content: searchContentFromSource(bodyByHref.get(item.href) ?? "")
		}))
	);

	searchCache.set(locale, index);
	return index;
}

// The doc rendered at the docs landing route for a locale (translated when a
// translation exists, otherwise the default-locale fallback).
export function getIndexDoc(locale: Locale): ResolvedDoc | undefined {
	return resolveDoc(locale, "index");
}

// Every non-index doc route for a locale, used to drive `getStaticPaths`.
export function getDocRoutes(locale: Locale): ResolvedDoc[] {
	return canonicalLogicalPaths
		.filter(logicalPath => logicalPath !== "index")
		.map(logicalPath => resolveDoc(locale, logicalPath))
		.filter((doc): doc is ResolvedDoc => Boolean(doc));
}

// Switch targets for the language switcher. Every locale resolves to a real
// generated page (translation or fallback), so all are navigable; `available`
// flags whether the target is an actual translation.
export function getLocaleAlternates(logicalPath: string): LocaleAlternate[] {
	return locales.map(locale => {
		const localized = entriesByLocale.get(locale)?.get(logicalPath);
		const available = locale === defaultLocale ? defaultEntries.has(logicalPath) : Boolean(localized);
		return { locale, href: routeForLogicalPath(locale, logicalPath), available };
	});
}
