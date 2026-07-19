import { LanguageSwitcher } from "@/i18n/LanguageSwitcher";
import { pathForRoute } from "@/routes/routePaths";
import taiwanSubmarineCablesMap from "@/shared/assets/taiwan_submarine_cables.svg?url";
import { useTheme } from "@/shared/theme/useTheme";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { PageShell } from "@netstamp/ui";
import { useState, type ReactNode } from "react";
import { Helmet } from "react-helmet-async";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import styles from "./AuthPage.module.css";

const AUTH_CAPTION_KEYS = ["layout.caption1", "layout.caption2", "layout.caption3", "layout.caption4", "layout.caption5"] as const;

interface AuthLayoutProps {
	title: string;
	description?: string;
	children: ReactNode;
}

export function AuthLayout({ title, description, children }: AuthLayoutProps) {
	const { t } = useTranslation("auth");
	const { theme } = useTheme();
	const [captionKey] = useState(() => AUTH_CAPTION_KEYS[Math.floor(Math.random() * AUTH_CAPTION_KEYS.length)]);
	const logo = theme === "dark" ? netstampLogo : netstampLogoDark;

	return (
		<PageShell className={styles.authShell}>
			<Helmet>
				<meta name="description" content={t("layout.metaDescription")} />
			</Helmet>

			<div className={styles.authLayout}>
				<section className={styles.authFormPane} aria-labelledby="auth-title">
					<div className={styles.authTopBar}>
						<Link className={styles.brandLink} to={pathForRoute("login")} aria-label={t("layout.loginAria")}>
							<img className={styles.brandLogo} src={logo} alt="Netstamp" />
						</Link>
						<LanguageSwitcher className={styles.authLanguageSwitcher} />
					</div>

					<div className={styles.authCard}>
						<div className={styles.authHeader}>
							<h1 id="auth-title">{title}</h1>
							{description ? <p>{description}</p> : null}
						</div>

						{children}
					</div>
				</section>

				<figure className={styles.authVisual}>
					<img className={styles.authMap} src={taiwanSubmarineCablesMap} alt={t("layout.mapAlt")} loading="eager" decoding="async" />
					<figcaption className={styles.authQuote}>{t(captionKey)}</figcaption>
				</figure>
			</div>
		</PageShell>
	);
}
