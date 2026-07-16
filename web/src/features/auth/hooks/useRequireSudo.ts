import { absoluteExternalAuthStartUrl } from "@/shared/api/client";
import { usePasswordSudoMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { usePromptDialog } from "@/shared/components/confirmContext";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { useQueryClient } from "@tanstack/react-query";

export function useRequireSudo(returnTo?: string) {
	const queryClient = useQueryClient();
	const prompt = usePromptDialog();
	const passwordSudoMutation = usePasswordSudoMutation();

	return async function requireSudo(action: () => void) {
		try {
			const status = await queryClient.fetchQuery({ ...authQueries.sudo(), staleTime: 0 });
			if (status.active) {
				action();
				return true;
			}
			if (status.methods.includes("password")) {
				const password = await prompt({
					title: "Confirm it’s you",
					message: "Sensitive account changes require recent authentication for five minutes.",
					inputLabel: "Current password",
					inputType: "password",
					confirmLabel: "Continue"
				});
				if (!password) return false;
				await passwordSudoMutation.mutateAsync({ password });
				action();
				return true;
			}
			const externalProvider = status.methods.find(method => method !== "password");
			if (externalProvider) {
				const url = new URL(absoluteExternalAuthStartUrl(externalProvider));
				url.searchParams.set("intent", "sudo");
				url.searchParams.set("returnTo", returnTo ?? window.location.pathname);
				window.location.assign(url.toString());
				return false;
			}
			pushToast({ title: "Authentication unavailable", message: "This account has no available method for recent authentication.", tone: "critical" });
		} catch (error) {
			pushToast({ title: "Authentication failed", message: requestErrorMessage(error, "Could not confirm your identity."), tone: "critical" });
		}
		return false;
	};
}
