import type { ComponentPropsWithoutRef, ReactNode } from "react";
import styles from "./Terminal.module.css";

export interface TerminalProps extends Omit<ComponentPropsWithoutRef<"pre">, "title"> {
	title?: ReactNode;
	meta?: ReactNode;
}

export function Terminal({ title = "controller shell", meta, className, children, ...props }: TerminalProps) {
	const classes = [styles.terminal, className].filter(Boolean).join(" ");

	return (
		<div className={["ns-frame", styles.shell].join(" ")}>
			<div className={styles.bar}>
				<span aria-hidden="true" />
				<span aria-hidden="true" />
				<span aria-hidden="true" />
				<strong>{title}</strong>
				{meta ? <em>{meta}</em> : null}
			</div>
			<pre className={classes} {...props}>
				<code>{children}</code>
			</pre>
		</div>
	);
}
