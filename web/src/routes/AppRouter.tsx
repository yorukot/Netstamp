import { SessionProvider } from "@/features/auth/session/SessionProvider";
import { queryClient } from "@/shared/api/queryClient";
import { CurrentProjectProvider } from "@/shared/api/useCurrentProject";
import { ConfirmProvider } from "@/shared/components/ConfirmProvider";
import { ToastProvider } from "@/shared/components/ToastProvider";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router-dom";
import { router } from "./router";

export function AppRouter() {
	return (
		<QueryClientProvider client={queryClient}>
			<SessionProvider>
				<CurrentProjectProvider>
					<ConfirmProvider>
						<RouterProvider router={router} />
						<ToastProvider />
					</ConfirmProvider>
				</CurrentProjectProvider>
			</SessionProvider>
		</QueryClientProvider>
	);
}
