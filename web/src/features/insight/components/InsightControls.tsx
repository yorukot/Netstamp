import { displayInsightTimeRange } from "@/features/insight/insightTime";
import type { InsightRefreshInterval, InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import { classNames } from "@/shared/utils/classNames";
import { relativeTimeOptions, relativeTimeRangeDurations } from "@/shared/utils/timeRanges";
import { Button, SelectField, TextField } from "@netstamp/ui";
import { useEffect, useId, useMemo, useRef, useState, type CSSProperties, type KeyboardEvent } from "react";
import { createPortal } from "react-dom";
import styles from "./InsightControls.module.css";

export type SegmentOption<TValue extends string> = {
	value: TValue;
	label: string;
};

const timeOptions: Array<{ value: InsightRelativeRange; label: string }> = relativeTimeOptions;
const timeRangeDurations: Record<InsightRelativeRange, number> = relativeTimeRangeDurations;

const refreshOptions: Array<{ value: InsightRefreshInterval; label: string }> = [
	{ value: "off", label: "Off" },
	{ value: "10s", label: "10s" },
	{ value: "30s", label: "30s" },
	{ value: "1m", label: "1m" },
	{ value: "5m", label: "5m" }
];

function formatDateTimeLocal(value: number) {
	const date = new Date(value);
	const offsetMs = date.getTimezoneOffset() * 60 * 1000;

	return new Date(value - offsetMs).toISOString().slice(0, 16);
}

function parseDateTimeLocal(value: string) {
	const parsed = new Date(value).getTime();

	return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function timePopoverStyle(anchor: HTMLElement | null): CSSProperties | undefined {
	if (!anchor || typeof window === "undefined") {
		return undefined;
	}

	const rect = anchor.getBoundingClientRect();
	const gap = 10;
	const availableWidth = Math.max(280, window.innerWidth - gap * 2);
	const width = Math.min(Math.max(rect.width, 460), availableWidth);
	const left = Math.min(Math.max(gap, rect.left), window.innerWidth - width - gap);
	const top = Math.max(gap, Math.min(rect.bottom + gap, Math.max(gap, window.innerHeight - gap - 220)));
	const maxHeight = Math.max(180, window.innerHeight - top - gap);

	return {
		left,
		maxHeight,
		top,
		width
	};
}

function compactPopoverStyle(anchor: HTMLElement | null): CSSProperties | undefined {
	if (!anchor || typeof window === "undefined") {
		return undefined;
	}

	const rect = anchor.getBoundingClientRect();
	const gap = 8;
	const availableWidth = Math.max(280, window.innerWidth - gap * 2);
	const width = Math.min(Math.max(rect.width, 520), availableWidth);
	const left = Math.min(Math.max(gap, rect.left), window.innerWidth - width - gap);
	const spaceBelow = window.innerHeight - rect.bottom - gap;
	const spaceAbove = rect.top - gap;
	const openAbove = spaceBelow < 260 && spaceAbove > spaceBelow;
	const maxHeight = Math.min(360, Math.max(180, openAbove ? spaceAbove : spaceBelow));
	const top = openAbove ? Math.max(gap, rect.top - gap - maxHeight) : Math.min(rect.bottom + gap, window.innerHeight - gap - maxHeight);

	return { left, maxHeight, top, width };
}

export function SegmentedControl<TValue extends string>({
	label,
	value,
	options,
	onChange
}: {
	label: string;
	value: TValue;
	options: Array<SegmentOption<TValue>>;
	onChange: (value: TValue) => void;
}) {
	return (
		<div className={styles.segmentField}>
			<span className={styles.segmentLabel}>{label}</span>
			<div className={classNames("ns-cut-frame", styles.segmentControl)} role="radiogroup" aria-label={label}>
				{options.map(option => (
					<button
						type="button"
						role="radio"
						aria-checked={value === option.value}
						className={styles.segmentButton}
						data-selected={value === option.value || undefined}
						onClick={() => onChange(option.value)}
						key={option.value}
					>
						{option.label}
					</button>
				))}
			</div>
		</div>
	);
}

export function FocusChip({ label, value, invalid, onClear }: { label: string; value: string; invalid?: boolean; onClear: () => void }) {
	return (
		<div className={classNames("ns-cut-frame", styles.focusChip)} data-invalid={invalid || undefined}>
			<span>{label}</span>
			<strong>{value}</strong>
			<button type="button" onClick={onClear}>
				Clear
			</button>
		</div>
	);
}

export interface AssignmentSelectOption {
	value: string;
	label: string;
	meta: string;
	searchText: string;
}

export function ScopeSelect({
	label,
	placeholder,
	options,
	value,
	disabled,
	onChange
}: {
	label: string;
	placeholder: string;
	options: AssignmentSelectOption[];
	value: string;
	disabled?: boolean;
	onChange: (value: string) => void;
}) {
	const listboxId = useId();
	const inputId = useId();
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const [activeIndex, setActiveIndex] = useState(0);
	const [popoverStyle, setPopoverStyle] = useState<CSSProperties>();
	const rootRef = useRef<HTMLDivElement>(null);
	const popoverRef = useRef<HTMLDivElement>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const optionsByValue = useMemo(() => new Map(options.map(option => [option.value, option])), [options]);
	const selectedOption = value ? optionsByValue.get(value) : undefined;
	const normalizedQuery = query.trim().toLowerCase();
	const filteredOptions = useMemo(() => {
		const matches = normalizedQuery ? options.filter(option => option.searchText.includes(normalizedQuery)) : options;

		return matches.slice(0, 50);
	}, [normalizedQuery, options]);

	function openPopover() {
		if (disabled) {
			return;
		}

		setPopoverStyle(compactPopoverStyle(rootRef.current));
		setOpen(true);
	}

	function commitValue(nextValue: string) {
		onChange(nextValue);
		setQuery("");
		setActiveIndex(0);
		setOpen(false);
		inputRef.current?.focus();
	}

	function clearValue() {
		onChange("");
		setQuery("");
		inputRef.current?.focus();
	}

	function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
		if (event.key === "ArrowDown" || event.key === "ArrowUp") {
			event.preventDefault();
			openPopover();
			setActiveIndex(current => {
				const total = filteredOptions.length;
				if (!total) {
					return 0;
				}

				const offset = event.key === "ArrowDown" ? 1 : -1;
				return (current + offset + total) % total;
			});
			return;
		}

		if (event.key === "Enter") {
			const activeOption = filteredOptions[activeIndex];
			if (!activeOption) {
				return;
			}

			event.preventDefault();
			commitValue(activeOption.value);
			return;
		}

		if (event.key === "Escape") {
			setOpen(false);
			return;
		}

		if (event.key === "Backspace" && !query && value) {
			event.preventDefault();
			clearValue();
		}
	}

	useEffect(() => {
		if (!open) {
			return;
		}

		function updatePosition() {
			setPopoverStyle(compactPopoverStyle(rootRef.current));
		}

		function handlePointerDown(event: PointerEvent) {
			const target = event.target as Node;

			if (rootRef.current?.contains(target) || popoverRef.current?.contains(target)) {
				return;
			}

			setOpen(false);
		}

		window.addEventListener("resize", updatePosition);
		window.addEventListener("scroll", updatePosition, true);
		window.addEventListener("pointerdown", handlePointerDown);

		return () => {
			window.removeEventListener("resize", updatePosition);
			window.removeEventListener("scroll", updatePosition, true);
			window.removeEventListener("pointerdown", handlePointerDown);
		};
	}, [open]);

	const popover =
		open && popoverStyle && typeof document !== "undefined"
			? createPortal(
					<div ref={popoverRef} id={listboxId} className={classNames("ns-cut-frame", "ns-scrollbar", styles.assignmentPopover)} style={popoverStyle} role="listbox">
						{filteredOptions.length ? (
							filteredOptions.map((option, index) => {
								const selected = option.value === value;
								const active = index === activeIndex;

								return (
									<button
										type="button"
										role="option"
										aria-selected={selected}
										className={styles.assignmentOption}
										data-active={active || undefined}
										data-selected={selected || undefined}
										onMouseDown={event => event.preventDefault()}
										onMouseEnter={() => setActiveIndex(index)}
										onClick={() => commitValue(option.value)}
										key={option.value}
									>
										<strong>{option.label}</strong>
										<span>{option.meta}</span>
									</button>
								);
							})
						) : (
							<div className={styles.assignmentEmpty}>No {label.toLowerCase()} match this search.</div>
						)}
					</div>,
					document.body
				)
			: null;

	return (
		<div className={styles.assignmentField}>
			<label className={styles.segmentLabel} htmlFor={inputId}>
				{label}
			</label>
			<div ref={rootRef} className={classNames("ns-cut-frame", styles.assignmentControl)} data-open={open || undefined} data-disabled={disabled || undefined} onClick={() => inputRef.current?.focus()}>
				{selectedOption ? (
					<span className={styles.assignmentToken}>
						<span>{selectedOption.label}</span>
					</span>
				) : null}
				<input
					ref={inputRef}
					id={inputId}
					value={query}
					disabled={disabled}
					placeholder={selectedOption ? "Filter or change selection" : placeholder}
					role="combobox"
					aria-expanded={open}
					aria-controls={open ? listboxId : undefined}
					aria-autocomplete="list"
					autoComplete="off"
					onFocus={openPopover}
					onChange={event => {
						setQuery(event.currentTarget.value);
						setActiveIndex(0);
						openPopover();
					}}
					onKeyDown={handleKeyDown}
				/>
				{value ? (
					<button
						type="button"
						className={styles.assignmentClear}
						aria-label={`Clear ${label}`}
						onMouseDown={event => event.preventDefault()}
						onClick={event => {
							event.stopPropagation();
							clearValue();
						}}
					>
						Clear
					</button>
				) : null}
			</div>
			{popover}
		</div>
	);
}

export function AssignmentMultiSelect({
	label,
	placeholder,
	options,
	selectedValues,
	disabled,
	onChange
}: {
	label: string;
	placeholder: string;
	options: AssignmentSelectOption[];
	selectedValues: string[];
	disabled?: boolean;
	onChange: (values: string[]) => void;
}) {
	const listboxId = useId();
	const inputId = useId();
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const [activeIndex, setActiveIndex] = useState(0);
	const [popoverStyle, setPopoverStyle] = useState<CSSProperties>();
	const rootRef = useRef<HTMLDivElement>(null);
	const popoverRef = useRef<HTMLDivElement>(null);
	const inputRef = useRef<HTMLInputElement>(null);
	const selectedSet = useMemo(() => new Set(selectedValues), [selectedValues]);
	const optionsByValue = useMemo(() => new Map(options.map(option => [option.value, option])), [options]);
	const selectedOptions = selectedValues.map(value => optionsByValue.get(value)).filter((option): option is AssignmentSelectOption => Boolean(option));
	const normalizedQuery = query.trim().toLowerCase();
	const filteredOptions = useMemo(() => {
		const matches = normalizedQuery ? options.filter(option => option.searchText.includes(normalizedQuery)) : options;

		return matches.slice(0, 50);
	}, [normalizedQuery, options]);
	const selectedSummary = selectedOptions.length ? `${selectedOptions.length} assignments selected` : "";

	function openPopover() {
		if (disabled) {
			return;
		}

		setPopoverStyle(compactPopoverStyle(rootRef.current));
		setOpen(true);
	}

	function toggleValue(value: string) {
		if (selectedSet.has(value)) {
			onChange(selectedValues.filter(selectedValue => selectedValue !== value));
			return;
		}

		onChange([...selectedValues, value]);
	}

	function clearValue(value: string) {
		onChange(selectedValues.filter(selectedValue => selectedValue !== value));
	}

	function clearAll() {
		setQuery("");
		onChange([]);
		inputRef.current?.focus();
	}

	function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
		if (event.key === "ArrowDown" || event.key === "ArrowUp") {
			event.preventDefault();
			openPopover();
			setActiveIndex(current => {
				const total = filteredOptions.length;
				if (!total) {
					return 0;
				}

				const offset = event.key === "ArrowDown" ? 1 : -1;
				return (current + offset + total) % total;
			});
			return;
		}

		if (event.key === "Enter") {
			const activeOption = filteredOptions[activeIndex];
			if (!activeOption) {
				return;
			}

			event.preventDefault();
			toggleValue(activeOption.value);
			setQuery("");
			setActiveIndex(0);
			openPopover();
			return;
		}

		if (event.key === "Escape") {
			setOpen(false);
			return;
		}

		if (event.key === "Backspace" && !query && selectedValues.length) {
			event.preventDefault();
			onChange(selectedValues.slice(0, -1));
		}
	}

	useEffect(() => {
		if (!open) {
			return;
		}

		function updatePosition() {
			setPopoverStyle(compactPopoverStyle(rootRef.current));
		}

		function handlePointerDown(event: PointerEvent) {
			const target = event.target as Node;

			if (rootRef.current?.contains(target) || popoverRef.current?.contains(target)) {
				return;
			}

			setOpen(false);
		}

		window.addEventListener("resize", updatePosition);
		window.addEventListener("scroll", updatePosition, true);
		window.addEventListener("pointerdown", handlePointerDown);

		return () => {
			window.removeEventListener("resize", updatePosition);
			window.removeEventListener("scroll", updatePosition, true);
			window.removeEventListener("pointerdown", handlePointerDown);
		};
	}, [open]);

	const popover =
		open && popoverStyle && typeof document !== "undefined"
			? createPortal(
					<div ref={popoverRef} id={listboxId} className={classNames("ns-cut-frame", "ns-scrollbar", styles.assignmentPopover)} style={popoverStyle} role="listbox" aria-multiselectable="true">
						{filteredOptions.length ? (
							filteredOptions.map((option, index) => {
								const selected = selectedSet.has(option.value);
								const active = index === activeIndex;

								return (
									<button
										type="button"
										role="option"
										aria-selected={selected}
										className={styles.assignmentOption}
										data-active={active || undefined}
										data-selected={selected || undefined}
										onMouseDown={event => event.preventDefault()}
										onMouseEnter={() => setActiveIndex(index)}
										onClick={() => {
											toggleValue(option.value);
											inputRef.current?.focus();
										}}
										key={option.value}
									>
										<strong>{option.label}</strong>
										<span>{option.meta}</span>
									</button>
								);
							})
						) : (
							<div className={styles.assignmentEmpty}>No assignments match this search.</div>
						)}
					</div>,
					document.body
				)
			: null;

	return (
		<div className={styles.assignmentField}>
			<label className={styles.segmentLabel} htmlFor={inputId}>
				{label}
			</label>
			<div ref={rootRef} className={classNames("ns-cut-frame", styles.assignmentControl)} data-open={open || undefined} data-disabled={disabled || undefined} onClick={() => inputRef.current?.focus()}>
				{selectedOptions.slice(0, 2).map(option => (
					<span className={styles.assignmentToken} key={option.value}>
						<span>{option.label}</span>
						<button
							type="button"
							aria-label={`Remove ${option.label}`}
							onMouseDown={event => event.preventDefault()}
							onClick={event => {
								event.stopPropagation();
								clearValue(option.value);
							}}
						>
							x
						</button>
					</span>
				))}
				{selectedOptions.length > 2 ? <span className={styles.assignmentOverflow}>+{selectedOptions.length - 2}</span> : null}
				<input
					ref={inputRef}
					id={inputId}
					value={query}
					disabled={disabled}
					placeholder={selectedOptions.length ? selectedSummary : placeholder}
					role="combobox"
					aria-expanded={open}
					aria-controls={open ? listboxId : undefined}
					aria-autocomplete="list"
					autoComplete="off"
					onFocus={openPopover}
					onChange={event => {
						setQuery(event.currentTarget.value);
						setActiveIndex(0);
						openPopover();
					}}
					onKeyDown={handleKeyDown}
				/>
				{selectedValues.length ? (
					<button
						type="button"
						className={styles.assignmentClear}
						aria-label="Clear assignment selection"
						onMouseDown={event => event.preventDefault()}
						onClick={event => {
							event.stopPropagation();
							clearAll();
						}}
					>
						Clear
					</button>
				) : null}
			</div>
			{popover}
		</div>
	);
}

