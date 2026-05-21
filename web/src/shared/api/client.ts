import createClient from "openapi-fetch";
import type { components, paths } from "./openapi";

export type ApiProblem = components["schemas"]["ProblemDetails"];

export class ApiError extends Error {
	readonly problem?: ApiProblem;
	readonly status: number;

	constructor(message: string, status: number, problem?: ApiProblem) {
		super(message);
		this.name = "ApiError";
		this.status = status;
		this.problem = problem;
	}
}

export const apiBaseUrl = import.meta.env.VITE_NETSTAMP_API_BASE_URL || "/api/v1";

export const apiClient = createClient<paths>({
	baseUrl: apiBaseUrl,
	credentials: "include"
});

export function apiUrl<TPath extends keyof paths & string>(path: TPath) {
	return `${apiBaseUrl.replace(/\/$/, "")}${path}`;
}

export function absoluteApiUrl<TPath extends keyof paths & string>(path: TPath) {
	return new URL(apiUrl(path), window.location.origin).toString();
}

interface ApiResult<TData> {
	data?: TData;
	error?: ApiProblem;
	response: Response;
}

export async function readApiData<TData>(request: Promise<ApiResult<TData>>): Promise<TData> {
	const { data, error, response } = await request;

	if (error) {
		throw new ApiError(error.detail || error.title || "API request failed", response.status, error);
	}

	if (data === undefined) {
		throw new ApiError("API response did not include a body", response.status);
	}

	return data;
}

export async function readEmptyApiResponse(request: Promise<ApiResult<unknown>>): Promise<void> {
	const { error, response } = await request;

	if (error) {
		throw new ApiError(error.detail || error.title || "API request failed", response.status, error);
	}
}
