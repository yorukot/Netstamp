import type { Navigate } from "@/routes/routeTypes";
import { useConfirmEmailVerificationMutation } from "@/shared/api/mutations";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button } from "@netstamp/ui";
import { type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

interface VerifyEmailPageProps {
	navigate: Navigate;
}

export function VerifyEmailPage({ navigate }: VerifyEmailPageProps) {
	const { t } = useTranslation("auth");
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") || "";
	const confirmEmail = useConfirmEmailVerificationMutation();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!token) {
			pushErrorToast(t("verify.missingToken"));
			return;
		}

		try {
			await confirmEmail.mutateAsync({ token });
			pushToast({
				title: t("verify.successTitle"),
				message: t("verify.successMessage"),
				tone: "success"
			});
			navigate("login");
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title={t("verify.title")} description={t("verify.description")}>
			{token ? (
				<form className={styles.form} onSubmit={handleSubmit}>
					<Button className={styles.submitButton} type="submit" size="lg" disabled={confirmEmail.isPending}>
						{confirmEmail.isPending ? t("verify.verifying") : t("verify.verify")}
					</Button>
					{confirmEmail.isError ? <div className={styles.notice}>{requestErrorMessage(confirmEmail.error, t("verify.error"))}</div> : null}
				</form>
			) : (
				<div className={styles.notice}>{t("verify.missingToken")}</div>
			)}
			<Link className={styles.modeLink} to="/login">
				{t("common.returnToLogin")}
			</Link>
		</AuthLayout>
	);
}
