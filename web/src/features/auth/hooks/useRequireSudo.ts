import { absoluteExternalAuthStartUrl } from "@/shared/api/client";
import { usePasswordSudoMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { useChoiceDialog, usePromptDialog } from "@/shared/components/confirmContext";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

export function useRequireSudo(returnTo?: string) {
	const { t } = useTranslation("auth");
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
				pushToast({ title: t("sudo.unavailableTitle"), message: t("sudo.unavailableMessage"), tone: "critical" });
				return false;
			}

			let method: (typeof status.methods)[number] | null = status.methods.length === 1 && status.methods[0] === "password" ? "password" : null;
			if (!method) {
				const authMethods = await queryClient.fetchQuery(authQueries.methods());
				const providerNames = new Map(authMethods.providers.map(provider => [provider.id, provider.displayName]));
				method = (await choose({
					title: t("sudo.verifyTitle"),
					message: t("sudo.chooseMethodMessage"),
					choices: status.methods.map(candidate => ({
						value: candidate,
						label: candidate === "password" ? t("sudo.usePassword") : t("sudo.continueWith", { provider: providerNames.get(candidate) ?? candidate })
					}))
				})) as (typeof status.methods)[number] | null;
			}
			if (!method) {
				return false;
			}
			if (method === "password") {
				const password = await prompt({
					title: t("sudo.confirmTitle"),
					message: t("sudo.confirmMessage"),
					inputLabel: t("sudo.currentPassword"),
					inputType: "password",
					confirmLabel: t("sudo.continue")
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
			pushToast({ title: t("sudo.failedTitle"), message: requestErrorMessage(error, t("sudo.failedMessage")), tone: "critical" });
		}
		return false;
	};
}
