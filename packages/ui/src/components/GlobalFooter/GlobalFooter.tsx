import styles from "./GlobalFooter.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";

function StarIcon() {
	return (
		<svg aria-hidden="true" viewBox="0 0 256 256" width="16" height="16" fill="currentColor">
			<path d="M234.5 99.2a12 12 0 0 0-10.3-8.2l-59.8-5.2-23.3-55.3a12 12 0 0 0-22.2 0L95.6 85.8 35.8 91a12 12 0 0 0-6.9 21.1l45.4 39.3-13.6 58.5a12 12 0 0 0 17.9 13l51.4-30.2 51.4 30.2a12 12 0 0 0 17.9-13l-13.6-58.5 45.4-39.3a12 12 0 0 0 3.4-12.9Z" />
		</svg>
	);
}

function GithubIcon() {
	return (
		<svg aria-hidden="true" viewBox="0 0 256 256" width="16" height="16" fill="currentColor">
			<path d="M128 20a108 108 0 0 0-34.2 210.5c5.4 1 7.4-2.3 7.4-5.2v-20c-30.1 6.5-36.4-12.8-36.4-12.8-4.9-12.5-12-15.8-12-15.8-9.8-6.7.8-6.6.8-6.6 10.8.8 16.5 11.1 16.5 11.1 9.6 16.4 25.2 11.7 31.3 8.9 1-7 3.8-11.7 6.8-14.4-24-2.7-49.3-12-49.3-53.5 0-11.8 4.2-21.5 11.1-29.1-1.1-2.7-4.8-13.8 1.1-28.7 0 0 9.1-2.9 29.7 11.1a102.2 102.2 0 0 1 54.2 0c20.6-14 29.7-11.1 29.7-11.1 5.9 14.9 2.2 26 1.1 28.7 6.9 7.6 11.1 17.3 11.1 29.1 0 41.6-25.3 50.8-49.4 53.4 3.9 3.4 7.3 10 7.3 20.2v29.5c0 2.9 2 6.2 7.5 5.2A108 108 0 0 0 128 20Z" />
		</svg>
	);
}

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
					<StarIcon />
					Give us a star on GitHub
					<GithubIcon />
				</a>
			</div>
		</footer>
	);
}
