import { getDocEntry, getDocsPages } from "@/data/docs";
import { markdownResponse } from "@/lib/docsMarkdown";
import type { APIRoute } from "astro";

export const getStaticPaths = () =>
	getDocsPages("zh-TW").map(page => ({
		params: { slug: page.contentId },
		props: { contentId: page.contentId }
	}));

export const GET: APIRoute = ({ props }) => {
	const { entry } = getDocEntry(props.contentId, "zh-TW");

	if (!entry) {
		return new Response("Documentation source not found.\n", { status: 404 });
	}

	return markdownResponse(entry);
};
