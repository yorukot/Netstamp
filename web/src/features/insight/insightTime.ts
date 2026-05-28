import type { InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import { formatTimeWindow, relativeTimeOptions } from "@/shared/utils/timeRanges";

const timeOptions: Array<{ value: InsightRelativeRange; label: string }> = relativeTimeOptions;

function timeLabel(value: InsightRelativeRange) {
	return timeOptions.find(option => option.value === value)?.label || value;
}

export function displayInsightTimeRange(timeMode: InsightTimeMode, timeRange: InsightRelativeRange, timeWindow: TimeWindow) {
	if (timeMode === "relative") {
		return timeLabel(timeRange);
	}

	return formatTimeWindow(timeWindow.from, timeWindow.to, " -> ");
}