export function InsightTimeControl({
	timeMode,
	timeRange,
	timeWindow,
	refresh,
	className,
	onApplyRelative,
	onApplyAbsolute,
	onRefresh,
	onRefreshChange
}: {
	timeMode: InsightTimeMode;
	timeRange: InsightRelativeRange;
	timeWindow: TimeWindow;
	refresh: InsightRefreshInterval;
	className?: string;
	onApplyRelative: (range: InsightRelativeRange) => void;
	onApplyAbsolute: (timeWindow: TimeWindow) => void;
	onRefresh: () => void;
	onRefreshChange: (refresh: InsightRefreshInterval) => void;
}) {
	const [open, setOpen] = useState(false);
	const [popoverStyle, setPopoverStyle] = useState<CSSProperties>();
	const rootRef = useRef<HTMLDivElement>(null);
	const popoverRef = useRef<HTMLDivElement>(null);
	const timeWindowKey = `${timeWindow.from}:${timeWindow.to}`;
	const initialAbsoluteDraft = {
		key: timeWindowKey,
		from: formatDateTimeLocal(timeWindow.from),
		to: formatDateTimeLocal(timeWindow.to)
	};
	const [absoluteDraft, setAbsoluteDraft] = useState(initialAbsoluteDraft);
	const activeAbsoluteDraft = absoluteDraft.key === timeWindowKey ? absoluteDraft : initialAbsoluteDraft;
	const absoluteFrom = activeAbsoluteDraft.from;
	const absoluteTo = activeAbsoluteDraft.to;
	const timeButtonId = useId();
	const absoluteFromMs = parseDateTimeLocal(absoluteFrom);
	const absoluteToMs = parseDateTimeLocal(absoluteTo);
	const canApplyAbsolute = absoluteFromMs !== null && absoluteToMs !== null && absoluteFromMs < absoluteToMs;

	function applyAbsolute() {
		if (absoluteFromMs === null || absoluteToMs === null || absoluteFromMs >= absoluteToMs) {
			return;
		}

		onApplyAbsolute({ from: absoluteFromMs, to: absoluteToMs });
		setOpen(false);
	}

	function applyNow() {
		if (timeMode === "relative") {
			onApplyRelative(timeRange);
			setOpen(false);
			return;
		}

		const duration = Math.max(timeWindow.to - timeWindow.from, timeRangeDurations["15m"]);
		const to = Date.now();
		onApplyAbsolute({ from: to - duration, to });
		setOpen(false);
	}

	function togglePopover() {
		if (open) {
			setOpen(false);
			return;
		}

		setPopoverStyle(timePopoverStyle(rootRef.current));
		setOpen(true);
	}

	useEffect(() => {
		if (!open) {
			return;
		}

		function updatePosition() {
			setPopoverStyle(timePopoverStyle(rootRef.current));
		}

		function handlePointerDown(event: PointerEvent) {
			const target = event.target as Node;

			if (rootRef.current?.contains(target) || popoverRef.current?.contains(target)) {
				return;
			}

			setOpen(false);
		}

		window.addEventListener("resize", updatePosition);
		window.addEventListener("scroll", updatePosition, true);
		window.addEventListener("pointerdown", handlePointerDown);

		return () => {
			window.removeEventListener("resize", updatePosition);
			window.removeEventListener("scroll", updatePosition, true);
			window.removeEventListener("pointerdown", handlePointerDown);
		};
	}, [open]);

	const timePopover =
		open && popoverStyle && typeof document !== "undefined"
			? createPortal(
					<div
						ref={popoverRef}
						id={`${timeButtonId}-panel`}
						className={classNames("ns-cut-frame", "ns-scrollbar", styles.timePopover)}
						style={popoverStyle}
						role="dialog"
						aria-labelledby={timeButtonId}
					>
						<section className={styles.timeSection}>
							<h4>Relative time</h4>
							<div className={styles.timePresetGrid}>
								{timeOptions.map(option => (
									<button
										type="button"
										className={classNames("ns-cut-frame", styles.timePreset)}
										data-selected={timeMode === "relative" && timeRange === option.value}
										onClick={() => {
											onApplyRelative(option.value);
											setOpen(false);
										}}
										key={option.value}
									>
										{option.label}
									</button>
								))}
							</div>
						</section>
						<section className={styles.timeSection}>
							<h4>Absolute time</h4>
							<div className={styles.absoluteGrid}>
								<TextField label="From" type="datetime-local" value={absoluteFrom} onChange={event => setAbsoluteDraft({ ...activeAbsoluteDraft, from: event.currentTarget.value })} />
								<TextField label="To" type="datetime-local" value={absoluteTo} onChange={event => setAbsoluteDraft({ ...activeAbsoluteDraft, to: event.currentTarget.value })} />
							</div>
							<div className={styles.timeActions}>
								<Button type="button" variant="outline" size="sm" onClick={applyNow}>
									Now
								</Button>
								<Button type="button" variant="secondary" size="sm" disabled={!canApplyAbsolute} onClick={applyAbsolute}>
									Apply time range
								</Button>
							</div>
						</section>
					</div>,
					document.body
				)
			: null;

	return (
		<div ref={rootRef} className={classNames(styles.timeControlRoot, className)}>
			<span className={styles.segmentLabel}>Time</span>
			<div className={styles.timeControls}>
				<button id={timeButtonId} type="button" className={classNames("ns-cut-frame", styles.timeTrigger)} aria-expanded={open} aria-controls={`${timeButtonId}-panel`} onClick={togglePopover}>
					<span>{displayInsightTimeRange(timeMode, timeRange, timeWindow)}</span>
				</button>
				<Button type="button" variant="outline" size="sm" className={styles.refreshButton} onClick={onRefresh}>
					Refresh
				</Button>
				<SelectField label="Refresh" value={refresh} options={refreshOptions} onChange={event => onRefreshChange(event.currentTarget.value as InsightRefreshInterval)} />
			</div>
			{timePopover}
		</div>
	);
}
