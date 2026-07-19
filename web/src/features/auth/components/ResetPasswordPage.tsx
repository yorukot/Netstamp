import type { Navigate } from "@/routes/routeTypes";
import { useConfirmPasswordResetMutation } from "@/shared/api/mutations";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

interface ResetPasswordPageProps {
	navigate: Navigate;
}

export function ResetPasswordPage({ navigate }: ResetPasswordPageProps) {
	const { t } = useTranslation("auth");
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") || "";
	const [newPassword, setNewPassword] = useState("");
	const [newPasswordAgain, setNewPasswordAgain] = useState("");
	const confirmReset = useConfirmPasswordResetMutation();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!token) {
			pushErrorToast(t("reset.missingToken"));
			return;
		}
		if (newPassword.length < 8) {
			pushErrorToast(t("common.passwordMinimum"));
			return;
		}
		if (newPassword !== newPasswordAgain) {
			pushErrorToast(t("common.passwordMismatch"));
			return;
		}

		try {
			await confirmReset.mutateAsync({ token, newPassword });
			pushToast({
				title: t("reset.successTitle"),
				message: t("reset.successMessage"),
				tone: "success"
			});
			navigate("login");
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title={t("reset.title")} description={t("reset.description")}>
			{token ? (
				<form className={styles.form} onSubmit={handleSubmit}>
					<TextField
						label={t("reset.newPassword")}
						name="newPassword"
						type="password"
						value={newPassword}
						minLength={8}
						autoComplete="new-password"
						helper={t("reset.helper")}
						onChange={event => setNewPassword(event.currentTarget.value)}
					/>
					<TextField
						label={t("reset.newPasswordAgain")}
						name="newPasswordAgain"
						type="password"
						value={newPasswordAgain}
						minLength={8}
						autoComplete="new-password"
						onChange={event => setNewPasswordAgain(event.currentTarget.value)}
					/>
					<Button className={styles.submitButton} type="submit" size="lg" disabled={confirmReset.isPending}>
						{confirmReset.isPending ? t("common.updating") : t("reset.update")}
					</Button>
					{confirmReset.isError ? <div className={styles.notice}>{requestErrorMessage(confirmReset.error, t("reset.error"))}</div> : null}
				</form>
			) : (
				<div className={styles.notice}>{t("reset.missingToken")}</div>
			)}
			<Link className={styles.modeLink} to="/login">
				{t("common.returnToLogin")}
			</Link>
		</AuthLayout>
	);
}
