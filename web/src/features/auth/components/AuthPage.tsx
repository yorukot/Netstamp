import { useAuth } from "@/features/auth/hooks/useAuth";
import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import { ApiError } from "@/shared/api/client";
import { useCreateEmailVerificationMutation } from "@/shared/api/mutations";
import { appFeatures, demoCredentials, demoMode } from "@/shared/config/features";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { Button, TextField } from "@netstamp/ui";
import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { AuthLayout } from "./AuthLayout";
import styles from "./AuthPage.module.css";

interface AuthPageProps {
	mode?: "login" | "register";
	navigate: Navigate;
}

function preloadForgotPasswordPage() {
	void import("./ForgotPasswordPage");
}

export function AuthPage({ mode = "login", navigate }: AuthPageProps) {
	const isRegister = mode === "register";
	const { submitting, login, register } = useAuth();
	const createEmailVerification = useCreateEmailVerificationMutation();
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [verificationEmail, setVerificationEmail] = useState("");
	const heading = isRegister ? "Sign Up" : "Login";
	const showDemoCredentials = demoMode && Boolean(demoCredentials);

	useEffect(() => {
		if (!isRegister) {
			preloadForgotPasswordPage();
		}
	}, [isRegister]);

	function fillDemoCredentials() {
		if (!demoCredentials) {
			return;
		}

		setEmail(demoCredentials.email);
		setPassword(demoCredentials.password);
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const formData = new FormData(event.currentTarget);
		const passwordAgain = String(formData.get("passwordAgain") || "");
		const payload = {
			email,
			password
		};

		if (isRegister) {
			if (password.length < 8) {
				pushErrorToast("Password must be at least 8 characters.");
				return;
			}

			if (password !== passwordAgain) {
				pushErrorToast("Password confirmation does not match.");
				return;
			}

			try {
				const result = await register({
					...payload,
					displayName: String(formData.get("displayName") || "")
				});
				if (result.emailVerificationRequired) {
					pushToast({
						title: "Verify your email",
						message: "We sent a verification link before you can log in.",
						tone: "success"
					});
					navigate("login");
					return;
				}
				navigate("onboarding");
			} catch {
				return;
			}
			return;
		}

		try {
			await login(payload);
			navigate("dashboard");
		} catch (error) {
			if (error instanceof ApiError && error.status === 403) {
				setVerificationEmail(email);
			}
			return;
		}
	}

	async function handleResendVerification() {
		const targetEmail = verificationEmail || email;
		if (!targetEmail) {
			pushErrorToast("Email is required.");
			return;
		}

		try {
			await createEmailVerification.mutateAsync({ email: targetEmail });
			pushToast({
				title: "Verification email sent",
				message: "Check your inbox for the latest verification link.",
				tone: "success"
			});
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title={heading} helmetTitle={isRegister ? "Sign Up" : "Login"}>
			<form className={styles.form} onSubmit={handleSubmit}>
				{isRegister ? <TextField label="Display Name" name="displayName" type="text" autoComplete="name" /> : null}
				<TextField
					label="Email"
					name="email"
					type="email"
					value={email}
					autoComplete={isRegister ? "email" : "username"}
					onChange={event => {
						setEmail(event.currentTarget.value);
						setVerificationEmail("");
					}}
				/>
				<div className={styles.passwordFieldGroup}>
					<TextField
						label="Password"
						name="password"
						type="password"
						value={password}
						minLength={isRegister ? 8 : undefined}
						autoComplete={isRegister ? "new-password" : "current-password"}
						onChange={event => setPassword(event.currentTarget.value)}
					/>
					{!isRegister ? (
						<Link className={styles.inlineLink} to={pathForRoute("forgotPassword")} onFocus={preloadForgotPasswordPage} onPointerEnter={preloadForgotPasswordPage}>
							Forgot password?
						</Link>
					) : null}
				</div>
				{isRegister ? <TextField label="Password, again" name="passwordAgain" type="password" minLength={8} autoComplete="new-password" /> : null}
				{!isRegister && showDemoCredentials && demoCredentials ? (
					<div className={styles.demoCredentials} aria-label="Demo account credentials">
						<div className={styles.demoCredentialText}>
							<span>Demo account</span>
							<strong>{demoCredentials.email}</strong>
							<code>{demoCredentials.password}</code>
						</div>
						<Button type="button" variant="secondary" size="sm" onClick={fillDemoCredentials}>
							Use demo account
						</Button>
					</div>
				) : null}
				<Button className={styles.submitButton} type="submit" size="lg" disabled={submitting}>
					{submitting ? "Submitting" : isRegister ? "Create Account" : "Log in"}
				</Button>
				{!isRegister && verificationEmail ? (
					<div className={styles.notice}>
						<div>Email verification is required for {verificationEmail}.</div>
						<Button type="button" variant="secondary" size="sm" disabled={createEmailVerification.isPending} onClick={handleResendVerification}>
							{createEmailVerification.isPending ? "Sending" : "Resend verification email"}
						</Button>
					</div>
				) : null}
			</form>
			{isRegister || appFeatures.registration ? (
				<Link className={styles.modeLink} to={pathForRoute(isRegister ? "login" : "register")}>
					{isRegister ? "Login" : "Sign Up"}
				</Link>
			) : null}
		</AuthLayout>
	);
}
