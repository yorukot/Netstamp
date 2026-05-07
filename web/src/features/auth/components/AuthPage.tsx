import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import { Badge, Button, PageShell, Panel, TextField } from "@netstamp/ui";
import type { FormEvent } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import { useAuthMock } from "../hooks/useAuthMock";
import styles from "./AuthPage.module.css";

interface AuthPageProps {
	mode?: "login" | "register";
	navigate: Navigate;
}

export function AuthPage({ mode = "login", navigate }: AuthPageProps) {
	const isRegister = mode === "register";
	const { submitting, login, register } = useAuthMock();

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const formData = new FormData(event.currentTarget);
		const email = String(formData.get("email") || "");
		const password = String(formData.get("password") || "");
		const payload = {
			email,
			password
		};

		if (isRegister) {
			await register({
				...payload,
				displayName: String(formData.get("displayName") || "")
			});
			navigate("onboarding");
			return;
		}

		await login(payload);
		navigate("dashboard");
	}

	return (
		<PageShell className={styles.authShell}>
			<Helmet>
				<title>{isRegister ? "Sign up" : "Log in"} - Netstamp</title>
				<meta name="description" content="Access the Netstamp distributed network observability console." />
			</Helmet>

			<section className={styles.authHero}>
				<Badge tone="accent">Controller access</Badge>
				<h1>{isRegister ? "Create your Netstamp workspace." : "Log in to your controller."}</h1>
				<p>
					{isRegister
						? "Start monitoring from probes you control. Set up your operator account, create a workspace, and connect your first probe."
						: "Review probe health, network checks, alerts, and recent results from your Netstamp controller."}
				</p>
			</section>

			<Panel className={styles.authCard} tone="glass" eyebrow="Account" title={isRegister ? "Sign up" : "Log in"}>
				<form className={styles.form} onSubmit={handleSubmit}>
					{isRegister ? <TextField label="Display Name" name="displayName" type="text" autoComplete="name" /> : null}
					<TextField
						label="Email"
						name="email"
						type="email"
						defaultValue={isRegister ? undefined : "elvis@netstamp.dev"}
						autoComplete={isRegister ? "email" : "username"}
						helper={isRegister ? "Use the email that will own this workspace." : "Use the email connected to your workspace."}
					/>
					<TextField
						label="Password"
						name="password"
						type="password"
						placeholder="***"
						autoComplete={isRegister ? "new-password" : "current-password"}
						helper={isRegister ? "Choose a password for controller access." : "Enter your account password."}
					/>
					{isRegister ? <TextField label="Password, again" name="passwordAgain" type="password" placeholder="***" autoComplete="new-password" /> : null}
					<Button type="submit" size="lg" disabled={submitting}>
						{submitting ? "Submitting" : isRegister ? "Create workspace" : "Log in"}
					</Button>
				</form>
				<Link className={styles.modeLink} to={pathForRoute(isRegister ? "login" : "register")}>
					{isRegister ? "or log in" : "or sign up"}
				</Link>
				<div className={styles.homeAction}>
					<Button className={styles.homeButton} variant="secondary" size="lg" onClick={() => navigate("landing")}>
						Go to home
					</Button>
				</div>
			</Panel>
		</PageShell>
	);
}
