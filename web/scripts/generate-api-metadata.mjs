import { readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

const openAPIPath = fileURLToPath(new URL("../../docs/public/openapi.json", import.meta.url));
const outputPath = fileURLToPath(new URL("../src/shared/api/metadata.generated.ts", import.meta.url));

export const apiBaseUrlFromOpenAPI = spec => {
	const serverUrl = spec?.servers?.[0]?.url;
	if (typeof serverUrl !== "string" || !/^\/api\/v[1-9]\d*$/.test(serverUrl)) {
		throw new Error(`OpenAPI servers[0].url must be a relative versioned API path, got ${JSON.stringify(serverUrl)}`);
	}
	return serverUrl;
};

export const generatedMetadataSource = apiBaseUrl =>
	`// This file was auto-generated from docs/public/openapi.json. Do not edit directly.\n\nexport const defaultApiBaseUrl = ${JSON.stringify(apiBaseUrl)} as const;\n`;

const generate = async () => {
	const spec = JSON.parse(await readFile(openAPIPath, "utf8"));
	await writeFile(outputPath, generatedMetadataSource(apiBaseUrlFromOpenAPI(spec)));
};

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
	await generate();
}
