import { ApiError } from "@/shared/api/client";

interface RequestErrorMessageOptions {
	prefixFallback?: boolean;
}

export function requestErrorMessage(error: unknown, fallback = "Something went wrong.", options: RequestErrorMessageOptions = {}) {
	let message = "";

	if (error instanceof ApiError) {
		const fieldErrors = error.problem?.errors?.map(item => item.message).filter(Boolean);
		message = fieldErrors?.length ? `${error.message}: ${fieldErrors.join(", ")}` : error.message;
	} else if (error instanceof Error) {
		message = error.message;
	}

	if (!message) {
		return fallback;
	}

	return options.prefixFallback ? `${fallback}: ${message}` : message;
}
