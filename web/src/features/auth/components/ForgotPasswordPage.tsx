import { pathForRoute } from "@/routes/routePaths";
import { useCreatePasswordResetMutation } from "@/shared/api/mutations";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

export function ForgotPasswordPage() {
	const { t } = useTranslation("auth");
	const [email, setEmail] = useState("");
	const [submitted, setSubmitted] = useState(false);
	const createReset = useCreatePasswordResetMutation();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		try {
			await createReset.mutateAsync({ email });
			setSubmitted(true);
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title={t("forgot.title")}>
			{submitted ? (
				<div className={styles.notice}>{t("forgot.sent")}</div>
			) : (
				<form className={styles.form} onSubmit={handleSubmit}>
					<TextField label={t("common.email")} name="email" type="email" value={email} autoComplete="username" onChange={event => setEmail(event.currentTarget.value)} />
					<Button className={styles.submitButton} type="submit" size="lg" disabled={createReset.isPending}>
						{createReset.isPending ? t("common.sending") : t("forgot.sendLink")}
					</Button>
					{createReset.isError ? <div className={styles.notice}>{requestErrorMessage(createReset.error, t("forgot.error"))}</div> : null}
				</form>
			)}
			<Link className={styles.modeLink} to={pathForRoute("login")}>
				{t("common.returnToLogin")}
			</Link>
		</AuthLayout>
	);
}
