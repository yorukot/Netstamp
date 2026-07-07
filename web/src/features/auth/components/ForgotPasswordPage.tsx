import { pathForRoute } from "@/routes/routePaths";
import { useCreatePasswordResetMutation } from "@/shared/api/mutations";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

export function ForgotPasswordPage() {
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
		<AuthLayout title="Forgot Password" helmetTitle="Forgot Password">
			{submitted ? (
				<div className={styles.notice}>If an account exists for that email, reset instructions have been sent.</div>
			) : (
				<form className={styles.form} onSubmit={handleSubmit}>
					<TextField label="Email" name="email" type="email" value={email} autoComplete="username" onChange={event => setEmail(event.currentTarget.value)} />
					<Button className={styles.submitButton} type="submit" size="lg" disabled={createReset.isPending}>
						{createReset.isPending ? "Sending" : "Send reset link"}
					</Button>
					{createReset.isError ? <div className={styles.notice}>{requestErrorMessage(createReset.error, "Password reset could not be started.")}</div> : null}
				</form>
			)}
			<Link className={styles.modeLink} to={pathForRoute("login")}>
				Return to login
			</Link>
		</AuthLayout>
	);
}
