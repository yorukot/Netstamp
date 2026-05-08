export type DocIcon = "activity" | "api" | "bolt" | "book" | "code" | "compass" | "cube" | "database" | "deployment" | "key" | "map" | "route" | "server" | "shield" | "terminal" | "users" | "wrench";

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

interface DocFrontmatter {
	title?: string;
	description?: string;
	editPath?: string;
	navTitle?: string;
	navSection?: string;
	navOrder?: string;
	order?: string;
	icon?: string;
	draft?: string;
}

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

const sectionOrder = new Map([
	["Start", 0],
	["Guides", 10],
	["Reference", 20]
]);

const docFiles = import.meta.glob<string>("../content/docs/**/*.mdx", {
	eager: true,
	import: "default",
	query: "?raw"
});

function cleanFrontmatterValue(value: string) {
	const trimmed = value.trim();
	if ((trimmed.startsWith('"') && trimmed.endsWith('"')) || (trimmed.startsWith("'") && trimmed.endsWith("'"))) {
		return trimmed.slice(1, -1);
	}

	return trimmed;
}

function parseFrontmatter(source: string): DocFrontmatter {
	const match = source.match(/^---\s*\n([\s\S]*?)\n---/);
	if (!match) return {};

	return match[1].split(/\r?\n/).reduce<DocFrontmatter>((frontmatter, line) => {
		const entry = line.match(/^([A-Za-z][\w-]*):\s*(.*)$/);
		if (!entry) return frontmatter;

		frontmatter[entry[1] as keyof DocFrontmatter] = cleanFrontmatterValue(entry[2]);
		return frontmatter;
	}, {});
}

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

function routeFromFilePath(filePath: string) {
	const route = filePath
		.replace("../content/docs", "/docs")
		.replace(/\.mdx$/, "")
		.replace(/\/index$/, "/");
	return route.endsWith("/") ? route : `${route}/`;
}

function editPathFromFilePath(filePath: string) {
	return filePath.replace("../content/", "docs/src/content/");
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

function iconFromFrontmatter(value: string | undefined, href: string): DocIcon {
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

function navOrderFromFrontmatter(frontmatter: DocFrontmatter, href: string) {
	const order = Number(frontmatter.navOrder ?? frontmatter.order);
	if (Number.isFinite(order)) return order;
	return href === "/docs/" ? 0 : 100;
}

export const docsPages: DocPage[] = Object.entries(docFiles)
	.map(([filePath, source]) => {
		const frontmatter = parseFrontmatter(source);
		const href = routeFromFilePath(filePath);
		const section = frontmatter.navSection ?? sectionFromHref(href);

		return {
			title: frontmatter.navTitle ?? frontmatter.title ?? fallbackTitleFromHref(href),
			href,
			icon: iconFromFrontmatter(frontmatter.icon, href),
			description: frontmatter.description ?? "Netstamp documentation page.",
			section,
			order: navOrderFromFrontmatter(frontmatter, href),
			editPath: frontmatter.editPath ?? editPathFromFilePath(filePath),
			draft: frontmatter.draft === "true"
		};
	})
	.filter(page => !page.draft)
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

const docSourcesByHref = new Map(Object.entries(docFiles).map(([filePath, source]) => [routeFromFilePath(filePath), source]));

export const docsSearchIndex: SearchEntry[] = docsNav.flatMap(group =>
	group.items.map(item => ({
		...item,
		keywords: `${group.title} ${item.title} ${item.description} ${item.href}`.toLowerCase(),
		content: searchContentFromSource(docSourcesByHref.get(item.href) ?? "")
	}))
);
