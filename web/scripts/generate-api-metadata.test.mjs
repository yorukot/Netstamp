import { describe, expect, it } from "vitest";
import { apiBaseUrlFromOpenAPI, generatedMetadataSource } from "./generate-api-metadata.mjs";

describe("generate API metadata", () => {
	it("reads the relative versioned base URL from the OpenAPI server", () => {
		expect(apiBaseUrlFromOpenAPI({ servers: [{ url: "/api/v2" }] })).toBe("/api/v2");
	});

	it.each([{}, { servers: [] }, { servers: [{ url: "https://netstamp.example.com/api/v1" }] }, { servers: [{ url: "/api/latest" }] }])("rejects an unusable OpenAPI server URL %#", spec => {
		expect(() => apiBaseUrlFromOpenAPI(spec)).toThrow("OpenAPI servers[0].url");
	});

	it("writes deterministic TypeScript source", () => {
		expect(generatedMetadataSource("/api/v1")).toBe('// This file was auto-generated from docs/public/openapi.json. Do not edit directly.\n\nexport const defaultApiBaseUrl = "/api/v1" as const;\n');
	});
});
