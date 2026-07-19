import type { CollectionEntry } from "astro:content";

type DocsEntry = CollectionEntry<"docs">;

export const markdownForDocEntry = (entry: DocsEntry) => {
	const body = entry.body.trim();

	return [`# ${entry.data.title}`, entry.data.description, body].filter(Boolean).join("\n\n") + "\n";
};

export const markdownResponse = (entry: DocsEntry) =>
	new Response(markdownForDocEntry(entry), {
		headers: {
			"Cache-Control": "public, max-age=300",
			"Content-Type": "text/markdown; charset=utf-8",
			"X-Content-Type-Options": "nosniff"
		}
	});
