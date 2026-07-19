import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { StarIcon } from "@phosphor-icons/react/dist/csr/Star";
import styles from "./GlobalFooter.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";

export interface GlobalFooterProps {
	className?: string;
	madeByLabel?: string;
	separator?: string;
	contributorsSeparator?: string;
	contributorsLabel?: string;
	starLabel?: string;
}

export const GlobalFooter = ({
	className,
	madeByLabel = "Netstamp / Made by",
	separator = ", ",
	contributorsSeparator = ", and ",
	contributorsLabel = "contributors",
	starLabel = "Give us a star on GitHub"
}: GlobalFooterProps) => {
	const classes = [styles.footer, className].filter(Boolean).join(" ");

	return (
		<footer className={classes}>
			<div className={styles.footerBottom}>
				<span>
					{madeByLabel}{" "}
					<a href="https://github.com/elvisdragonmao" target="_blank" rel="noreferrer">
						Elvis Mao
					</a>
					{separator}
					<a href="https://github.com/yorukot" target="_blank" rel="noreferrer">
						Yorukot
					</a>
					{contributorsSeparator}
					<a href={githubUrl} target="_blank" rel="noreferrer">
						{contributorsLabel}
					</a>
				</span>
				<a href={githubUrl} target="_blank" rel="noreferrer">
					<StarIcon size={16} weight="fill" aria-hidden="true" focusable="false" />
					{starLabel}
					<GithubLogoIcon size={16} weight="fill" aria-hidden="true" focusable="false" />
				</a>
			</div>
		</footer>
	);
};
