import { pathForRoute } from "@/routes/routePaths";
import taiwanSubmarineCablesMap from "@/shared/assets/taiwan_submarine_cables.svg?url";
import { useTheme } from "@/shared/theme/useTheme";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { PageShell } from "@netstamp/ui";
import type { ReactNode } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import styles from "./AuthPage.module.css";

interface AuthLayoutProps {
	title: string;
	description?: string;
	helmetTitle: string;
	children: ReactNode;
}

export function AuthLayout({ title, description, helmetTitle, children }: AuthLayoutProps) {
	const { theme } = useTheme();
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
							{description ? <p>{description}</p> : null}
						</div>

						{children}
					</div>
				</section>

				<figure className={styles.authVisual}>
					<img className={styles.authMap} src={taiwanSubmarineCablesMap} alt="Taiwan submarine cable route map" loading="eager" decoding="async" />
				</figure>
			</div>
		</PageShell>
	);
}
