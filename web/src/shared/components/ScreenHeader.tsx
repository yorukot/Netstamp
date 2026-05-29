import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import styles from "./ScreenHeader.module.css";

interface ScreenHeaderProps {
	title: ReactNode;
	backLink?: {
		label: string;
		to: string;
	};
	actions?: ReactNode;
}

export function ScreenHeader({ title, backLink, actions }: ScreenHeaderProps) {
	return (
		<header className={styles.header}>
			<div className={styles.titleBlock}>
				{backLink ? (
					<Link className={styles.backLink} to={backLink.to}>
						{backLink.label}
					</Link>
				) : null}
				<h1>{title}</h1>
			</div>
			{actions ? <div className={styles.actions}>{actions}</div> : null}
		</header>
	);
}
