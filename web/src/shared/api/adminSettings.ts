import { ApiError, readApiData, readEmptyApiResponse } from "./client";
import type { components } from "./openapi";

export type AdminSettingsProvider = "oidc" | "google" | "github";
export type AdminAccessSettings = components["schemas"]["AdminAccessSettings"];
export type AdminAccessSettingsPatch = components["schemas"]["AdminAccessSettingsPatch"];
export type AdminSMTPTLSMode = "starttls" | "implicit" | "none";
export type AdminSMTPSettings = components["schemas"]["AdminSMTPSettings"];
export type AdminSMTPSettingsPatch = components["schemas"]["AdminSMTPSettingsPatch"];
export type AdminOIDCProviderSettings = Omit<components["schemas"]["AdminOIDCProviderSettings"], "callbackUrl"> & { callbackUrl?: string };
export type AdminOIDCProviderSettingsPatch = components["schemas"]["AdminOIDCProviderSettingsPatch"];
export type AdminGoogleProviderSettings = Omit<components["schemas"]["AdminGoogleProviderSettings"], "callbackUrl"> & { callbackUrl?: string };
export type AdminGoogleProviderSettingsPatch = components["schemas"]["AdminGoogleProviderSettingsPatch"];
export type AdminGitHubProviderSettings = Omit<components["schemas"]["AdminGitHubProviderSettings"], "callbackUrl"> & { callbackUrl?: string };
export type AdminGitHubProviderSettingsPatch = components["schemas"]["AdminGitHubProviderSettingsPatch"];

export interface VersionedAdminSettings<TSettings> {
	settings: TSettings;
	etag: string;
}

interface ApiTransportResult<TData> {
	data?: TData;
	error?: unknown;
	response: Response;
}

const unwrapSettings = <TSettings>(data: unknown): TSettings => {
	if (data && typeof data === "object" && "settings" in data) {
		return (data as { settings: TSettings }).settings;
	}

	throw new ApiError("Settings response did not include settings", 500);
};

export const readVersionedAdminSettings = async <TSettings>(request: Promise<ApiTransportResult<unknown>>): Promise<VersionedAdminSettings<TSettings>> => {
	const result = await request;
	const data = await readApiData(Promise.resolve(result));
	const etag = result.response.headers.get("ETag");

	if (!etag) {
		throw new ApiError("Settings response did not include an ETag", result.response.status);
	}

	return {
		settings: unwrapSettings<TSettings>(data),
		etag
	};
};

export const readAdminSettingsValidation = (request: Promise<ApiTransportResult<unknown>>) => readEmptyApiResponse(request);
