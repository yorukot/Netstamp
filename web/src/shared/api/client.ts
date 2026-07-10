import type { Middleware } from "openapi-fetch";
import createClient from "openapi-fetch";
import type { components, paths } from "./openapi";

export type ApiProblem = components["schemas"]["ProblemDetails"];
export type ApiErrorCode = NonNullable<ApiProblem["code"]>;

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

let csrfToken: string | null = null;

const csrfMiddleware: Middleware = {
	async onRequest({ request, schemaPath }) {
		if (csrfSafeRequest(request.method) || csrfSkippedPath(schemaPath)) {
			return undefined;
		}

		const headers = new Headers(request.headers);
		headers.set("X-CSRF-Token", await getCSRFToken());
		return new Request(request, { headers });
	},
	onResponse({ response }) {
		if (response.status === 401) {
			clearCSRFToken();
		}
		return undefined;
	}
};

apiClient.use(csrfMiddleware);

export function apiUrl<TPath extends keyof paths & string>(path: TPath) {
	return `${apiBaseUrl.replace(/\/$/, "")}${path}`;
}

export function absoluteApiUrl<TPath extends keyof paths & string>(path: TPath) {
	return new URL(apiUrl(path), window.location.origin).toString();
}

export function clearCSRFToken() {
	csrfToken = null;
}

async function getCSRFToken() {
	if (csrfToken) {
		return csrfToken;
	}

	const response = await fetch(apiUrl("/auth/csrf"), {
		credentials: "include",
		headers: {
			Accept: "application/json"
		}
	});
	if (!response.ok) {
		throw new ApiError(response.statusText || "CSRF token request failed", response.status);
	}
	const data = (await response.json()) as components["schemas"]["CSRFTokenResponse"];
	csrfToken = data.csrfToken;
	return csrfToken;
}

function csrfSafeRequest(method: string) {
	return ["GET", "HEAD", "OPTIONS"].includes(method.toUpperCase());
}

function csrfSkippedPath(path: string) {
	return ["/auth/login", "/auth/register", "/auth/password-resets", "/auth/email-verifications"].includes(path);
}

export function apiProblemCode(error: unknown): ApiErrorCode | undefined {
	if (error instanceof ApiError) {
		return error.problem?.code;
	}

	return apiProblemFromError(error)?.code;
}

export function hasApiProblemCode(error: unknown, ...codes: ApiErrorCode[]) {
	const code = apiProblemCode(error);
	return code !== undefined && codes.includes(code);
}

interface ApiResult<TData> {
	data?: TData;
	error?: unknown;
	response: Response;
}

function apiProblemFromError(error: unknown) {
	if (!error || typeof error !== "object" || Array.isArray(error)) {
		return undefined;
	}

	const candidate = error as Partial<ApiProblem>;
	if (typeof candidate.detail === "string" || typeof candidate.title === "string" || typeof candidate.status === "number" || Array.isArray(candidate.errors)) {
		return candidate as ApiProblem;
	}

	return undefined;
}

function apiErrorMessage(error: unknown, response: Response) {
	const problem = apiProblemFromError(error);

	if (problem?.detail || problem?.title) {
		return problem.detail || problem.title || "API request failed";
	}
	if (typeof error === "string" && error.trim()) {
		return error;
	}

	return response.statusText || "API request failed";
}

function throwApiError(error: unknown, response: Response): never {
	throw new ApiError(apiErrorMessage(error, response), response.status, apiProblemFromError(error));
}

export async function readApiData<TData>(request: Promise<ApiResult<TData>>): Promise<TData> {
	const { data, error, response } = await request;

	if (!response.ok) {
		throwApiError(error, response);
	}

	if (data === undefined) {
		throw new ApiError("API response did not include a body", response.status);
	}

	return data;
}

export async function readEmptyApiResponse(request: Promise<ApiResult<unknown>>): Promise<void> {
	const { error, response } = await request;

	if (!response.ok) {
		throwApiError(error, response);
	}
}
