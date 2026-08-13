// @ts-check
import { unified } from "@astrojs/markdown-remark";
import mdx from "@astrojs/mdx";
import react from "@astrojs/react";
import { supportedLocales } from "@netstamp/i18n";
import { defineConfig } from "astro/config";
import remarkDirective from "remark-directive";

const site = process.env.PUBLIC_SITE_URL || "https://netstamp.dev";

const calloutLabels = {
	en: { note: "Note", tip: "Tip", warning: "Warning", caution: "Caution" },
	"zh-TW": { note: "注意", tip: "提示", warning: "警告", caution: "小心" }
};

const calloutIcons = {
	note: "ph-note",
	tip: "ph-lightbulb",
	warning: "ph-warning",
	caution: "ph-warning"
};

const codeBlockLabels = {
	en: { copy: "Copy", copied: "Copied" },
	"zh-TW": { copy: "複製", copied: "已複製" }
};

const netstampCodeTheme = {
	name: "netstamp",
	type: "dark",
	colors: {
		"editor.background": "transparent",
		"editor.foreground": "var(--ns-text-muted)"
	},
	tokenColors: [
		{
			scope: ["comment", "punctuation.definition.comment"],
			settings: { foreground: "var(--ns-text-low)", fontStyle: "italic" }
		},
		{
			scope: ["keyword", "keyword.operator", "storage", "storage.type"],
			settings: { foreground: "var(--ns-primary-hover)" }
		},
		{
			scope: ["string", "constant", "constant.numeric", "constant.language"],
			settings: { foreground: "var(--ns-secondary-hover)" }
		},
		{
			scope: ["entity.name.function", "entity.name.type", "entity.name.tag", "support.function", "support.type"],
			settings: { foreground: "var(--ns-secondary)" }
		},
		{
			scope: ["entity.other.attribute-name", "variable.parameter", "variable.other.property"],
			settings: { foreground: "var(--ns-text)" }
		},
		{
			scope: ["punctuation", "meta.brace", "meta.delimiter"],
			settings: { foreground: "var(--ns-text-subtle)" }
		},
		{
			scope: ["invalid", "invalid.illegal"],
			settings: { foreground: "var(--ns-critical)" }
		}
	]
};

const localeForFile = file => (String(file.path || "").includes("/zh-TW/") ? "zh-TW" : "en");

function remarkCallouts() {
	return (tree, file) => {
		const locale = String(file.path || "").includes("/zh-TW/") ? "zh-TW" : "en";
		const localizedCalloutLabels = calloutLabels[locale];
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
					{ type: "text", value: localizedCalloutLabels[name] }
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
			if (node.type === "containerDirective" && Object.hasOwn(localizedCalloutLabels, node.name)) {
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

const rehypeTableScrollers = () => {
	return (tree, file) => {
		const locale = localeForFile(file);
		const label = locale === "zh-TW" ? "可水平捲動的表格" : "Scrollable table";

		const visit = node => {
			const children = node.children ?? [];

			for (let index = 0; index < children.length; index++) {
				const child = children[index];

				if (child.type === "element" && child.tagName === "table") {
					children[index] = {
						type: "element",
						tagName: "div",
						properties: {
							ariaLabel: label,
							className: ["docTableScroller"],
							role: "region",
							tabIndex: 0
						},
						children: [child]
					};
					continue;
				}

				visit(child);
			}
		};

		visit(tree);
	};
};

const rehypeCodeBlocks = () => {
	return (tree, file) => {
		const labels = codeBlockLabels[localeForFile(file)];
		const createElement = (tagName, properties = {}, children = []) => ({ type: "element", tagName, properties, children });
		const createText = value => ({ type: "text", value });

		const languageForPre = node => {
			if (typeof node.properties?.dataLanguage === "string") {
				return node.properties.dataLanguage;
			}

			const code = node.children?.find(child => child.type === "element" && child.tagName === "code");
			const classNames = Array.isArray(code?.properties?.className) ? code.properties.className : [code?.properties?.className];
			const languageClass = classNames.find(className => typeof className === "string" && className.startsWith("language-"));

			return typeof languageClass === "string" ? languageClass.slice("language-".length) : "text";
		};

		const createHeader = language =>
			createElement("div", { className: ["docCodeHeader"] }, [
				createElement(
					"span",
					{ ariaHidden: "true", className: ["docCodeWindowControls"] },
					Array.from({ length: 3 }, () => createElement("span", { className: ["docCodeWindowControl"] }))
				),
				createElement("strong", { className: ["docCodeLanguage"] }, [createText(language)]),
				createElement(
					"button",
					{
						ariaLabel: labels.copy,
						className: ["docCodeCopyButton"],
						dataCopiedLabel: labels.copied,
						dataCopyLabel: labels.copy,
						dataNsCodeCopy: "",
						title: labels.copy,
						type: "button"
					},
					[createElement("span", { dataNsCodeCopyLabel: "" }, [createText(labels.copy)])]
				)
			]);

		const visit = node => {
			const children = node.children ?? [];

			for (let index = 0; index < children.length; index++) {
				const child = children[index];

				if (child.type === "element" && child.tagName === "pre" && child.children?.some(grandchild => grandchild.type === "element" && grandchild.tagName === "code")) {
					child.properties ??= {};
					const currentClasses = child.properties.className ?? child.properties.class;
					const classNames = Array.isArray(currentClasses) ? currentClasses : typeof currentClasses === "string" ? currentClasses.split(/\s+/) : [];
					delete child.properties.class;
					child.properties.className = [...classNames, "docCodePre"];
					child.properties.tabIndex = 0;
					children[index] = createElement("div", { className: ["docCodeBlock", "ns-frame"], dataNsCodeBlock: "" }, [createHeader(languageForPre(child)), child]);
					continue;
				}

				visit(child);
			}
		};

		visit(tree);
	};
};

// https://astro.build/config
export default defineConfig({
	site,
	output: "static",
	redirects: {
		"/docs": { status: 301, destination: "/docs/getting-started/quick-start/" },
		"/zh-TW/docs": { status: 301, destination: "/zh-TW/docs/getting-started/quick-start/" }
	},
	i18n: {
		defaultLocale: "en",
		locales: [...supportedLocales],
		routing: {
			prefixDefaultLocale: false,
			redirectToDefaultLocale: false
		}
	},
	markdown: {
		shikiConfig: {
			theme: netstampCodeTheme
		},
		processor: unified({
			remarkPlugins: [remarkDirective, remarkCallouts],
			rehypePlugins: [rehypeTableScrollers, rehypeCodeBlocks]
		})
	},
	integrations: [react(), mdx()]
});
