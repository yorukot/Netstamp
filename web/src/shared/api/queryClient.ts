import { pushErrorToast } from "@/shared/toast/toastStore";
import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { ApiError } from "./client";

function messageForError(error: unknown) {
	if (error instanceof ApiError) {
		return error.message;
	}

	if (error instanceof Error) {
		return error.message;
	}

	return "Something went wrong.";
}

export const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onError: error => pushErrorToast(messageForError(error))
	}),
	queryCache: new QueryCache({
		onError: (error, query) => {
			if (error instanceof ApiError && error.status === 401 && query.queryKey[0] === "auth") {
				return;
			}

			pushErrorToast(messageForError(error));
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
