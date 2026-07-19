export interface HeadEntry {
	tag: string;
	content?: string;
	attrs?: Record<string, string>;
}

interface PageSeoOptions {
	title: string;
	description: string;
	canonicalUrl: string;
	imageUrl: string;
	ogType?: "article" | "website";
	noindex?: boolean;
	locale?: SupportedLocale;
}

interface SchemaGraphOptions {
	title: string;
	description: string;
	pageUrl: string;
	siteUrl: string;
	logoUrl: string;
	imageUrl: string;
	schemas?: Record<string, unknown>[];
}

interface BreadcrumbItem {
	name: string;
	url: string;
}

const fallbackSiteUrl = "https://netstamp.dev";

export function resolveSiteUrl(site?: URL | string | null) {
	const rawSite = site?.toString() || import.meta.env.PUBLIC_SITE_URL || fallbackSiteUrl;
	const parsedSite = new URL(rawSite);
	return parsedSite.toString().replace(/\/$/, "");
}

export function absoluteUrl(pathOrUrl: string, siteUrl = fallbackSiteUrl) {
	if (/^https?:\/\//i.test(pathOrUrl)) return pathOrUrl;

	const baseUrl = siteUrl.endsWith("/") ? siteUrl : `${siteUrl}/`;
	const relativePath = pathOrUrl.startsWith("/") ? pathOrUrl.slice(1) : pathOrUrl;
	return new URL(relativePath, baseUrl).toString();
}

export function createPageSeoHead({ title, description, canonicalUrl, imageUrl, ogType = "website", noindex = false, locale = "en" }: PageSeoOptions): HeadEntry[] {
	const robots = noindex ? "noindex,nofollow,noarchive" : "index,follow,max-image-preview:large";

	return [
		{ tag: "link", attrs: { rel: "canonical", href: canonicalUrl } },
		{ tag: "meta", attrs: { name: "description", content: description } },
		{ tag: "meta", attrs: { name: "robots", content: robots } },
		{ tag: "meta", attrs: { name: "theme-color", content: "#000000" } },
		{ tag: "meta", attrs: { property: "og:site_name", content: "Netstamp" } },
		{ tag: "meta", attrs: { property: "og:type", content: ogType } },
		{ tag: "meta", attrs: { property: "og:title", content: title } },
		{ tag: "meta", attrs: { property: "og:description", content: description } },
		{ tag: "meta", attrs: { property: "og:url", content: canonicalUrl } },
		{ tag: "meta", attrs: { property: "og:image", content: imageUrl } },
		{ tag: "meta", attrs: { property: "og:image:secure_url", content: imageUrl } },
		{ tag: "meta", attrs: { property: "og:image:type", content: "image/png" } },
		{ tag: "meta", attrs: { property: "og:image:width", content: "1200" } },
		{ tag: "meta", attrs: { property: "og:image:height", content: "600" } },
		{ tag: "meta", attrs: { property: "og:image:alt", content: locale === "zh-TW" ? "Netstamp 網路可觀測性地圖預覽" : "Netstamp network observability map preview" } },
		{ tag: "meta", attrs: { property: "og:locale", content: locale === "zh-TW" ? "zh_TW" : "en_US" } },
		{ tag: "meta", attrs: { property: "og:locale:alternate", content: locale === "zh-TW" ? "en_US" : "zh_TW" } },
		{ tag: "meta", attrs: { name: "twitter:card", content: "summary_large_image" } },
		{ tag: "meta", attrs: { name: "twitter:title", content: title } },
		{ tag: "meta", attrs: { name: "twitter:description", content: description } },
		{ tag: "meta", attrs: { name: "twitter:image", content: imageUrl } },
		{ tag: "meta", attrs: { name: "twitter:image:alt", content: locale === "zh-TW" ? "Netstamp 網路可觀測性地圖預覽" : "Netstamp network observability map preview" } }
	];
}

export function createBreadcrumbSchema(items: BreadcrumbItem[]) {
	return {
		"@type": "BreadcrumbList",
		itemListElement: items.map((item, index) => ({
			"@type": "ListItem",
			position: index + 1,
			name: item.name,
			item: item.url
		}))
	};
}

export function createSoftwareApplicationSchema(siteUrl: string, imageUrl: string, description: string) {
	return {
		"@type": "SoftwareApplication",
		"@id": `${siteUrl}/#software`,
		name: "Netstamp",
		applicationCategory: "DeveloperApplication",
		operatingSystem: "Web, Linux",
		description,
		url: siteUrl,
		image: imageUrl,
		isAccessibleForFree: true,
		license: "https://www.apache.org/licenses/LICENSE-2.0",
		offers: {
			"@type": "Offer",
			price: "0",
			priceCurrency: "USD"
		},
		publisher: {
			"@id": `${siteUrl}/#organization`
		}
	};
}

export function createTechArticleSchema(siteUrl: string, pageUrl: string, title: string, description: string) {
	return {
		"@type": "TechArticle",
		"@id": `${pageUrl}#article`,
		headline: title,
		description,
		url: pageUrl,
		mainEntityOfPage: {
			"@id": `${pageUrl}#webpage`
		},
		isPartOf: {
			"@id": `${siteUrl}/docs/#docs`
		},
		publisher: {
			"@id": `${siteUrl}/#organization`
		}
	};
}

export function createApiReferenceSchema(siteUrl: string, pageUrl: string, title: string, description: string) {
	return {
		"@type": "TechArticle",
		"@id": `${pageUrl}#api-reference`,
		headline: title,
		description,
		url: pageUrl,
		about: {
			"@type": "WebAPI",
			name: "Netstamp Controller API",
			documentation: pageUrl,
			provider: {
				"@id": `${siteUrl}/#organization`
			}
		},
		mainEntityOfPage: {
			"@id": `${pageUrl}#webpage`
		},
		publisher: {
			"@id": `${siteUrl}/#organization`
		}
	};
}

export function createSchemaGraph({ title, description, pageUrl, siteUrl, logoUrl, imageUrl, schemas = [] }: SchemaGraphOptions) {
	return {
		"@context": "https://schema.org",
		"@graph": [
			{
				"@type": "Organization",
				"@id": `${siteUrl}/#organization`,
				name: "Netstamp",
				url: siteUrl,
				logo: {
					"@type": "ImageObject",
					url: logoUrl
				},
				image: imageUrl,
				sameAs: ["https://github.com/yorukot/netstamp"]
			},
			{
				"@type": "WebSite",
				"@id": `${siteUrl}/#website`,
				name: "Netstamp",
				url: siteUrl,
				description,
				publisher: {
					"@id": `${siteUrl}/#organization`
				}
			},
			{
				"@type": "WebPage",
				"@id": `${pageUrl}#webpage`,
				name: title,
				description,
				url: pageUrl,
				isPartOf: {
					"@id": `${siteUrl}/#website`
				},
				primaryImageOfPage: {
					"@type": "ImageObject",
					url: imageUrl,
					width: 1200,
					height: 600
				},
				publisher: {
					"@id": `${siteUrl}/#organization`
				}
			},
			...schemas
		]
	};
}
import type { SupportedLocale } from "@netstamp/i18n";
