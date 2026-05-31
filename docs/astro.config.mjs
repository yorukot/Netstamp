// @ts-check
import mdx from "@astrojs/mdx";
import react from "@astrojs/react";
import { defineConfig } from "astro/config";
import { fileURLToPath } from "node:url";
import remarkDirective from "remark-directive";

const site = process.env.PUBLIC_SITE_URL || "https://netstamp.dev";

const calloutLabels = {
	note: "Note",
	tip: "Tip",
	warning: "Warning",
	caution: "Caution"
};

const calloutIcons = {
	note: "ph-note",
	tip: "ph-lightbulb",
	warning: "ph-warning",
	caution: "ph-warning"
};

function remarkCallouts() {
	return tree => {
		function getText(node) {
			if (!node) return "";
			if (node.type === "text") return node.value ?? "";
			return (node.children ?? []).map(getText).join("");
		}

		function createCalloutTitle(name) {
			return {
				type: "paragraph",
				data: {
					hName: "div",
					hProperties: {
						className: ["calloutTitle"]
					}
				},
				children: [
					{
						type: "mdxJsxTextElement",
						name: calloutIcons[name],
						attributes: [
							{ type: "mdxJsxAttribute", name: "aria-hidden", value: "true" },
							{ type: "mdxJsxAttribute", name: "data-callout-icon", value: "" },
							{ type: "mdxJsxAttribute", name: "size", value: "16" },
							{ type: "mdxJsxAttribute", name: "weight", value: "bold" }
						],
						children: []
					},
					{ type: "text", value: calloutLabels[name] }
				]
			};
		}

		function createCallout(name, children) {
			return {
				type: "blockquote",
				data: {
					hName: "aside",
					hProperties: {
						className: ["callout", `callout-${name}`]
					}
				},
				children: [createCalloutTitle(name), ...children]
			};
		}

		function visit(node) {
			if (node.type === "containerDirective" && Object.hasOwn(calloutLabels, node.name)) {
				Object.assign(node, createCallout(node.name, node.children));
			}

			const children = node.children ?? [];
			for (let index = 0; index < children.length; index++) {
				const child = children[index];
				const text = child.type === "paragraph" ? getText(child).trim() : "";
				const inlineMatch = text.match(/^:::(note|tip|warning|caution)\s+([\s\S]+?)\s*:::$/);

				if (inlineMatch) {
					children[index] = createCallout(inlineMatch[1], [
						{
							type: "paragraph",
							children: [{ type: "text", value: inlineMatch[2] }]
						}
					]);
					continue;
				}

				const startMatch = text.match(/^:::(note|tip|warning|caution)$/);
				if (startMatch) {
					const endIndex = children.findIndex((candidate, candidateIndex) => candidateIndex > index && candidate.type === "paragraph" && getText(candidate).trim() === ":::");

					if (endIndex > index) {
						const body = children.slice(index + 1, endIndex);
						children.splice(index, endIndex - index + 1, createCallout(startMatch[1], body));
						continue;
					}
				}

				visit(child);
			}
		}

		visit(tree);
	};
}

function remarkTerminalCodeBlocks() {
	return tree => {
		function createAttribute(name, value) {
			return { type: "mdxJsxAttribute", name, value };
		}

		function createTerminalBlock(node) {
			const attributes = [createAttribute("title", node.lang || "code")];

			if (node.meta) {
				attributes.push(createAttribute("meta", node.meta));
			}

			return {
				type: "mdxJsxFlowElement",
				name: "Terminal",
				attributes,
				children: [{ type: "text", value: node.value }]
			};
		}

		function visit(node) {
			const children = node.children ?? [];

			for (let index = 0; index < children.length; index++) {
				const child = children[index];

				if (child.type === "code") {
					children[index] = createTerminalBlock(child);
					continue;
				}

				visit(child);
			}
		}

		visit(tree);
	};
}

// https://astro.build/config
export default defineConfig({
	site,
	output: "static",
	integrations: [react(), mdx({ remarkPlugins: [remarkDirective, remarkCallouts, remarkTerminalCodeBlocks] })],
	vite: {
		resolve: {
			alias: {
				"@": fileURLToPath(new URL("./src", import.meta.url))
			}
		}
	}
});
