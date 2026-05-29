import type { ReactNode } from "react";
import styles from "./ScreenHeader.module.css";

interface ScreenHeaderProps {
	title: ReactNode;
	actions?: ReactNode;
}

export function ScreenHeader({ title, actions }: ScreenHeaderProps) {
	return (
		<header className={styles.header}>
			<div className={styles.titleBlock}>
				<h1>{title}</h1>
			</div>
			{actions ? <div className={styles.actions}>{actions}</div> : null}
		</header>
	);
}
