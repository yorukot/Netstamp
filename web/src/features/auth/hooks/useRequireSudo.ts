import { absoluteExternalAuthStartUrl } from "@/shared/api/client";
import { usePasswordSudoMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { useChoiceDialog, usePromptDialog } from "@/shared/components/confirmContext";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { useQueryClient } from "@tanstack/react-query";

export function useRequireSudo(returnTo?: string) {
	const queryClient = useQueryClient();
	const choose = useChoiceDialog();
	const prompt = usePromptDialog();
	const passwordSudoMutation = usePasswordSudoMutation();

	return async function requireSudo(action: () => void, options: { returnTo?: string } = {}) {
		try {
			const status = await queryClient.fetchQuery({ ...authQueries.sudo(), staleTime: 0 });
			if (status.active) {
				action();
				return true;
			}
			if (status.methods.length === 0) {
				pushToast({ title: "Authentication unavailable", message: "This account has no available method for recent authentication.", tone: "critical" });
				return false;
			}

			let method: (typeof status.methods)[number] | null = status.methods.length === 1 && status.methods[0] === "password" ? "password" : null;
			if (!method) {
				const authMethods = await queryClient.fetchQuery(authQueries.methods());
				const providerNames = new Map(authMethods.providers.map(provider => [provider.id, provider.displayName]));
				method = (await choose({
					title: "Verify it’s you",
					message: "Sensitive account changes require recent authentication for five minutes. Choose a sign-in method to continue.",
					choices: status.methods.map(candidate => ({
						value: candidate,
						label: candidate === "password" ? "Use password" : `Continue with ${providerNames.get(candidate) ?? candidate}`
					}))
				})) as (typeof status.methods)[number] | null;
			}
			if (!method) {
				return false;
			}
			if (method === "password") {
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

			const url = new URL(absoluteExternalAuthStartUrl(method));
			url.searchParams.set("intent", "sudo");
			url.searchParams.set("returnTo", options.returnTo ?? returnTo ?? window.location.pathname);
			window.location.assign(url.toString());
			return false;
		} catch (error) {
			pushToast({ title: "Authentication failed", message: requestErrorMessage(error, "Could not confirm your identity."), tone: "critical" });
		}
		return false;
	};
}
