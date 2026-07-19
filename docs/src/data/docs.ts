import { getCollection, type CollectionEntry } from "astro:content";
import { localeFromAstro, pathForLocale, uiForLocale } from "../i18n/ui";

export type DocIcon = "activity" | "api" | "bolt" | "book" | "code" | "codeBlock" | "compass" | "cube" | "database" | "deployment" | "key" | "map" | "route" | "server" | "shield" | "users" | "wrench";

export interface DocNavItem {
	title: string;
	href: string;
	icon: DocIcon;
	description: string;
}

export interface DocPage extends DocNavItem {
	section: string;
	sectionKey: "guides" | "reference" | "start";
	order: number;
	editPath: string;
	contentId: string;
	translated: boolean;
}

export interface DocNavGroup {
	title: string;
	items: DocNavItem[];
}

export interface SearchEntry extends DocNavItem {
	keywords: string;
	content: string;
}

export const githubBaseUrl = "https://github.com/yorukot/netstamp/blob/main";

type DocsEntry = CollectionEntry<"docs">;

const docIconNames = new Set<DocIcon>([
	"activity",
	"api",
	"bolt",
	"book",
	"code",
	"codeBlock",
	"compass",
	"cube",
	"database",
	"deployment",
	"key",
	"map",
	"route",
	"server",
	"shield",
	"users",
	"wrench"
]);
const docEntries = await getCollection("docs", entry => !entry.data.draft);

const searchContentFromSource = (source: string) =>
	source
		.replace(/^---\s*\n[\s\S]*?\n---/, " ")
		.replace(/```[A-Za-z0-9_-]*\s*/g, " ")
		.replace(/import\s+[^;]+;?/g, " ")
		.replace(/<[^>]+>/g, " ")
		.replace(/[{}#[\]()`*_~>|-]/g, " ")
		.replace(/\s+/g, " ")
		.trim()
		.toLowerCase();

const entryLocale = (entry: DocsEntry) => (entry.id.toLowerCase() === "zh-tw" || entry.id.toLowerCase().startsWith("zh-tw/") ? "zh-TW" : "en");
const contentIdFromEntry = (entry: DocsEntry) => {
	if (entry.id.toLowerCase() === "en" || entry.id.toLowerCase() === "zh-tw") return "index";
	return entry.id
		.replace(/^(en|zh-tw)\//i, "")
		.replace(/\.mdx$/, "")
		.replace(/\/index$/, "");
};
const sectionKeyFromContentId = (contentId: string): DocPage["sectionKey"] => (contentId === "index" ? "start" : contentId.startsWith("guides/") ? "guides" : "reference");
const routeFromContentId = (contentId: string, locale: string | undefined) => pathForLocale(contentId === "index" ? "/docs/" : `/docs/${contentId}/`, locale);

const editPathFromEntry = (entry: DocsEntry) => {
	if (entry.data.editPath) return entry.data.editPath;
	if (entry.filePath?.startsWith("docs/")) return entry.filePath;
	if (entry.filePath) return `docs/${entry.filePath}`;
	return `docs/src/content/docs/${entry.id.endsWith(".mdx") ? entry.id : `${entry.id}.mdx`}`;
};

const iconFromEntry = (entry: DocsEntry, contentId: string): DocIcon => {
	const value = entry.data.icon;
	if (value && docIconNames.has(value as DocIcon)) return value as DocIcon;
	if (contentId === "index") return "compass";
	if (contentId.startsWith("guides/")) return "bolt";
	if (contentId.includes("api")) return "api";
	if (contentId.includes("config")) return "wrench";
	if (contentId.includes("deploy")) return "deployment";
	if (contentId.includes("architecture")) return "route";
	if (contentId.includes("ui")) return "cube";
	return "book";
};

const navOrderFromEntry = (entry: DocsEntry, contentId: string) => {
	const order = entry.data.navOrder ?? entry.data.order;
	if (typeof order === "number" && Number.isFinite(order)) return order;
	return contentId === "index" ? 0 : 100;
};

const localizedEntryFor = (contentId: string, locale: string | undefined) => {
	const resolvedLocale = localeFromAstro(locale);
	const localized = docEntries.find(entry => entryLocale(entry) === resolvedLocale && contentIdFromEntry(entry) === contentId);
	const english = docEntries.find(entry => entryLocale(entry) === "en" && contentIdFromEntry(entry) === contentId);
	return { entry: localized ?? english, translated: Boolean(localized), english };
};

export const getDocEntry = (contentId: string, locale: string | undefined) => localizedEntryFor(contentId, locale);

export const getDocsPages = (locale: string | undefined): DocPage[] => {
	const resolvedLocale = localeFromAstro(locale);
	const ui = uiForLocale(resolvedLocale);
	const englishEntries = docEntries.filter(entry => entryLocale(entry) === "en");

	return englishEntries
		.map(englishEntry => {
			const contentId = contentIdFromEntry(englishEntry);
			const { entry = englishEntry, translated } = localizedEntryFor(contentId, resolvedLocale);
			const sectionKey = sectionKeyFromContentId(contentId);
			return {
				title: entry.data.navTitle ?? entry.data.title,
				href: routeFromContentId(contentId, resolvedLocale),
				icon: iconFromEntry(entry, contentId),
				description: entry.data.description ?? ui.meta.docsDescription,
				section: ui.docs.sections[sectionKey],
				sectionKey,
				order: navOrderFromEntry(englishEntry, contentId),
				editPath: editPathFromEntry(entry),
				contentId,
				translated
			};
		})
		.sort((a, b) => {
			const order = { start: 0, guides: 10, reference: 20 } as const;
			return order[a.sectionKey] - order[b.sectionKey] || a.order - b.order || a.title.localeCompare(b.title, resolvedLocale);
		});
};

export const getDocsNav = (locale: string | undefined): DocNavGroup[] =>
	Array.from(
		getDocsPages(locale).reduce((groups, page) => {
			const items = groups.get(page.section) ?? [];
			items.push({ title: page.title, href: page.href, icon: page.icon, description: page.description });
			groups.set(page.section, items);
			return groups;
		}, new Map<string, DocNavItem[]>())
	).map(([title, items]) => ({ title, items }));

export const getDocsSearchIndex = (locale: string | undefined): SearchEntry[] => {
	const pages = getDocsPages(locale);
	const entriesByContentId = new Map(pages.map(page => [page.contentId, localizedEntryFor(page.contentId, locale).entry]));
	return pages.map(page => ({
		title: page.title,
		href: page.href,
		icon: page.icon,
		description: page.description,
		keywords: `${page.section} ${page.title} ${page.description} ${page.href}`.toLowerCase(),
		content: searchContentFromSource(entriesByContentId.get(page.contentId)?.body ?? "")
	}));
};
