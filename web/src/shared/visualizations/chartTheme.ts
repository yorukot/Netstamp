const sansFontToken = "--ns-font-sans";
const fallbackSansFont = "sans-serif";
const mutedTextColor = "#64748B";
const tooltipTextColor = "#111827";

export function chartFontFamily() {
	if (typeof document === "undefined") {
		return fallbackSansFont;
	}

	return getComputedStyle(document.documentElement).getPropertyValue(sansFontToken).trim() || fallbackSansFont;
}

export function chartAxisLabel(fontSize = 10) {
	return { color: mutedTextColor, fontFamily: chartFontFamily(), fontSize };
}

export function chartMutedTextStyle(fontSize = 10) {
	return { color: mutedTextColor, fontFamily: chartFontFamily(), fontSize };
}

export function chartTooltipTextStyle(fontSize?: number) {
	return {
		color: tooltipTextColor,
		fontFamily: chartFontFamily(),
		...(fontSize ? { fontSize } : {})
	};
}
