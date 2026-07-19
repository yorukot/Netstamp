import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties, type PointerEvent, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import styles from "./RunTimeline.module.css";

export interface RunTimelinePoint {
	id: string;
	timestampMs: number;
	rangeFromMs?: number;
	rangeToMs?: number;
	runStartedAt?: string;
	label: string;
	value: number | null;
	valueLabel: string;
	ariaLabel: string;
	tone?: "normal" | "warning" | "critical";
	hasLoss?: boolean;
	hasChange?: boolean;
}

type TimelinePointStyle = CSSProperties & {
	"--ns-timeline-x"?: string;
	"--ns-timeline-y"?: string;
};

type TimelineSelectionStyle = CSSProperties & {
	"--ns-selection-left"?: string;
	"--ns-selection-width"?: string;
};

interface RunTimelineProps {
	points: RunTimelinePoint[];
	selectedPointId?: string;
	selectedValueLabel?: string;
	emptyState?: ReactNode;
	timeRangeBounds?: { from: number; to: number };
	minTimeRangeMs?: number;
	onSelectPoint: (point: RunTimelinePoint) => void;
	onSelectTimeRange?: (range: { from: number; to: number }) => void;
}

function clamp(value: number, min: number, max: number) {
	return Math.min(max, Math.max(min, value));
}

