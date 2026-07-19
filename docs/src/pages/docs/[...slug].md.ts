import { getDocEntry, getDocsPages } from "@/data/docs";
import { markdownResponse } from "@/lib/docsMarkdown";
import type { APIRoute } from "astro";

export const getStaticPaths = () =>
	getDocsPages("en").map(page => ({
		params: { slug: page.contentId },
		props: { contentId: page.contentId }
	}));

export const GET: APIRoute = ({ props }) => {
	const { entry } = getDocEntry(props.contentId, "en");

	if (!entry) {
		return new Response("Documentation source not found.\n", { status: 404 });
	}

	return markdownResponse(entry);
};
