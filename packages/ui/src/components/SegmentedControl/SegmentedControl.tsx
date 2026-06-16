import type { ComponentPropsWithoutRef, KeyboardEvent, ReactNode } from "react";
import { useRef } from "react";
import styles from "./SegmentedControl.module.css";

export type SegmentedControlSize = "sm" | "md";

export interface SegmentedControlOption {
	value: string;
	label: ReactNode;
	disabled?: boolean;
}

export interface SegmentedControlProps extends Omit<ComponentPropsWithoutRef<"div">, "onChange"> {
	options: SegmentedControlOption[];
	value: string;
	ariaLabel: string;
	onValueChange: (value: string) => void;
	size?: SegmentedControlSize;
}

function enabledOptions(options: SegmentedControlOption[]) {
	return options.filter(option => !option.disabled);
}

function nextEnabledOption(options: SegmentedControlOption[], value: string, direction: 1 | -1) {
	const enabled = enabledOptions(options);

	if (!enabled.length) {
		return undefined;
	}

	const currentIndex = Math.max(
		0,
		enabled.findIndex(option => option.value === value)
	);
	const nextIndex = (currentIndex + direction + enabled.length) % enabled.length;

	return enabled[nextIndex];
}

export function SegmentedControl({ options, value, ariaLabel, onValueChange, size = "md", className, ...props }: SegmentedControlProps) {
	const optionRefs = useRef<Record<string, HTMLButtonElement | null>>({});
	const classes = [styles.group, styles[size], className].filter(Boolean).join(" ");

	function selectOption(option: SegmentedControlOption | undefined) {
		if (!option || option.disabled) {
			return;
		}

		onValueChange(option.value);
		optionRefs.current[option.value]?.focus();
	}

	function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
		if (event.key === "ArrowRight" || event.key === "ArrowDown") {
			event.preventDefault();
			selectOption(nextEnabledOption(options, value, 1));
			return;
		}

		if (event.key === "ArrowLeft" || event.key === "ArrowUp") {
			event.preventDefault();
			selectOption(nextEnabledOption(options, value, -1));
			return;
		}

		if (event.key === "Home") {
			event.preventDefault();
			selectOption(enabledOptions(options)[0]);
			return;
		}

		if (event.key === "End") {
			event.preventDefault();
			const enabled = enabledOptions(options);
			selectOption(enabled[enabled.length - 1]);
		}
	}

	return (
		<div className={classes} role="radiogroup" aria-label={ariaLabel} onKeyDown={handleKeyDown} {...props}>
			{options.map(option => {
				const active = option.value === value;

				return (
					<button
						key={option.value}
						ref={node => {
							optionRefs.current[option.value] = node;
						}}
						type="button"
						className={styles.option}
						role="radio"
						aria-checked={active}
						tabIndex={active ? 0 : -1}
						disabled={option.disabled}
						data-active={active || undefined}
						onClick={() => selectOption(option)}
					>
						<span>{option.label}</span>
					</button>
				);
			})}
		</div>
	);
}
