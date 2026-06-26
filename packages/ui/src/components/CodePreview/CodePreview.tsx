import type { ComponentPropsWithoutRef, ReactNode } from "react";
import styles from "./CodePreview.module.css";

export interface CodePreviewProps extends Omit<ComponentPropsWithoutRef<"pre">, "title"> {
	title?: ReactNode;
	meta?: ReactNode;
	actions?: ReactNode;
	children: ReactNode;
}

export function CodePreview({ title, meta, actions, className, children, ...props }: CodePreviewProps) {
	return (
		<div className={["ns-frame", styles.preview].join(" ")}>
			{title || meta || actions ? (
				<div className={styles.header}>
					<div className={styles.copy}>
						{title ? <strong>{title}</strong> : null}
						{meta ? <span>{meta}</span> : null}
					</div>
					{actions ? <div className={styles.actions}>{actions}</div> : null}
				</div>
			) : null}
			<pre className={[styles.code, className].filter(Boolean).join(" ")} {...props}>
				<code>{children}</code>
			</pre>
		</div>
	);
}
