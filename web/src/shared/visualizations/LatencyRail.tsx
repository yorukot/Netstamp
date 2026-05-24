import { classNames } from "@/shared/utils/classNames";
import type { CSSProperties } from "react";
import styles from "./LatencyRail.module.css";

export type LatencyRailTone = "success" | "warning" | "critical" | "muted";

type RailStyle = CSSProperties & {
	"--ns-hop-range-start"?: string;
	"--ns-hop-range-end"?: string;
	"--ns-hop-rtt"?: string;
};

interface LatencyRailProps {
	minValue?: number | null;
	avgValue?: number | null;
	maxValue?: number | null;
	scaleMax: number;
	valueLabel: string;
	tone?: LatencyRailTone;
	className?: string;
}

function clampPercent(value: number) {
	return `${Math.max(0, Math.min(100, value))}%`;
}

function hasValue(value: number | null | undefined): value is number {
	return typeof value === "number" && Number.isFinite(value);
}

export function LatencyRail({ minValue, avgValue, maxValue, scaleMax, valueLabel, tone = "muted", className }: LatencyRailProps) {
	const scale = Math.max(1, scaleMax);
	const start = ((minValue ?? avgValue ?? 0) / scale) * 100;
	const end = ((maxValue ?? avgValue ?? 0) / scale) * 100;
	const avg = ((avgValue ?? 0) / scale) * 100;
	const style: RailStyle = {
		"--ns-hop-range-start": clampPercent(start),
		"--ns-hop-range-end": clampPercent(end),
		"--ns-hop-rtt": clampPercent(avg)
	};

	return (
		<span className={classNames(styles.railCell, styles[`railCell${tone}`], className)} style={style}>
			<span className={styles.railTrack}>
				{hasValue(avgValue) ? <span className={styles.railRange} /> : null}
				{hasValue(avgValue) ? <span className={styles.railPoint} /> : null}
			</span>
			<span className={styles.railValue}>{valueLabel}</span>
		</span>
	);
}
