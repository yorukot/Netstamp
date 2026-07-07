import type { Navigate } from "@/routes/routeTypes";
import { useConfirmEmailVerificationMutation } from "@/shared/api/mutations";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button } from "@netstamp/ui";
import { type FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

interface VerifyEmailPageProps {
	navigate: Navigate;
}

export function VerifyEmailPage({ navigate }: VerifyEmailPageProps) {
	const [searchParams] = useSearchParams();
	const token = searchParams.get("token") || "";
	const confirmEmail = useConfirmEmailVerificationMutation();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!token) {
			pushErrorToast("Verification link is missing a token.");
			return;
		}

		try {
			await confirmEmail.mutateAsync({ token });
			pushToast({
				title: "Email verified",
				message: "You can log in now.",
				tone: "success"
			});
			navigate("login");
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title="Verify your email" description="Confirm this account before signing in to Netstamp." helmetTitle="Verify email">
			{token ? (
				<form className={styles.form} onSubmit={handleSubmit}>
					<Button className={styles.submitButton} type="submit" size="lg" disabled={confirmEmail.isPending}>
						{confirmEmail.isPending ? "Verifying" : "Verify email"}
					</Button>
					{confirmEmail.isError ? <div className={styles.notice}>{requestErrorMessage(confirmEmail.error, "Email could not be verified.")}</div> : null}
				</form>
			) : (
				<div className={styles.notice}>Verification link is missing a token.</div>
			)}
			<Link className={styles.modeLink} to="/login">
				Return to login
			</Link>
		</AuthLayout>
	);
}
