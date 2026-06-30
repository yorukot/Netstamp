import type { Navigate } from "@/routes/routeTypes";
import { useConfirmPasswordResetMutation } from "@/shared/api/mutations";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

interface ResetPasswordPageProps {
	navigate: Navigate;
}

export function ResetPasswordPage({ navigate }: ResetPasswordPageProps) {
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") || "";
	const [newPassword, setNewPassword] = useState("");
	const [newPasswordAgain, setNewPasswordAgain] = useState("");
	const confirmReset = useConfirmPasswordResetMutation();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!token) {
			pushErrorToast("Reset link is missing a token.");
			return;
		}
		if (newPassword.length < 8) {
			pushErrorToast("Password must be at least 8 characters.");
			return;
		}
		if (newPassword !== newPasswordAgain) {
			pushErrorToast("Password confirmation does not match.");
			return;
		}

		try {
			await confirmReset.mutateAsync({ token, newPassword });
			pushToast({
				title: "Password updated",
				message: "Log in with your new password.",
				tone: "success"
			});
			navigate("login");
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title="Set a new password" description="Choose a new controller password for your Netstamp account." helmetTitle="Set new password">
			{token ? (
				<form className={styles.form} onSubmit={handleSubmit}>
					<TextField
						label="New password"
						name="newPassword"
						type="password"
						value={newPassword}
						minLength={8}
						autoComplete="new-password"
						helper="Use at least 8 characters."
						onChange={event => setNewPassword(event.currentTarget.value)}
					/>
					<TextField
						label="New password, again"
						name="newPasswordAgain"
						type="password"
						value={newPasswordAgain}
						minLength={8}
						autoComplete="new-password"
						onChange={event => setNewPasswordAgain(event.currentTarget.value)}
					/>
					<Button className={styles.submitButton} type="submit" size="lg" disabled={confirmReset.isPending}>
						{confirmReset.isPending ? "Updating" : "Update password"}
					</Button>
					{confirmReset.isError ? <div className={styles.notice}>{requestErrorMessage(confirmReset.error, "Password could not be updated.")}</div> : null}
				</form>
			) : (
				<div className={styles.notice}>Reset link is missing a token.</div>
			)}
			<Link className={styles.modeLink} to="/login">
				Return to login
			</Link>
		</AuthLayout>
	);
}
