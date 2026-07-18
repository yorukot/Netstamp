import { pushErrorToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { isSessionUnavailableError } from "./client";

interface ErrorMeta {
	suppressGlobalErrorToast?: boolean;
}

function suppressesGlobalErrorToast(meta: unknown) {
	return Boolean(meta && typeof meta === "object" && (meta as ErrorMeta).suppressGlobalErrorToast === true);
}

export const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onError: (error, _variables, _context, mutation) => {
			if (isSessionUnavailableError(error) || suppressesGlobalErrorToast(mutation.options.meta)) {
				return;
			}

			pushErrorToast(requestErrorMessage(error));
		}
	}),
	queryCache: new QueryCache({
		onError: (error, query) => {
			if (isSessionUnavailableError(error) || suppressesGlobalErrorToast(query.meta)) {
				return;
			}

			pushErrorToast(requestErrorMessage(error));
		}
	}),
	defaultOptions: {
		queries: {
			gcTime: 10 * 60 * 1000,
			refetchOnWindowFocus: false,
			retry: (failureCount, error) => !isSessionUnavailableError(error) && failureCount < 1,
			staleTime: 60 * 1000
		},
		mutations: {
			retry: 0
		}
	}
});
