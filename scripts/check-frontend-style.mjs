#!/usr/bin/env node

import { readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const scanRoots = ["web/src", "docs/src", "packages/ui/src", "packages/ui/.storybook"];
const filePattern = /\.(astro|css|js|jsx|mjs|ts|tsx)$/;
const rawColorPattern = /#[0-9a-fA-F]{3,8}\b|rgba?\(/g;
const focusPattern = /outline:\s*(?:none|0)\b|:focus(?!-visible)|:focus-within/g;
const pxPattern = /-?\d+(?:\.\d+)?px\b/g;

const rawColorAllowlist = new Map([
	["docs/src/lib/seo.ts", "browser theme-color metadata requires a concrete color"],
	["docs/src/scripts/docLayout.ts", "detached plain-text popup cannot inherit app CSS variables"],
	["docs/src/scripts/homepageRedesign.js", "canvas and WebGL scene colors require concrete computed strings"],
	["packages/ui/src/stories/Foundations.stories.tsx", "token specimen documents exact token values"],
	["packages/ui/.storybook/preview.ts", "Storybook background swatches document exact token values"],
	["web/src/shared/visualizations/chartTheme.ts", "canvas and SSR chart fallbacks require computed color strings"]
]);

const focusAllowlist = new Map([
	["packages/ui/src/components/Field/Field.module.css", "inner controls suppress native outlines while the frame uses :has(:focus-visible)"],
	["packages/ui/src/components/SearchableSelect/SearchableSelect.module.css", "inner controls suppress native outlines while the frame/search box uses :has(:focus-visible)"],
	["web/src/features/insight/components/InsightControls.module.css", "multi-select input suppresses native outline while the wrapper uses :has(:focus-visible)"],
	["web/src/features/project/components/RoleSelect.module.css", "native select suppresses inner outline while the wrapper uses :has(:focus-visible)"]
]);

const pxFileAllowlist = new Map([
	["docs/src/scripts/docLayout.ts", "detached plain-text popup cannot inherit app sizing tokens"],
	["packages/ui/.storybook/preview.ts", "Storybook viewport presets require pixel dimensions"],
	["web/src/features/status-pages/components/PublicStatusPage.tsx", "IntersectionObserver rootMargin is kept in px for browser compatibility"]
]);

const pxLineAllowPatterns = [
	/\bborder(?:-[a-z]+)?\s*:/,
	/\bborder-width\s*:/,
	/\boutline(?:-[a-z]+)?\s*:/,
	/\bstroke-width\s*:/,
	/\bclip:\s*rect\(/,
	/\bclip-path:\s*inset\(/,
	/\b(width|height):\s*1(?:\.0)?px\b/,
	/\bheight:\s*1\.5px\b/,
	/\bheight:\s*2px\b/,
	/\bbottom:\s*-1px\b/,
	/\bgap:\s*1px\b/,
	/\bleft:\s*-9999px\b/
];

const ignoredPathParts = new Set(["node_modules", "dist", ".astro"]);

function relativePath(filePath) {
	return path.relative(repoRoot, filePath).split(path.sep).join("/");
}

function shouldSkip(filePath) {
	const relative = relativePath(filePath);
	if (relative === "packages/ui/src/styles/tokens.css") return true;
	if (relative.endsWith(".d.ts") || relative.endsWith(".svg")) return true;
	return relative.split("/").some(part => ignoredPathParts.has(part));
}

function walk(dir) {
	const absoluteDir = path.join(repoRoot, dir);
	const files = [];

	for (const entry of readdirSync(absoluteDir)) {
		const absolutePath = path.join(absoluteDir, entry);
		if (shouldSkip(absolutePath)) continue;

		const stats = statSync(absolutePath);
		if (stats.isDirectory()) {
			files.push(...walk(relativePath(absolutePath)));
			continue;
		}

		if (stats.isFile() && filePattern.test(entry)) {
			files.push(absolutePath);
		}
	}

	return files;
}

function lineNumberForIndex(source, index) {
	let line = 1;
	for (let cursor = 0; cursor < index; cursor += 1) {
		if (source.charCodeAt(cursor) === 10) line += 1;
	}
	return line;
}

function collectMatches(filePath, pattern, allowlist) {
	const relative = relativePath(filePath);
	if (allowlist.has(relative)) return [];

	const source = readFileSync(filePath, "utf8");
	const matches = [];

	for (const match of source.matchAll(pattern)) {
		const line = lineNumberForIndex(source, match.index ?? 0);
		const excerpt = source.split(/\r?\n/)[line - 1]?.trim() ?? "";
		matches.push({ relative, line, excerpt });
	}

	return matches;
}

function collectPxIssues(filePath) {
	const relative = relativePath(filePath);
	if (pxFileAllowlist.has(relative)) return [];

	const source = readFileSync(filePath, "utf8");
	const lines = source.split(/\r?\n/);
	const issues = [];

	for (const match of source.matchAll(pxPattern)) {
		const line = lineNumberForIndex(source, match.index ?? 0);
		const excerpt = lines[line - 1]?.trim() ?? "";
		if (pxLineAllowPatterns.some(pattern => pattern.test(excerpt))) continue;
		issues.push({ relative, line, excerpt });
	}

	return issues;
}

const files = scanRoots.flatMap(walk);
const rawColorIssues = files.flatMap(file => collectMatches(file, rawColorPattern, rawColorAllowlist));
const focusIssues = files.flatMap(file => collectMatches(file, focusPattern, focusAllowlist));
const pxIssues = files.flatMap(collectPxIssues);
const issues = [
	...rawColorIssues.map(issue => ({ ...issue, type: "raw color" })),
	...focusIssues.map(issue => ({ ...issue, type: "focus selector" })),
	...pxIssues.map(issue => ({ ...issue, type: "px unit" }))
];

if (issues.length > 0) {
	console.error("Frontend style check failed.");
	for (const issue of issues) {
		console.error(`${issue.relative}:${issue.line}: ${issue.type}: ${issue.excerpt}`);
	}
	console.error("");
	console.error("Use --ns-* tokens for implementation colors and :focus-visible / :has(:focus-visible) for keyboard focus. Add an allowlist entry only for documented non-CSS-variable surfaces.");
	process.exit(1);
}

console.log(`Frontend style check passed (${files.length} files scanned).`);
