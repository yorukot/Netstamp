import { pushErrorToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { ApiError } from "./client";

interface MutationErrorMeta {
	suppressGlobalErrorToast?: boolean;
}

function suppressesGlobalErrorToast(meta: unknown) {
	return Boolean(meta && typeof meta === "object" && (meta as MutationErrorMeta).suppressGlobalErrorToast === true);
}

export const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onError: (error, _variables, _context, mutation) => {
			if (suppressesGlobalErrorToast(mutation.options.meta)) {
				return;
			}

			pushErrorToast(requestErrorMessage(error));
		}
	}),
	queryCache: new QueryCache({
		onError: (error, query) => {
			if (error instanceof ApiError && error.status === 401 && query.queryKey[0] === "auth") {
				return;
			}

			pushErrorToast(requestErrorMessage(error));
		}
	}),
	defaultOptions: {
		queries: {
			gcTime: 10 * 60 * 1000,
			refetchOnWindowFocus: false,
			retry: 1,
			staleTime: 60 * 1000
		},
		mutations: {
			retry: 0
		}
	}
});
