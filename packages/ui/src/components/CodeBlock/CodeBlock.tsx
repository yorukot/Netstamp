import type { ComponentPropsWithoutRef, MouseEvent, ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { Button } from "../Button/Button";
import styles from "./CodeBlock.module.css";

export interface CodeBlockProps extends Omit<ComponentPropsWithoutRef<"pre">, "title"> {
	title?: ReactNode;
	meta?: ReactNode;
	actions?: ReactNode;
	copyable?: boolean;
	copyDisabled?: boolean;
	copyLabel?: string;
	copiedLabel?: string;
	children: ReactNode;
}

async function writeClipboardText(value: string) {
	if (navigator.clipboard?.writeText) {
		await navigator.clipboard.writeText(value);
		return;
	}

	const textarea = document.createElement("textarea");
	textarea.value = value;
	textarea.setAttribute("readonly", "");
	textarea.style.position = "fixed";
	textarea.style.left = "-100vw";
	document.body.append(textarea);
	textarea.select();
	document.execCommand("copy");
	textarea.remove();
}

export function CodeBlock({ title = "code block", meta, actions, copyable = true, copyDisabled = false, copyLabel = "Copy", copiedLabel = "Copied", className, children, ...props }: CodeBlockProps) {
	const [copied, setCopied] = useState(false);
	const timeoutRef = useRef<number | null>(null);
	const classes = [styles.code, className].filter(Boolean).join(" ");

	useEffect(() => {
		return () => {
			if (timeoutRef.current) {
				window.clearTimeout(timeoutRef.current);
			}
		};
	}, []);

	async function copyCode(event: MouseEvent<HTMLButtonElement>) {
		const block = event.currentTarget.closest<HTMLElement>("[data-ns-code-block]");
		const code = block?.querySelector("code")?.textContent ?? "";

		if (!code) {
			return;
		}

		try {
			await writeClipboardText(code);
		} catch {
			setCopied(false);
			return;
		}

		setCopied(true);

		if (timeoutRef.current) {
			window.clearTimeout(timeoutRef.current);
		}

		timeoutRef.current = window.setTimeout(() => {
			setCopied(false);
			timeoutRef.current = null;
		}, 1600);
	}

	return (
		<div className={["ns-frame", styles.block].join(" ")} data-ns-code-block>
			{title || meta || actions || copyable ? (
				<div className={styles.header}>
					<div className={styles.copy}>
						{title ? <strong>{title}</strong> : null}
						{meta ? <span>{meta}</span> : null}
					</div>
					{actions || copyable ? (
						<div className={styles.actions}>
							{actions}
							{copyable ? (
								<Button
									type="button"
									size="sm"
									variant="ghost"
									aria-label={copied ? copiedLabel : copyLabel}
									title={copied ? copiedLabel : copyLabel}
									disabled={copyDisabled}
									data-ns-code-copy
									data-copy-label={copyLabel}
									data-copied-label={copiedLabel}
									onClick={copyCode}
								>
									<span data-ns-code-copy-label>{copied ? copiedLabel : copyLabel}</span>
								</Button>
							) : null}
						</div>
					) : null}
				</div>
			) : null}
			<pre className={classes} {...props}>
				<code>{children}</code>
			</pre>
		</div>
	);
}
