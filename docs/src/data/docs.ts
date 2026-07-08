import { getCollection, type CollectionEntry } from "astro:content";

export type DocIcon = "activity" | "api" | "bolt" | "book" | "code" | "codeBlock" | "compass" | "cube" | "database" | "deployment" | "key" | "map" | "route" | "server" | "shield" | "users" | "wrench";

export interface DocNavItem {
	title: string;
	href: string;
	icon: DocIcon;
	description: string;
}

export interface DocPage extends DocNavItem {
	section: string;
	order: number;
	editPath: string;
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

const sectionOrder = new Map([
	["Start", 0],
	["Guides", 10],
	["Reference", 20]
]);

const docEntries = await getCollection("docs", entry => !entry.data.draft);

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

function normalizedEntryId(entryId: string) {
	return entryId.replace(/\.mdx$/, "").replace(/\/index$/, "");
}

function routeFromEntryId(entryId: string) {
	const route = `/docs/${normalizedEntryId(entryId)}`.replace(/\.mdx$/, "").replace(/\/index$/, "/");
	return route.endsWith("/") ? route : `${route}/`;
}

function editPathFromEntry(entry: DocsEntry) {
	if (entry.data.editPath) return entry.data.editPath;
	if (entry.filePath?.startsWith("docs/")) return entry.filePath;
	if (entry.filePath) return `docs/${entry.filePath}`;
	return `docs/src/content/docs/${entry.id.endsWith(".mdx") ? entry.id : `${entry.id}.mdx`}`;
}

function fallbackTitleFromHref(href: string) {
	const parts = href.split("/").filter(Boolean);
	const slug = parts.at(-1) ?? "overview";
	return slug === "docs" ? "Overview" : toTitleCase(slug);
}

function sectionFromHref(href: string) {
	if (href === "/docs/") return "Start";

	const parts = href.split("/").filter(Boolean);
	return toTitleCase(parts[1] ?? "Docs");
}

function iconFromEntry(entry: DocsEntry, href: string): DocIcon {
	const value = entry.data.icon;
	if (value && docIconNames.has(value as DocIcon)) return value as DocIcon;
	if (href === "/docs/") return "compass";
	if (href.includes("/guides/")) return "bolt";
	if (href.includes("api")) return "api";
	if (href.includes("config")) return "wrench";
	if (href.includes("deploy")) return "deployment";
	if (href.includes("architecture")) return "route";
	if (href.includes("ui")) return "cube";

	return "book";
}

function navOrderFromEntry(entry: DocsEntry, href: string) {
	const order = entry.data.navOrder ?? entry.data.order;
	if (typeof order === "number" && Number.isFinite(order)) return order;
	return href === "/docs/" ? 0 : 100;
}

export const docsPages: DocPage[] = docEntries
	.map(entry => {
		const href = routeFromEntryId(entry.id);
		const section = entry.data.navSection ?? sectionFromHref(href);

		return {
			title: entry.data.navTitle ?? entry.data.title ?? fallbackTitleFromHref(href),
			href,
			icon: iconFromEntry(entry, href),
			description: entry.data.description ?? "Netstamp documentation page.",
			section,
			order: navOrderFromEntry(entry, href),
			editPath: editPathFromEntry(entry)
		};
	})
	.sort((a, b) => {
		const sectionDelta = (sectionOrder.get(a.section) ?? 100) - (sectionOrder.get(b.section) ?? 100) || a.section.localeCompare(b.section);
		if (sectionDelta) return sectionDelta;

		return a.order - b.order || a.title.localeCompare(b.title);
	});

export const docsNav: DocNavGroup[] = Array.from(
	docsPages.reduce((groups, page) => {
		const items = groups.get(page.section) ?? [];
		items.push({
			title: page.title,
			href: page.href,
			icon: page.icon,
			description: page.description
		});
		groups.set(page.section, items);

		return groups;
	}, new Map<string, DocNavItem[]>())
).map(([title, items]) => ({ title, items }));

const docContentByHref = new Map(docEntries.map(entry => [routeFromEntryId(entry.id), entry.body ?? ""]));

export const docsSearchIndex: SearchEntry[] = docsNav.flatMap(group =>
	group.items.map(item => ({
		...item,
		keywords: `${group.title} ${item.title} ${item.description} ${item.href}`.toLowerCase(),
		content: searchContentFromSource(docContentByHref.get(item.href) ?? "")
	}))
);
