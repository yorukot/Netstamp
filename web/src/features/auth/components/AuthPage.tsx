import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { Button, PageShell, Panel, TextField } from "@netstamp/ui";
import { useState, type FormEvent } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import taiwanSubmarineCablesMap from "../../../../../docs/src/assets/taiwan_submarine_cables.svg?url";
import { useAuth } from "../hooks/useAuth";
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
	const [caption] = useState(() => AUTH_CAPTIONS[Math.floor(Math.random() * AUTH_CAPTIONS.length)]);

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const formData = new FormData(event.currentTarget);
		const email = String(formData.get("email") || "");
		const password = String(formData.get("password") || "");
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

			<figure className={styles.authVisual}>
				<img className={styles.authMap} src={taiwanSubmarineCablesMap} alt="" aria-hidden="true" loading="eager" decoding="async" />
				<figcaption className={styles.authCaption}>{caption}</figcaption>
			</figure>

			<Panel className={styles.authCard} tone="glass" eyebrow="Account" title={isRegister ? "Sign up" : "Log in"}>
				<form className={styles.form} onSubmit={handleSubmit}>
					{isRegister ? <TextField label="Display Name" name="displayName" type="text" autoComplete="name" /> : null}
					<TextField
						label="Email"
						name="email"
						type="email"
						autoComplete={isRegister ? "email" : "username"}
						helper={isRegister ? "Use the email that will own the first project." : "Use the email connected to your controller account."}
					/>
					<TextField
						label="Password"
						name="password"
						type="password"
						minLength={isRegister ? 8 : undefined}
						autoComplete={isRegister ? "new-password" : "current-password"}
						helper={isRegister ? "Choose a password for controller access." : "Enter your account password."}
					/>
					{isRegister ? <TextField label="Password, again" name="passwordAgain" type="password" minLength={8} autoComplete="new-password" /> : null}
					<Button type="submit" size="lg" disabled={submitting}>
						{submitting ? "Submitting" : isRegister ? "Create project" : "Log in"}
					</Button>
				</form>
				<Link className={styles.modeLink} to={pathForRoute(isRegister ? "login" : "register")}>
					{isRegister ? "or log in" : "or sign up"}
				</Link>
				<div className={styles.homeAction}>
					<Button className={styles.homeButton} variant="secondary" size="lg" asChild>
						<a href="/">Go to home</a>
					</Button>
				</div>
			</Panel>
		</PageShell>
	);
}
