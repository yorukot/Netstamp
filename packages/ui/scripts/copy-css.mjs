import { copyFile, mkdir, readdir } from "node:fs/promises";
import { dirname, extname, join, relative } from "node:path";

const sourceRoot = new URL("../src/", import.meta.url);
const outputRoot = new URL("../dist/", import.meta.url);

async function copyCssFiles(directoryUrl) {
	const entries = await readdir(directoryUrl, { withFileTypes: true });

	await Promise.all(
		entries.map(async entry => {
			if (entry.name === "stories") {
				return;
			}

			const sourceUrl = new URL(`${entry.name}${entry.isDirectory() ? "/" : ""}`, directoryUrl);

			if (entry.isDirectory()) {
				await copyCssFiles(sourceUrl);
				return;
			}

			if (extname(entry.name) !== ".css") {
				return;
			}

			const sourcePath = sourceUrl.pathname;
			const relativePath = relative(sourceRoot.pathname, sourcePath);
			const outputPath = join(outputRoot.pathname, relativePath);
			await mkdir(dirname(outputPath), { recursive: true });
			await copyFile(sourcePath, outputPath);
		})
	);
}

await copyCssFiles(sourceRoot);
