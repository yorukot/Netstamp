import { pathForRoute } from "@/routes/routePaths";
import { clearAuthenticatedClientState } from "@/shared/api/mutations/auth";
import { subscribeToSessionUnavailable } from "@/shared/api/sessionUnavailable";
import { useConfirm } from "@/shared/components/confirmContext";
import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useLayoutEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useSession } from "./SessionContext";

export function SessionExpiryPrompt() {
	const { t } = useTranslation("auth");
	const { session } = useSession();
	const confirm = useConfirm();
	const queryClient = useQueryClient();
	const sessionRef = useRef(session);
	const promptingRef = useRef(false);
	const promptedUserRef = useRef<string | null>(null);

	useLayoutEffect(() => {
		sessionRef.current = session;
		if (!session) {
			promptedUserRef.current = null;
		}
	}, [session]);

	useEffect(
		() =>
			subscribeToSessionUnavailable(() => {
				const userId = sessionRef.current?.user.id;
				if (!userId || promptingRef.current || promptedUserRef.current === userId) {
					return;
				}

				promptingRef.current = true;
				promptedUserRef.current = userId;
				void confirm({
					title: t("session.expiredTitle"),
					message: t("session.expiredMessage"),
					confirmLabel: t("session.signInAgain"),
					cancelLabel: t("session.notNow")
				})
					.then(shouldSignIn => {
						if (!shouldSignIn) {
							return;
						}

						clearAuthenticatedClientState(queryClient);
						window.location.assign(pathForRoute("login"));
					})
					.finally(() => {
						promptingRef.current = false;
					});
			}),
		[confirm, queryClient, t]
	);

	return null;
}