export function RunTimeline({ points, selectedPointId, selectedValueLabel, emptyState, timeRangeBounds, minTimeRangeMs = 1000, onSelectPoint, onSelectTimeRange }: RunTimelineProps) {
	const { t } = useTranslation("insight");
	const chartRef = useRef<HTMLDivElement | null>(null);
	const [selection, setSelection] = useState<{ anchorMs: number; focusMs: number } | null>(null);
	const selectionRef = useRef(selection);
	const sortedPoints = useMemo(() => [...points].sort((a, b) => a.timestampMs - b.timestampMs), [points]);
	const fallbackPoint: RunTimelinePoint = { id: "empty", timestampMs: timeRangeBounds?.from ?? 0, label: "-", value: null, valueLabel: "-", ariaLabel: t("legend.noTimelinePoints") };
	const firstPoint = sortedPoints[0] ?? fallbackPoint;
	const lastPoint = sortedPoints[sortedPoints.length - 1] ?? fallbackPoint;
	const firstTime = timeRangeBounds?.from ?? firstPoint.timestampMs;
	const lastTime = timeRangeBounds?.to ?? lastPoint.timestampMs;
	const timeSpan = Math.max(1, lastTime - firstTime);
	const values = sortedPoints.map(point => point.value).filter((value): value is number => typeof value === "number");
	const minValue = values.length ? Math.min(...values) : 0;
	const maxValue = values.length ? Math.max(...values) : 1;
	const valueSpan = Math.max(1, maxValue - minValue);
	const selectedPoint = sortedPoints.find(point => point.id === selectedPointId);
	const axisSelectedValue = selectedValueLabel || t("legend.selected", { value: selectedPoint?.valueLabel || lastPoint.valueLabel });
	const viewPoints = sortedPoints.map(point => {
		const x = 6 + ((point.timestampMs - firstTime) / timeSpan) * 88;
		const y = typeof point.value === "number" ? (maxValue === minValue ? 49 : 78 - ((point.value - minValue) / valueSpan) * 58) : 82;

		return { ...point, x, y };
	});
	const polylinePoints = viewPoints.map(point => `${point.x},${point.y}`).join(" ");
	const selectionRange = selection
		? {
				from: Math.min(selection.anchorMs, selection.focusMs),
				to: Math.max(selection.anchorMs, selection.focusMs)
			}
		: null;
	const selectionStyle: TimelineSelectionStyle | undefined = selectionRange
		? {
				"--ns-selection-left": `${6 + ((selectionRange.from - firstTime) / timeSpan) * 88}%`,
				"--ns-selection-width": `${((selectionRange.to - selectionRange.from) / timeSpan) * 88}%`
			}
		: undefined;

	const timeFromPointer = useCallback(
		(event: PointerEvent<HTMLDivElement> | globalThis.PointerEvent) => {
			if (!chartRef.current) {
				return firstTime;
			}

			const rect = chartRef.current.getBoundingClientRect();
			const percent = clamp(((event.clientX - rect.left) / Math.max(1, rect.width)) * 100, 6, 94);
			return Math.trunc(firstTime + ((percent - 6) / 88) * timeSpan);
		},
		[firstTime, timeSpan]
	);

	function beginSelection(event: PointerEvent<HTMLDivElement>) {
		if (!onSelectTimeRange || (event.target as HTMLElement).closest("button")) {
			return;
		}

		const anchorMs = timeFromPointer(event);
		event.currentTarget.setPointerCapture(event.pointerId);
		const nextSelection = { anchorMs, focusMs: anchorMs };
		selectionRef.current = nextSelection;
		setSelection(nextSelection);
	}

	useEffect(() => {
		selectionRef.current = selection;
	}, [selection]);

	useEffect(() => {
		if (!selection || !onSelectTimeRange) {
			return undefined;
		}
		const selectTimeRange = onSelectTimeRange;

		function handlePointerMove(event: globalThis.PointerEvent) {
			setSelection(current => {
				const next = current ? { ...current, focusMs: timeFromPointer(event) } : current;
				selectionRef.current = next;
				return next;
			});
		}

		function handlePointerUp() {
			const activeSelection = selectionRef.current;

			if (!activeSelection) {
				return;
			}

			const from = Math.trunc(Math.min(activeSelection.anchorMs, activeSelection.focusMs));
			const to = Math.trunc(Math.max(activeSelection.anchorMs, activeSelection.focusMs));

			if (to - from >= minTimeRangeMs) {
				selectTimeRange({ from, to });
			}

			setSelection(null);
		}

		window.addEventListener("pointermove", handlePointerMove);
		window.addEventListener("pointerup", handlePointerUp, { once: true });

		return () => {
			window.removeEventListener("pointermove", handlePointerMove);
			window.removeEventListener("pointerup", handlePointerUp);
		};
	}, [minTimeRangeMs, onSelectTimeRange, selection, timeFromPointer]);

	if (!sortedPoints.length) {
		return emptyState;
	}

	return (
		<div className={styles.timeline}>
			<div className={styles.timelineChart} ref={chartRef} onPointerDown={beginSelection} data-selectable={Boolean(onSelectTimeRange) || undefined}>
				<svg className={styles.timelineSvg} viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="20" y2="20" />
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="49" y2="49" />
					<line className={styles.timelineAxisLine} x1="6" x2="94" y1="78" y2="78" />
					{polylinePoints ? <polyline className={styles.timelineLine} points={polylinePoints} /> : null}
				</svg>
				{selectionRange ? <div className={styles.timelineSelection} style={selectionStyle} aria-hidden="true" /> : null}
				{viewPoints.map(point => {
					const style: TimelinePointStyle = {
						"--ns-timeline-x": `${point.x}%`,
						"--ns-timeline-y": `${point.y}%`
					};
					const selected = selectedPointId === point.id;
					const tone = point.tone || (point.hasLoss ? "critical" : "normal");

					return (
						<button
							type="button"
							className={styles.timelinePoint}
							style={style}
							data-selected={selected || undefined}
							data-tone={tone}
							data-loss={point.hasLoss || undefined}
							data-changed={point.hasChange || undefined}
							onClick={() => onSelectPoint(point)}
							aria-label={point.ariaLabel}
							key={point.id}
						>
							<span className={styles.timelinePointCore} />
							<span className={styles.timelinePointLabel}>
								{point.label}
								<br />
								{point.valueLabel}
							</span>
						</button>
					);
				})}
				<div className={styles.timelineAxisLabels}>
					<span>{firstPoint.label}</span>
					<strong>{axisSelectedValue}</strong>
					<span>{lastPoint.label}</span>
				</div>
			</div>
			<div className={styles.timelineLegend}>
				<span data-tone="normal">{t("legend.normal")}</span>
				<span data-tone="warning">{t("legend.highRtt")}</span>
				<span data-tone="critical">{t("legend.loss")}</span>
				<span data-tone="changed">{t("legend.routeChange")}</span>
				<span data-tone="selected">{t("legend.selectedRun")}</span>
			</div>
		</div>
	);
}
