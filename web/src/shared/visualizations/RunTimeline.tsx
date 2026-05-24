import type { CSSProperties, ReactNode } from "react";
import styles from "./RunTimeline.module.css";

export interface RunTimelinePoint {
	id: string;
	timestampMs: number;
	label: string;
	value: number | null;
	valueLabel: string;
	ariaLabel: string;
	hasLoss?: boolean;
	hasChange?: boolean;
}

type TimelinePointStyle = CSSProperties & {
	"--ns-timeline-x"?: string;
	"--ns-timeline-y"?: string;
};

interface RunTimelineProps {
	points: RunTimelinePoint[];
	selectedPointId?: string;
	selectedValueLabel?: string;
	emptyState?: ReactNode;
	onSelectPoint: (id: string) => void;
}

export function RunTimeline({ points, selectedPointId, selectedValueLabel, emptyState, onSelectPoint }: RunTimelineProps) {
	const sortedPoints = [...points].sort((a, b) => a.timestampMs - b.timestampMs);

	if (!sortedPoints.length) {
		return emptyState;
	}

	const firstPoint = sortedPoints[0];
	const lastPoint = sortedPoints[sortedPoints.length - 1];
	const firstTime = firstPoint.timestampMs;
	const lastTime = lastPoint.timestampMs;
	const timeSpan = Math.max(1, lastTime - firstTime);
	const values = sortedPoints.map(point => point.value).filter((value): value is number => typeof value === "number");
	const minValue = values.length ? Math.min(...values) : 0;
	const maxValue = values.length ? Math.max(...values) : 1;
	const valueSpan = Math.max(1, maxValue - minValue);
	const selectedPoint = sortedPoints.find(point => point.id === selectedPointId);
	const axisSelectedValue = selectedValueLabel || `${selectedPoint?.valueLabel || lastPoint.valueLabel} selected`;
	const viewPoints = sortedPoints.map(point => {
		const x = 6 + ((point.timestampMs - firstTime) / timeSpan) * 88;
		const y = typeof point.value === "number" ? (maxValue === minValue ? 49 : 78 - ((point.value - minValue) / valueSpan) * 58) : 82;

		return { ...point, x, y };
	});
	const polylinePoints = viewPoints.map(point => `${point.x},${point.y}`).join(" ");

	return (
		<div className={styles.timeline}>
			<div className={styles.timelineChart}>
				<svg className={styles.timelineSvg} viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="20" y2="20" />
					<line className={styles.timelineGridLine} x1="6" x2="94" y1="49" y2="49" />
					<line className={styles.timelineAxisLine} x1="6" x2="94" y1="78" y2="78" />
					{polylinePoints ? <polyline className={styles.timelineLine} points={polylinePoints} /> : null}
				</svg>
				{viewPoints.map(point => {
					const style: TimelinePointStyle = {
						"--ns-timeline-x": `${point.x}%`,
						"--ns-timeline-y": `${point.y}%`
					};
					const selected = selectedPointId === point.id;

					return (
						<button
							type="button"
							className={styles.timelinePoint}
							style={style}
							data-selected={selected || undefined}
							data-loss={point.hasLoss || undefined}
							data-changed={point.hasChange || undefined}
							onClick={() => onSelectPoint(point.id)}
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
				<span>RTT line</span>
				<span>Loss</span>
				<span>Route change</span>
				<span>Selected run</span>
			</div>
		</div>
	);
}
