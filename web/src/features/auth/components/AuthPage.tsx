import { useAuth } from "@/features/auth/hooks/useAuth";
import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import taiwanSubmarineCablesMap from "@/shared/assets/taiwan_submarine_cables.svg?url";
import { appFeatures, demoCredentials, demoMode } from "@/shared/config/features";
import { useTheme } from "@/shared/theme/useTheme";
import { pushErrorToast } from "@/shared/toast/toastStore";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Button, PageShell, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import styles from "./AuthPage.module.css";

const AUTH_CAPTIONS = [
	"Major ISPs do not always choose routes based on the lowest latency. Cost, interconnection agreements, and commercial routing policies can also affect where your packets go.",
	"During daily peak hours, Taiwan’s academic network often approaches its capacity limit, affecting connection quality to some overseas websites, academic databases, and cloud services.",
	"Did you know? Submarine cables are only a few centimeters thick, yet they carry almost all communications between Taiwan and the rest of the world.",
	"Did you know? More than 99% of global international data traffic is still carried by submarine cables.",
	"Did you know? Vietnam once had 3 of its 5 international submarine cables fail at the same time, making access to overseas services difficult."
] as const;

interface AuthPageProps {
	mode?: "login" | "register";
	navigate: Navigate;
}

export function AuthPage({ mode = "login", navigate }: AuthPageProps) {
	const isRegister = mode === "register";
	const { submitting, login, register } = useAuth();
	const { theme } = useTheme();
	const [caption] = useState(() => AUTH_CAPTIONS[Math.floor(Math.random() * AUTH_CAPTIONS.length)]);
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const heading = isRegister ? "Create your Netstamp account" : "Log in to your account";
	const intro = isRegister ? "Create controller access and start with your first project." : "Enter your credentials to access the Netstamp console.";
	const logo = theme === "dark" ? netstampLogo : netstampLogoDark;
	const showDemoCredentials = demoMode && Boolean(demoCredentials);

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
				await register({
					...payload,
					displayName: String(formData.get("displayName") || "")
				});
				navigate("onboarding");
			} catch {
				return;
			}
			return;
		}

		try {
			await login(payload);
			navigate("dashboard");
		} catch {
			return;
		}
	}

	return (
		<PageShell className={styles.authShell}>
			<Helmet>
				<title>{isRegister ? "Sign up" : "Log in"} - Netstamp</title>
				<meta name="description" content="Access the Netstamp distributed network observability console." />
			</Helmet>

			<div className={styles.authLayout}>
				<section className={styles.authFormPane} aria-labelledby="auth-title">
					<a className={styles.brandLink} href="/" aria-label="Netstamp home">
						<img className={styles.brandLogo} src={logo} alt="Netstamp" />
					</a>

					<div className={styles.authCard}>
						<div className={styles.authHeader}>
							<h1 id="auth-title">{heading}</h1>
							<p>{intro}</p>
						</div>

						<form className={styles.form} onSubmit={handleSubmit}>
							{isRegister ? <TextField label="Display Name" name="displayName" type="text" autoComplete="name" /> : null}
							<TextField
								label="Email"
								name="email"
								type="email"
								value={email}
								autoComplete={isRegister ? "email" : "username"}
								helper={isRegister ? "Use the email that will own the first project." : "Use the email connected to your controller account."}
								onChange={event => setEmail(event.currentTarget.value)}
							/>
							<TextField
								label="Password"
								name="password"
								type="password"
								value={password}
								minLength={isRegister ? 8 : undefined}
								autoComplete={isRegister ? "new-password" : "current-password"}
								helper={isRegister ? "Choose a password for controller access." : "Enter your account password."}
								onChange={event => setPassword(event.currentTarget.value)}
							/>
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
								{submitting ? "Submitting" : isRegister ? "Create project" : "Log in"}
							</Button>
						</form>
						{isRegister || appFeatures.registration ? (
							<Link className={styles.modeLink} to={pathForRoute(isRegister ? "login" : "register")}>
								{isRegister ? "Already have an account? Log in" : "Do not have an account? Sign up"}
							</Link>
						) : null}
						<div className={styles.homeAction}>
							<Button className={styles.homeButton} variant="secondary" size="lg" asChild>
								<a href="/">Go to home</a>
							</Button>
						</div>
					</div>
				</section>

				<figure className={styles.authVisual}>
					<img className={styles.authMap} src={taiwanSubmarineCablesMap} alt="Taiwan submarine cable route map" loading="eager" decoding="async" />
					<figcaption className={styles.authQuote}>{caption}</figcaption>
				</figure>
			</div>
		</PageShell>
	);
}
