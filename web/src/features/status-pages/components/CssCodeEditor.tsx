import { useLayoutEffect, useMemo, useRef, type KeyboardEvent, type ReactNode, type UIEvent } from "react";
import styles from "./CssCodeEditor.module.css";

interface CssCodeEditorProps {
	value: string;
	onChange: (value: string) => void;
}

const cssTokenPattern =
	/(\/\*[\s\S]*?\*\/|"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|#[\da-fA-F]{3,8}\b|@[a-zA-Z-]+|--[\w-]+|(?:\d*\.)?\d+(?:px|rem|em|%|s|ms|vh|vw|svh|dvh|deg|fr)?\b|[a-zA-Z_][\w-]*|[{}:;,()[\]]|\s+|.)/gs;

function tokenClass(token: string, context: "rule" | "value", remaining: string) {
	if (token.startsWith("/*")) return styles.comment;
	if (token.startsWith('"') || token.startsWith("'")) return styles.string;
	if (token.startsWith("#")) return styles.color;
	if (token.startsWith("@")) return styles.atRule;
	if (token.startsWith("--")) return styles.variable;
	if (/^(?:\d|\.\d)/.test(token)) return styles.number;
	if (/^[{}:;,()[\]]$/.test(token)) return styles.punctuation;
	if (/^[a-zA-Z_][\w-]*$/.test(token)) {
		if (/^\s*\(/.test(remaining)) return styles.function;
		if (/^\s*:/.test(remaining)) return styles.property;
		return context === "value" ? styles.keyword : styles.selector;
	}
	return undefined;
}

function highlightedCSS(value: string) {
	const nodes: ReactNode[] = [];
	let key = 0;
	let context: "rule" | "value" = "rule";

	for (const match of value.matchAll(cssTokenPattern)) {
		const token = match[0];
		const remaining = value.slice((match.index ?? 0) + token.length);
		const className = tokenClass(token, context, remaining);
		if (className) {
			nodes.push(
				<span key={key++} className={className}>
					{token}
				</span>
			);
		} else {
			nodes.push(token);
		}

		if (token === ":") context = "value";
		if (token === ";" || token === "{" || token === "}") context = "rule";
	}

	return nodes;
}

export function CssCodeEditor({ value, onChange }: CssCodeEditorProps) {
	const textareaRef = useRef<HTMLTextAreaElement | null>(null);
	const highlightRef = useRef<HTMLPreElement | null>(null);
	const pendingSelection = useRef<{ start: number; end: number } | null>(null);
	const highlighted = useMemo(() => highlightedCSS(value), [value]);

	useLayoutEffect(() => {
		const selection = pendingSelection.current;
		if (!selection || !textareaRef.current) return;
		textareaRef.current.setSelectionRange(selection.start, selection.end);
		pendingSelection.current = null;
	}, [value]);

	function commit(nextValue: string, start: number, end = start) {
		pendingSelection.current = { start, end };
		onChange(nextValue);
	}

	function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
		if (event.metaKey || event.ctrlKey || event.altKey) return;
		const textarea = event.currentTarget;
		const start = textarea.selectionStart;
		const end = textarea.selectionEnd;
		const selected = value.slice(start, end);

		if (event.key === "{") {
			event.preventDefault();
			commit(`${value.slice(0, start)}{${selected}}${value.slice(end)}`, start + 1, selected ? end + 1 : start + 1);
			return;
		}

		if (event.key === "}" && start === end && value[start] === "}") {
			event.preventDefault();
			textarea.setSelectionRange(start + 1, start + 1);
			return;
		}

		if (event.key === "Enter" && start === end && value[start - 1] === "{" && value[start] === "}") {
			event.preventDefault();
			const lineStart = value.lastIndexOf("\n", start - 1) + 1;
			const indent = value.slice(lineStart, start - 1).match(/^\s*/)?.[0] ?? "";
			const insertion = `\n${indent}\t\n${indent}`;
			commit(`${value.slice(0, start)}${insertion}${value.slice(end)}`, start + indent.length + 2);
			return;
		}

		if (event.key === "Tab") {
			event.preventDefault();
			commit(`${value.slice(0, start)}\t${value.slice(end)}`, start + 1);
		}
	}

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
					ref={textareaRef}
					id="status-page-custom-css"
					value={value}
					spellCheck={false}
					aria-describedby="status-page-custom-css-help"
					onChange={event => onChange(event.currentTarget.value)}
					onKeyDown={handleKeyDown}
					onScroll={syncScroll}
				/>
			</div>
			<p id="status-page-custom-css-help">
				Applied after built-in styles. Use semantic elements or stable hooks such as <code>.ns-status-page</code> and <code>.ns-status-block</code>.
			</p>
		</div>
	);
}
