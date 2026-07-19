import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const root = process.cwd();
const issues = [];
const webLocalesRoot = path.join(root, "web/src/i18n/locales");
const docsLocalesRoot = path.join(root, "docs/src/content/docs");
const docsUiLocalesRoot = path.join(root, "docs/src/i18n/locales");
const supportedLocales = new Set(["en", "zh-TW"]);

const report = ({ surface, namespace, key = "(root)", type, source, translation }) => {
	issues.push({ surface, namespace, key, type, source, translation });
};

const directories = async directory => (await readdir(directory, { withFileTypes: true })).filter(entry => entry.isDirectory()).map(entry => entry.name);
const files = async (directory, extension) =>
	(await readdir(directory, { withFileTypes: true }))
		.filter(entry => entry.isFile() && entry.name.endsWith(extension))
		.map(entry => entry.name)
		.sort();

const placeholders = value => [...String(value).matchAll(/{{\s*([A-Za-z0-9_.-]+)\s*}}/g)].map(match => match[1]).sort();
const sameList = (left, right) => left.length === right.length && left.every((value, index) => value === right[index]);

const compareValues = (source, translation, context) => {
	const sourceType = Array.isArray(source) ? "array" : typeof source;
	const translationType = Array.isArray(translation) ? "array" : typeof translation;

	if (sourceType !== translationType) {
		report({ ...context, type: "type mismatch", source, translation });
		return;
	}

	if (sourceType === "object" && source && translation) {
		const sourceKeys = Object.keys(source);
		const translationKeys = Object.keys(translation);
		for (const key of sourceKeys) {
			const keyPath = context.key ? `${context.key}.${key}` : key;
			if (!(key in translation)) {
				report({ ...context, key: keyPath, type: "missing key", source: source[key], translation: undefined });
				continue;
			}
			compareValues(source[key], translation[key], { ...context, key: keyPath });
		}
		for (const key of translationKeys.filter(key => !(key in source))) {
			report({ ...context, key: context.key ? `${context.key}.${key}` : key, type: "extra key", source: undefined, translation: translation[key] });
		}
		return;
	}

	if (sourceType === "array") {
		if (source.length !== translation.length) report({ ...context, type: "array length mismatch", source, translation });
		for (let index = 0; index < Math.min(source.length, translation.length); index += 1) compareValues(source[index], translation[index], { ...context, key: `${context.key}[${index}]` });
		return;
	}

	if (sourceType === "string") {
		if (!translation.trim()) report({ ...context, type: "empty translation", source, translation });
		const sourcePlaceholders = placeholders(source);
		const translationPlaceholders = placeholders(translation);
		if (!sameList(sourcePlaceholders, translationPlaceholders)) report({ ...context, type: "interpolation mismatch", source, translation });
	}
};

const duplicateJsonKeys = source => {
	const duplicates = [];
	let index = 0;
	const whitespace = () => {
		while (/\s/.test(source[index] ?? "")) index += 1;
	};
	const string = () => {
		const start = index;
		index += 1;
		while (index < source.length) {
			if (source[index] === "\\") index += 2;
			else if (source[index++] === '"') break;
		}
		return JSON.parse(source.slice(start, index));
	};
	const value = currentPath => {
		whitespace();
		if (source[index] === "{") {
			index += 1;
			const keys = new Set();
			whitespace();
			while (source[index] !== "}" && index < source.length) {
				const key = string();
				if (keys.has(key)) duplicates.push(currentPath ? `${currentPath}.${key}` : key);
				keys.add(key);
				whitespace();
				index += 1;
				value(currentPath ? `${currentPath}.${key}` : key);
				whitespace();
				if (source[index] === ",") index += 1;
				whitespace();
			}
			index += 1;
			return;
		}
		if (source[index] === "[") {
			index += 1;
			let item = 0;
			whitespace();
			while (source[index] !== "]" && index < source.length) {
				value(`${currentPath}[${item++}]`);
				whitespace();
				if (source[index] === ",") index += 1;
				whitespace();
			}
			index += 1;
			return;
		}
		if (source[index] === '"') {
			string();
			return;
		}
		while (index < source.length && !/[,}\]]/.test(source[index])) index += 1;
	};
	value("");
	return duplicates;
};

const readJson = async (file, context) => {
	const source = await readFile(file, "utf8");
	try {
		for (const key of duplicateJsonKeys(source)) report({ ...context, key, type: "duplicate key", source: undefined, translation: undefined });
		return JSON.parse(source);
	} catch (error) {
		report({ ...context, type: "invalid JSON", source: error instanceof Error ? error.message : String(error), translation: undefined });
		return undefined;
	}
};

for (const locale of await directories(webLocalesRoot)) {
	if (!supportedLocales.has(locale)) report({ surface: "web", namespace: locale, type: "invalid locale directory", source: "en or zh-TW", translation: locale });
}

