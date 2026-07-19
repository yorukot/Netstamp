import { BookOpenTextIcon } from "@phosphor-icons/react/dist/csr/BookOpenText";
import { BracketsCurlyIcon } from "@phosphor-icons/react/dist/csr/BracketsCurly";
import { DiscordLogoIcon } from "@phosphor-icons/react/dist/csr/DiscordLogo";
import { FileTextIcon } from "@phosphor-icons/react/dist/csr/FileText";
import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { StarIcon } from "@phosphor-icons/react/dist/csr/Star";
import styles from "./GlobalFooter.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";
const discordUrl = "https://discord.gg/9mdkf6dyTy";
const licenseUrl = `${githubUrl}/blob/main/LICENSE`;

export interface GlobalFooterProps {
	className?: string;
	eyebrowLabel?: string;
	description?: string;
	navigationLabel?: string;
	exploreLabel?: string;
	docsLabel?: string;
	docsHref?: string;
	apiLabel?: string;
	apiHref?: string;
	projectLabel?: string;
	githubLabel?: string;
	discordLabel?: string;
	licenseLabel?: string;
	openSourceLabel?: string;
	starLabel?: string;
}

export const GlobalFooter = ({
	className,
	eyebrowLabel = "Network observability",
	description = "Measure latency, packet loss, DNS, and routes from probes you control.",
	navigationLabel = "Footer navigation",
	exploreLabel = "Explore",
	docsLabel = "Documentation",
	docsHref = "/docs/",
	apiLabel = "API reference",
	apiHref = "/openapi/",
	projectLabel = "Project",
	githubLabel = "GitHub",
	discordLabel = "Discord",
	licenseLabel = "Apache 2.0 license",
	openSourceLabel = "Open source / Self-hosted",
	starLabel = "Star Netstamp on GitHub"
}: GlobalFooterProps) => {
	const classes = [styles.footer, className].filter(Boolean).join(" ");

	return (
		<footer className={classes}>
			<div className={styles.inner}>
				<div className={styles.brandBlock}>
					<div className={styles.brandLine}>
						<span className={styles.brandMarker} aria-hidden="true" />
						<strong>Netstamp</strong>
						<span className={styles.eyebrow}>{eyebrowLabel}</span>
					</div>
					<p>{description}</p>
				</div>

				<nav className={styles.navigation} aria-label={navigationLabel}>
					<div className={styles.linkGroup}>
						<span className={styles.groupLabel}>{exploreLabel}</span>
						<a href={docsHref}>
							<BookOpenTextIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
							{docsLabel}
						</a>
						<a href={apiHref}>
							<BracketsCurlyIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
							{apiLabel}
						</a>
					</div>

					<div className={styles.linkGroup}>
						<span className={styles.groupLabel}>{projectLabel}</span>
						<a href={githubUrl} target="_blank" rel="noreferrer">
							<GithubLogoIcon size="1rem" weight="fill" aria-hidden="true" focusable="false" />
							{githubLabel}
						</a>
						<a href={discordUrl} target="_blank" rel="noreferrer">
							<DiscordLogoIcon size="1rem" weight="fill" aria-hidden="true" focusable="false" />
							{discordLabel}
						</a>
						<a href={licenseUrl} target="_blank" rel="noreferrer">
							<FileTextIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
							{licenseLabel}
						</a>
					</div>
				</nav>

				<div className={styles.footerBottom}>
					<span>{openSourceLabel}</span>
					<a href={githubUrl} target="_blank" rel="noreferrer">
						<StarIcon size="1rem" weight="fill" aria-hidden="true" focusable="false" />
						{starLabel}
					</a>
				</div>
			</div>
		</footer>
	);
};
