import { currentLocale, i18n } from "@/i18n";
import { ApiError } from "@/shared/api/client";

interface RequestErrorMessageOptions {
	prefixFallback?: boolean;
}

const localizedApiErrorMessage = (error: ApiError) => {
	if (error.status === 401 || error.status === 403) return i18n.t("unauthorized", { ns: "errors" });
	if (error.status === 404) return i18n.t("notFound", { ns: "errors" });
	if (error.status === 400 || error.status === 422) return i18n.t("validation", { ns: "errors" });
	if (error.status >= 500) return i18n.t("server", { ns: "errors" });

	return undefined;
};

export const requestErrorMessage = (error: unknown, fallback?: string, options: RequestErrorMessageOptions = {}) => {
	const fallbackMessage = fallback || i18n.t("generic", { ns: "errors" });
	let message = "";

	if (error instanceof ApiError) {
		if (currentLocale() !== "en") {
			message = localizedApiErrorMessage(error) || fallbackMessage;
		} else {
			const fieldErrors = error.problem?.errors?.map(item => item.message).filter(Boolean);
			message = fieldErrors?.length ? `${error.message}: ${fieldErrors.join(", ")}` : error.message;
		}
	} else if (error instanceof Error) {
		const networkFailure = error instanceof TypeError || /failed to fetch|network error/i.test(error.message);
		message = networkFailure ? i18n.t("network", { ns: "errors" }) : currentLocale() === "en" ? error.message : fallbackMessage;
	}

	if (!message) {
		return fallbackMessage;
	}

	return options.prefixFallback ? `${fallbackMessage}: ${message}` : message;
};
