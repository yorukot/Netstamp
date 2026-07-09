import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { StarIcon } from "@phosphor-icons/react/dist/csr/Star";
import styles from "./GlobalFooter.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";

export interface GlobalFooterProps {
	className?: string;
}

export function GlobalFooter({ className }: GlobalFooterProps) {
	const classes = [styles.footer, className].filter(Boolean).join(" ");

	return (
		<footer className={classes}>
			<div className={styles.footerBottom}>
				<span>
					Netstamp / Made by{" "}
					<a href="https://github.com/elvisdragonmao" target="_blank" rel="noreferrer">
						Elvis Mao
					</a>
					,{" "}
					<a href="https://github.com/yorukot" target="_blank" rel="noreferrer">
						Yorukot
					</a>
					, and{" "}
					<a href={githubUrl} target="_blank" rel="noreferrer">
						contributors
					</a>
				</span>
				<a href={githubUrl} target="_blank" rel="noreferrer">
					<StarIcon size={16} weight="fill" aria-hidden="true" focusable="false" />
					Give us a star on GitHub
					<GithubLogoIcon size={16} weight="fill" aria-hidden="true" focusable="false" />
				</a>
			</div>
		</footer>
	);
}
