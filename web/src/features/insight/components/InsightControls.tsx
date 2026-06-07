import { displayInsightTimeRange } from "@/features/insight/insightTime";
import type { InsightRefreshInterval, InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import { classNames } from "@/shared/utils/classNames";
import { relativeTimeOptions, relativeTimeRangeDurations } from "@/shared/utils/timeRanges";
import {
	Button,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuPortal,
	DropdownMenuRoot,
	DropdownMenuTrigger,
	PopoverAnchor,
	PopoverContent,
	PopoverPortal,
	PopoverRoot,
	PopoverTrigger,
	SelectField,
	TextField
} from "@netstamp/ui";
import { CaretDown, X } from "@phosphor-icons/react";
import { useId, useMemo, useRef, useState, type KeyboardEvent, type PointerEvent as ReactPointerEvent } from "react";
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
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const optionsByValue = useMemo(() => new Map(options.map(option => [option.value, option])), [options]);
	const selectedOption = value ? optionsByValue.get(value) : undefined;
	const normalizedQuery = query.trim().toLowerCase();
	const filteredOptions = useMemo(() => {
		const matches = normalizedQuery ? options.filter(option => option.searchText.includes(normalizedQuery)) : options;

		return matches.slice(0, 50);
	}, [normalizedQuery, options]);
	const summary = selectedOption ? `${selectedOption.label} · ${selectedOption.meta}` : placeholder;

	function commitValue(nextValue: string) {
		onChange(nextValue);
		setOpen(false);
		setQuery("");
	}

	return (
		<DropdownMenuRoot
			open={open}
			onOpenChange={nextOpen => {
				setOpen(nextOpen);
				if (!nextOpen) {
					setQuery("");
				}
			}}
		>
			<div className={styles.assignmentField}>
				<span className={styles.segmentLabel}>{label}</span>
				<DropdownMenuTrigger asChild disabled={disabled}>
					<button type="button" className={classNames("ns-cut-frame", styles.assignmentTrigger)} disabled={disabled}>
						<span>{summary}</span>
						<CaretDown className={styles.controlIcon} size={18} weight="bold" aria-hidden="true" focusable="false" />
					</button>
				</DropdownMenuTrigger>
				<DropdownMenuPortal>
					<DropdownMenuContent className={classNames("ns-cut-frame", "ns-scrollbar", styles.assignmentPopover)} align="start" sideOffset={8} collisionPadding={8}>
						<div className={styles.assignmentSearch} onKeyDown={event => event.stopPropagation()}>
							<input value={query} placeholder={`Search ${label.toLowerCase()}`} autoComplete="off" autoFocus onChange={event => setQuery(event.currentTarget.value)} />
						</div>
						{value ? (
							<DropdownMenuItem className={classNames(styles.assignmentOption, styles.assignmentActionOption)} onSelect={() => commitValue("")}>
								<strong>Clear selection</strong>
								<span>{selectedOption?.label ?? "Selected scope"}</span>
							</DropdownMenuItem>
						) : null}
						{filteredOptions.length ? (
							filteredOptions.map(option => (
								<DropdownMenuItem className={styles.assignmentOption} data-state={option.value === value ? "checked" : undefined} onSelect={() => commitValue(option.value)} key={option.value}>
									<strong>{option.label}</strong>
									<span>{option.meta}</span>
								</DropdownMenuItem>
							))
						) : (
							<div className={styles.assignmentEmpty}>No {label.toLowerCase()} match this search.</div>
						)}
					</DropdownMenuContent>
				</DropdownMenuPortal>
			</div>
		</DropdownMenuRoot>
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
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const listboxId = useId();
	const inputId = useId();
	const [activeIndex, setActiveIndex] = useState(0);
	const inputRef = useRef<HTMLInputElement>(null);
	const selectedSet = useMemo(() => new Set(selectedValues), [selectedValues]);
	const optionsByValue = useMemo(() => new Map(options.map(option => [option.value, option])), [options]);
	const selectedOptions = selectedValues.map(value => optionsByValue.get(value)).filter((option): option is AssignmentSelectOption => Boolean(option));
	const normalizedQuery = query.trim().toLowerCase();
	const filteredOptions = useMemo(() => {
		const matches = normalizedQuery ? options.filter(option => option.searchText.includes(normalizedQuery)) : options;

		return matches.slice(0, 50);
	}, [normalizedQuery, options]);
	const selectedSummary = selectedOptions.length
		? selectedOptions.length > 2
			? `${selectedOptions[0]?.label}, ${selectedOptions[1]?.label}, +${selectedOptions.length - 2}`
			: selectedOptions.map(option => option.label).join(", ")
		: placeholder;

	function openPopover() {
		if (disabled) {
			return;
		}

		setOpen(true);
	}

	function focusInputSoon() {
		if (typeof window === "undefined") {
			inputRef.current?.focus();
			return;
		}

		window.requestAnimationFrame(() => inputRef.current?.focus());
	}

	function handleControlPointerDown(event: ReactPointerEvent<HTMLDivElement>) {
		if (event.button !== 0 || disabled || event.target instanceof HTMLButtonElement) {
			return;
		}

		openPopover();
		focusInputSoon();
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

	return (
		<PopoverRoot
			open={open}
			onOpenChange={nextOpen => {
				setOpen(disabled ? false : nextOpen);
			}}
		>
			<div className={styles.assignmentField}>
				<label className={styles.segmentLabel} htmlFor={inputId}>
					{label}
				</label>
				<PopoverAnchor asChild>
					<div
						className={classNames("ns-cut-frame", styles.assignmentControl)}
						data-open={open || undefined}
						data-disabled={disabled || undefined}
						onPointerDownCapture={handleControlPointerDown}
						onClick={() => inputRef.current?.focus()}
					>
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
									<X className={styles.tokenIcon} size={12} weight="bold" aria-hidden="true" focusable="false" />
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
				</PopoverAnchor>
				<PopoverPortal>
					<PopoverContent
						id={listboxId}
						className={classNames("ns-cut-frame", "ns-scrollbar", styles.assignmentPopover)}
						role="listbox"
						aria-multiselectable="true"
						align="start"
						side="bottom"
						sideOffset={8}
						avoidCollisions={false}
						collisionPadding={8}
						onOpenAutoFocus={event => event.preventDefault()}
						onCloseAutoFocus={event => event.preventDefault()}
					>
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
					</PopoverContent>
				</PopoverPortal>
			</div>
		</PopoverRoot>
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

	return (
		<PopoverRoot open={open} onOpenChange={setOpen}>
			<div className={classNames(styles.timeControlRoot, className)}>
				<span className={styles.segmentLabel}>Time</span>
				<div className={styles.timeControls}>
					<PopoverTrigger asChild>
						<button id={timeButtonId} type="button" className={classNames("ns-cut-frame", styles.timeTrigger)} aria-controls={`${timeButtonId}-panel`}>
							<span>{displayInsightTimeRange(timeMode, timeRange, timeWindow)}</span>
							<CaretDown className={styles.controlIcon} size={18} weight="bold" aria-hidden="true" focusable="false" />
						</button>
					</PopoverTrigger>
					<Button type="button" variant="outline" size="sm" className={styles.refreshButton} onClick={onRefresh}>
						Refresh
					</Button>
					<SelectField label="Refresh" value={refresh} options={refreshOptions} onChange={event => onRefreshChange(event.currentTarget.value as InsightRefreshInterval)} />
				</div>
				<PopoverPortal>
					<PopoverContent
						id={`${timeButtonId}-panel`}
						className={classNames("ns-cut-frame", "ns-scrollbar", styles.timePopover)}
						role="dialog"
						aria-labelledby={timeButtonId}
						align="start"
						sideOffset={10}
						collisionPadding={10}
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
					</PopoverContent>
				</PopoverPortal>
			</div>
		</PopoverRoot>
	);
}
