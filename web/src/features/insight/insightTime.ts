import type { InsightRelativeRange, InsightTimeMode, TimeWindow } from "@/features/insight/insightTypes";
import { i18n } from "@/i18n";
import { formatTimeWindow } from "@/shared/utils/timeRanges";

const insightT = i18n.getFixedT(null, "insight") as (key: string) => string;
const timeKeys: Record<InsightRelativeRange, string> = {
	"15m": "controls.last15m",
	"1h": "controls.last1h",
	"6h": "controls.last6h",
	"24h": "controls.last24h",
	"7d": "controls.last7d",
	"30d": "controls.last30d"
};

function timeLabel(value: InsightRelativeRange) {
	return insightT(timeKeys[value]) || value;
}

export function displayInsightTimeRange(timeMode: InsightTimeMode, timeRange: InsightRelativeRange, timeWindow: TimeWindow) {
	if (timeMode === "relative") {
		return timeLabel(timeRange);
	}

	return formatTimeWindow(timeWindow.from, timeWindow.to, " -> ");
}
