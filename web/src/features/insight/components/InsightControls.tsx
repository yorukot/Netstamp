import { displayInsightTimeRange } from "@/features/insight/insightTime";
import type { InsightRefreshInterval, InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import { classNames } from "@/shared/utils/classNames";
import { relativeTimeOptions, relativeTimeRangeDurations } from "@/shared/utils/timeRanges";
import { Button, Input, Select } from "@netstamp/ui";
import { useEffect, useId, useRef, useState, type CSSProperties } from "react";
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
			<div className={styles.segmentControl} role="radiogroup" aria-label={label}>
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
		<div className={styles.focusChip} data-invalid={invalid || undefined}>
			<span>{label}</span>
			<strong>{value}</strong>
			<button type="button" onClick={onClear}>
				Clear
			</button>
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
					<div ref={popoverRef} id={`${timeButtonId}-panel`} className={styles.timePopover} style={popoverStyle} role="dialog" aria-labelledby={timeButtonId}>
						<section className={styles.timeSection}>
							<h4>Relative time</h4>
							<div className={styles.timePresetGrid}>
								{timeOptions.map(option => (
									<button
										type="button"
										className={styles.timePreset}
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
								<label>
									<span>From</span>
									<Input variant="compact" type="datetime-local" value={absoluteFrom} onChange={event => setAbsoluteDraft({ ...activeAbsoluteDraft, from: event.currentTarget.value })} />
								</label>
								<label>
									<span>To</span>
									<Input variant="compact" type="datetime-local" value={absoluteTo} onChange={event => setAbsoluteDraft({ ...activeAbsoluteDraft, to: event.currentTarget.value })} />
								</label>
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
				<button id={timeButtonId} type="button" className={styles.timeTrigger} aria-expanded={open} aria-controls={`${timeButtonId}-panel`} onClick={togglePopover}>
					<span>{displayInsightTimeRange(timeMode, timeRange, timeWindow)}</span>
				</button>
				<Button type="button" variant="outline" size="sm" className={styles.refreshButton} onClick={onRefresh}>
					Refresh
				</Button>
				<label className={styles.refreshField}>
					<span>Refresh</span>
					<Select
						variant="compact"
						value={refresh}
						className={styles.refreshSelect}
						aria-label="Refresh interval"
						onChange={event => onRefreshChange(event.currentTarget.value as InsightRefreshInterval)}
					>
						{refreshOptions.map(option => (
							<option value={option.value} key={option.value}>
								{option.label}
							</option>
						))}
					</Select>
				</label>
			</div>
			{timePopover}
		</div>
	);
}
