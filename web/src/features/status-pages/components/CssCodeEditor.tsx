import { useMemo, useRef, type ReactNode, type UIEvent } from "react";
import styles from "./CssCodeEditor.module.css";

interface CssCodeEditorProps {
	value: string;
	onChange: (value: string) => void;
}

const tokenPattern = /(\/\*[\s\S]*?\*\/|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|#[\da-fA-F]{3,8}\b|--[\w-]+|(?:\d*\.)?\d+(?:px|rem|em|%|s|ms|vh|vw)?\b|[{}:;,])/g;

function tokenClass(token: string) {
	if (token.startsWith("/*")) return styles.comment;
	if (token.startsWith('"') || token.startsWith("'")) return styles.string;
	if (token.startsWith("#")) return styles.color;
	if (token.startsWith("--")) return styles.variable;
	if (/^\d/.test(token) || token.startsWith(".")) return styles.number;
	if (/^[{}:;,]$/.test(token)) return styles.punctuation;
	return undefined;
}

function highlightedCSS(value: string) {
	const nodes: ReactNode[] = [];
	let cursor = 0;
	let key = 0;

	for (const match of value.matchAll(tokenPattern)) {
		const index = match.index ?? 0;
		if (index > cursor) nodes.push(value.slice(cursor, index));
		const token = match[0];
		nodes.push(
			<span key={key++} className={tokenClass(token)}>
				{token}
			</span>
		);
		cursor = index + token.length;
	}
	if (cursor < value.length) nodes.push(value.slice(cursor));
	return nodes;
}

export function CssCodeEditor({ value, onChange }: CssCodeEditorProps) {
	const highlightRef = useRef<HTMLPreElement | null>(null);
	const highlighted = useMemo(() => highlightedCSS(value), [value]);

	function syncScroll(event: UIEvent<HTMLTextAreaElement>) {
		if (!highlightRef.current) return;
		highlightRef.current.scrollTop = event.currentTarget.scrollTop;
		highlightRef.current.scrollLeft = event.currentTarget.scrollLeft;
	}

	return (
		<div className={styles.field}>
			<div className={styles.labelRow}>
				<label htmlFor="status-page-custom-css">Custom CSS</label>
				<span>CSS</span>
			</div>
			<div className={styles.editor}>
				<pre ref={highlightRef} aria-hidden="true" className={styles.highlight}>
					<code>{highlighted}</code>
				</pre>
				<textarea
					id="status-page-custom-css"
					value={value}
					spellCheck={false}
					aria-describedby="status-page-custom-css-help"
					onChange={event => onChange(event.currentTarget.value)}
					onScroll={syncScroll}
				/>
			</div>
			<p id="status-page-custom-css-help">Applied after the built-in public page styles.</p>
		</div>
	);
}
