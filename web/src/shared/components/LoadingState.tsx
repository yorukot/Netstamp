import { classNames } from "@/shared/utils/classNames";
import styles from "./LoadingState.module.css";

interface LoadingStateProps {
	label: string;
	detail?: string;
	className?: string;
	size?: "compact" | "chart";
}

export function LoadingState({ label, detail, className, size = "chart" }: LoadingStateProps) {
	return (
		<div className={classNames(styles.loadingState, styles[`loadingState${size}`], className)} role="status" aria-live="polite">
			<div className={styles.loadingCopy}>
				<strong>{label}</strong>
				{detail ? <span>{detail}</span> : null}
			</div>
			<div className={styles.loadingRows} aria-hidden="true">
				<span />
				<span />
				<span />
			</div>
		</div>
	);
}
