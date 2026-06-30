import { pathForRoute } from "@/routes/routePaths";
import taiwanSubmarineCablesMap from "@/shared/assets/taiwan_submarine_cables.svg?url";
import { useTheme } from "@/shared/theme/useTheme";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { PageShell } from "@netstamp/ui";
import { useState, type ReactNode } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import styles from "./AuthPage.module.css";

const AUTH_CAPTIONS = [
	"Major ISPs do not always choose routes based on the lowest latency. Cost, interconnection agreements, and commercial routing policies can also affect where your packets go.",
	"During daily peak hours, Taiwan's academic network often approaches its capacity limit, affecting connection quality to some overseas websites, academic databases, and cloud services.",
	"Did you know? Submarine cables are only a few centimeters thick, yet they carry almost all communications between Taiwan and the rest of the world.",
	"Did you know? More than 99% of global international data traffic is still carried by submarine cables.",
	"Did you know? Vietnam once had 3 of its 5 international submarine cables fail at the same time, making access to overseas services difficult."
] as const;

interface AuthLayoutProps {
	title: string;
	description: string;
	helmetTitle: string;
	children: ReactNode;
}

export function AuthLayout({ title, description, helmetTitle, children }: AuthLayoutProps) {
	const { theme } = useTheme();
	const [caption] = useState(() => AUTH_CAPTIONS[Math.floor(Math.random() * AUTH_CAPTIONS.length)]);
	const logo = theme === "dark" ? netstampLogo : netstampLogoDark;

	return (
		<PageShell className={styles.authShell}>
			<Helmet>
				<title>{helmetTitle} - Netstamp</title>
				<meta name="description" content="Access the Netstamp distributed network observability console." />
			</Helmet>

			<div className={styles.authLayout}>
				<section className={styles.authFormPane} aria-labelledby="auth-title">
					<Link className={styles.brandLink} to={pathForRoute("login")} aria-label="Netstamp login">
						<img className={styles.brandLogo} src={logo} alt="Netstamp" />
					</Link>

					<div className={styles.authCard}>
						<div className={styles.authHeader}>
							<h1 id="auth-title">{title}</h1>
							<p>{description}</p>
						</div>

						{children}
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
