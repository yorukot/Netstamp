import type { ComponentPropsWithoutRef, KeyboardEvent, ReactNode } from "react";
import { useRef } from "react";
import styles from "./Tabs.module.css";

export type TabsSize = "sm" | "md";

export interface TabItem {
	value: string;
	label: ReactNode;
	badge?: ReactNode;
	disabled?: boolean;
	panelId?: string;
}

export interface TabsProps extends Omit<ComponentPropsWithoutRef<"div">, "onChange"> {
	tabs: TabItem[];
	value: string;
	ariaLabel: string;
	onValueChange: (value: string) => void;
	size?: TabsSize;
}

function nextEnabledTab(tabs: TabItem[], value: string, direction: 1 | -1) {
	const enabledTabs = tabs.filter(tab => !tab.disabled);

	if (!enabledTabs.length) {
		return undefined;
	}

	const currentIndex = Math.max(
		0,
		enabledTabs.findIndex(tab => tab.value === value)
	);
	const nextIndex = (currentIndex + direction + enabledTabs.length) % enabledTabs.length;

	return enabledTabs[nextIndex];
}

export function Tabs({ tabs, value, ariaLabel, onValueChange, size = "md", className, ...props }: TabsProps) {
	const tabRefs = useRef<Record<string, HTMLButtonElement | null>>({});
	const classes = [styles.tabs, styles[size], className].filter(Boolean).join(" ");

	function selectTab(tab: TabItem | undefined) {
		if (!tab || tab.disabled) {
			return;
		}

		onValueChange(tab.value);
		tabRefs.current[tab.value]?.focus();
	}

	function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
		if (event.key === "ArrowRight" || event.key === "ArrowDown") {
			event.preventDefault();
			selectTab(nextEnabledTab(tabs, value, 1));
			return;
		}

		if (event.key === "ArrowLeft" || event.key === "ArrowUp") {
			event.preventDefault();
			selectTab(nextEnabledTab(tabs, value, -1));
			return;
		}

		if (event.key === "Home") {
			event.preventDefault();
			selectTab(tabs.find(tab => !tab.disabled));
			return;
		}

		if (event.key === "End") {
			event.preventDefault();
			selectTab([...tabs].reverse().find(tab => !tab.disabled));
		}
	}

	return (
		<div className={classes} {...props}>
			<div className={styles.list} role="tablist" aria-label={ariaLabel} onKeyDown={handleKeyDown}>
				{tabs.map(tab => {
					const active = tab.value === value;

					return (
						<button
							key={tab.value}
							ref={node => {
								tabRefs.current[tab.value] = node;
							}}
							type="button"
							className={styles.tab}
							role="tab"
							aria-selected={active}
							aria-controls={tab.panelId}
							tabIndex={active ? 0 : -1}
							disabled={tab.disabled}
							data-active={active || undefined}
							onClick={() => selectTab(tab)}
						>
							<span className={styles.label}>{tab.label}</span>
							{tab.badge ? <span className={styles.badge}>{tab.badge}</span> : null}
						</button>
					);
				})}
			</div>
		</div>
	);
}