const sourceNamespaces = await files(path.join(webLocalesRoot, "en"), ".json");
const targetNamespaces = await files(path.join(webLocalesRoot, "zh-TW"), ".json");
for (const namespaceFile of sourceNamespaces) {
	const namespace = namespaceFile.replace(/\.json$/, "");
	if (!targetNamespaces.includes(namespaceFile)) {
		report({ surface: "web", namespace, type: "missing namespace", source: namespaceFile, translation: undefined });
		continue;
	}
	const source = await readJson(path.join(webLocalesRoot, "en", namespaceFile), { surface: "web", namespace });
	const translation = await readJson(path.join(webLocalesRoot, "zh-TW", namespaceFile), { surface: "web", namespace });
	if (source !== undefined && translation !== undefined) compareValues(source, translation, { surface: "web", namespace, key: "" });
}
for (const namespaceFile of targetNamespaces.filter(file => !sourceNamespaces.includes(file))) {
	report({ surface: "web", namespace: namespaceFile.replace(/\.json$/, ""), type: "extra namespace", source: undefined, translation: namespaceFile });
}

for (const locale of await directories(docsUiLocalesRoot)) {
	if (!supportedLocales.has(locale)) report({ surface: "docs-ui", namespace: locale, type: "invalid locale directory", source: "en or zh-TW", translation: locale });
}

const sourceDocsUiNamespaces = await files(path.join(docsUiLocalesRoot, "en"), ".json");
const targetDocsUiNamespaces = await files(path.join(docsUiLocalesRoot, "zh-TW"), ".json");
for (const namespaceFile of sourceDocsUiNamespaces) {
	const namespace = namespaceFile.replace(/\.json$/, "");
	if (!targetDocsUiNamespaces.includes(namespaceFile)) {
		report({ surface: "docs-ui", namespace, type: "missing namespace", source: namespaceFile, translation: undefined });
		continue;
	}
	const source = await readJson(path.join(docsUiLocalesRoot, "en", namespaceFile), { surface: "docs-ui", namespace });
	const translation = await readJson(path.join(docsUiLocalesRoot, "zh-TW", namespaceFile), { surface: "docs-ui", namespace });
	if (source !== undefined && translation !== undefined) compareValues(source, translation, { surface: "docs-ui", namespace, key: "" });
}
for (const namespaceFile of targetDocsUiNamespaces.filter(file => !sourceDocsUiNamespaces.includes(file))) {
	report({ surface: "docs-ui", namespace: namespaceFile.replace(/\.json$/, ""), type: "extra namespace", source: undefined, translation: namespaceFile });
}

for (const locale of await directories(docsLocalesRoot)) {
	if (!supportedLocales.has(locale)) report({ surface: "docs", namespace: locale, type: "invalid locale directory", source: "en or zh-TW", translation: locale });
}

const recursiveFiles = async directory => {
	const output = [];
	for (const entry of await readdir(directory, { withFileTypes: true })) {
		const entryPath = path.join(directory, entry.name);
		if (entry.isDirectory()) output.push(...(await recursiveFiles(entryPath)));
		else if (entry.isFile() && entry.name.endsWith(".mdx")) output.push(entryPath);
	}
	return output;
};

const englishDocsRoot = path.join(docsLocalesRoot, "en");
const targetDocsRoot = path.join(docsLocalesRoot, "zh-TW");
for (const sourceFile of await recursiveFiles(englishDocsRoot)) {
	const relative = path.relative(englishDocsRoot, sourceFile);
	const targetFile = path.join(targetDocsRoot, relative);
	let source;
	let translation;
	try {
		[source, translation] = await Promise.all([readFile(sourceFile, "utf8"), readFile(targetFile, "utf8")]);
	} catch {
		report({ surface: "docs", namespace: relative, type: "missing translated document", source: relative, translation: undefined });
		continue;
	}
	const fencedBlocks = text => [...text.matchAll(/```[^\n]*\n([\s\S]*?)```/g)].map(match => match[1].trim());
	const sourceBlocks = fencedBlocks(source);
	const targetBlocks = fencedBlocks(translation);
	if (!sameList(sourceBlocks, targetBlocks)) report({ surface: "docs", namespace: relative, type: "code block mismatch", source: sourceBlocks, translation: targetBlocks });
	if (!translation.trim()) report({ surface: "docs", namespace: relative, type: "empty translation", source: source, translation });
}

if (issues.length) {
	for (const issue of issues) {
		console.error(`\n[${issue.surface}/${issue.namespace}] ${issue.key ?? "(root)"}: ${issue.type}`);
		console.error(`  source: ${JSON.stringify(issue.source)}`);
		console.error(`  translation: ${JSON.stringify(issue.translation)}`);
	}
	console.error(`\nFound ${issues.length} localization issue(s).`);
	process.exit(1);
}

console.log(
	`Localization check passed: ${sourceNamespaces.length} web namespaces, ${sourceDocsUiNamespaces.length} docs UI namespaces, and ${(await recursiveFiles(englishDocsRoot)).length} documentation pages.`
);
