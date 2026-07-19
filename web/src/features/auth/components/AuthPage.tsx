import { useAuth } from "@/features/auth/hooks/useAuth";
import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import { absoluteExternalAuthStartUrl, hasApiProblemCode } from "@/shared/api/client";
import { useCreateEmailVerificationMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { appFeatures, demoCredentials, demoMode } from "@/shared/config/features";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { Button, TextField } from "@netstamp/ui";
import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { GoogleLogoIcon } from "@phosphor-icons/react/dist/csr/GoogleLogo";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("auth");
	const isRegister = mode === "register";
	const { submitting, login, register } = useAuth();
	const createEmailVerification = useCreateEmailVerificationMutation();
	const methodsQuery = useQuery(authQueries.methods());
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [verificationEmail, setVerificationEmail] = useState("");
	const heading = isRegister ? t("login.signUp") : t("login.login");
	const showDemoCredentials = demoMode && Boolean(demoCredentials);
	const providers = methodsQuery.data?.providers ?? [];
	const callbackQuery = new URLSearchParams(window.location.search);
	const authError = callbackQuery.get("auth_error") || callbackQuery.get("oidc_error");
	const authProvider = callbackQuery.get("auth_provider");
	const failedProviderName = providers.find(provider => provider.id === authProvider)?.displayName || t("common.externalSignIn");

	function startExternalLogin(provider: string) {
		const url = new URL(absoluteExternalAuthStartUrl(provider));
		url.searchParams.set("intent", "login");
		url.searchParams.set("returnTo", "/");
		window.location.assign(url.toString());
	}

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
				pushErrorToast(t("common.passwordMinimum"));
				return;
			}

			if (password !== passwordAgain) {
				pushErrorToast(t("common.passwordMismatch"));
				return;
			}

			try {
				const result = await register({
					...payload,
					displayName: String(formData.get("displayName") || "")
				});
				if (result.emailVerificationRequired) {
					pushToast({
						title: t("login.verifyRegistrationTitle"),
						message: t("login.verifyRegistrationMessage"),
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
			if (hasApiProblemCode(error, "AUTH_EMAIL_VERIFICATION_REQUIRED")) {
				setVerificationEmail(email);
			}
			return;
		}
	}

	async function handleResendVerification() {
		const targetEmail = verificationEmail || email;
		if (!targetEmail) {
			pushErrorToast(t("login.emailRequired"));
			return;
		}

		try {
			await createEmailVerification.mutateAsync({ email: targetEmail });
			pushToast({
				title: t("login.verificationSentTitle"),
				message: t("login.verificationSentMessage"),
				tone: "success"
			});
		} catch {
			return;
		}
	}

	return (
		<AuthLayout title={heading}>
			{!isRegister && providers.length ? (
				<div className={styles.providerButtons}>
					{providers.map(provider => (
						<Button key={provider.id} type="button" variant="outline" size="lg" className={styles.providerButton} onClick={() => startExternalLogin(provider.id)}>
							{provider.id === "google" ? <GoogleLogoIcon size="1.25rem" weight="bold" aria-hidden="true" focusable="false" /> : null}
							{provider.id === "github" ? <GithubLogoIcon size="1.25rem" weight="bold" aria-hidden="true" focusable="false" /> : null}
							{t("login.continueWith", { provider: provider.displayName })}
						</Button>
					))}
				</div>
			) : null}
			{!isRegister && authError ? <div className={styles.notice}>{t("login.externalFailure", { provider: failedProviderName })}</div> : null}
			<form className={styles.form} onSubmit={handleSubmit}>
				{isRegister ? <TextField label={t("login.displayName")} name="displayName" type="text" autoComplete="name" /> : null}
				<TextField
					label={t("common.email")}
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
						label={t("common.password")}
						name="password"
						type="password"
						value={password}
						minLength={isRegister ? 8 : undefined}
						autoComplete={isRegister ? "new-password" : "current-password"}
						onChange={event => setPassword(event.currentTarget.value)}
					/>
					{!isRegister ? (
						<Link className={styles.inlineLink} to={pathForRoute("forgotPassword")} onFocus={preloadForgotPasswordPage} onPointerEnter={preloadForgotPasswordPage}>
							{t("login.forgotPassword")}
						</Link>
					) : null}
				</div>
				{isRegister ? <TextField label={t("login.passwordAgain")} name="passwordAgain" type="password" minLength={8} autoComplete="new-password" /> : null}
				{!isRegister && showDemoCredentials && demoCredentials ? (
					<div className={styles.demoCredentials} aria-label={t("login.demoCredentialsAria")}>
						<div className={styles.demoCredentialText}>
							<span>{t("login.demoAccount")}</span>
							<strong>{demoCredentials.email}</strong>
							<code>{demoCredentials.password}</code>
						</div>
						<Button type="button" variant="secondary" size="sm" onClick={fillDemoCredentials}>
							{t("login.useDemoAccount")}
						</Button>
					</div>
				) : null}
				<Button className={styles.submitButton} type="submit" size="lg" disabled={submitting}>
					{submitting ? t("common.submitting") : isRegister ? t("login.createAccount") : t("login.login")}
				</Button>
				{!isRegister && verificationEmail ? (
					<div className={styles.notice}>
						<div>{t("login.verificationRequired", { email: verificationEmail })}</div>
						<Button type="button" variant="secondary" size="sm" disabled={createEmailVerification.isPending} onClick={handleResendVerification}>
							{createEmailVerification.isPending ? t("login.sendingVerification") : t("login.resendVerification")}
						</Button>
					</div>
				) : null}
			</form>
			{isRegister || appFeatures.registration ? (
				<Link className={styles.modeLink} to={pathForRoute(isRegister ? "login" : "register")}>
					{isRegister ? t("login.login") : t("login.signUp")}
				</Link>
			) : null}
		</AuthLayout>
	);
}
