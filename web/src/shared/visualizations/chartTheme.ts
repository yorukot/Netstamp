const sansFontToken = "--ns-font-sans";
const fallbackSansFont = "sans-serif";

export interface ChartTheme {
	axisLine: string;
	baselineLine: string;
	critical: string;
	criticalBand: string;
	criticalBandStrong: string;
	metal: string;
	primary: string;
	primaryBrush: string;
	primaryBrushBorder: string;
	primaryFill: string;
	secondary: string;
	seriesPalette: string[];
	spreadFill: string;
	splitLine: string;
	success: string;
	textMuted: string;
	tooltipBackground: string;
	tooltipBorder: string;
	tooltipText: string;
	warning: string;
	warningBand: string;
}

function tokenValue(name: string, fallback: string) {
	if (typeof document === "undefined") {
		return fallback;
	}

	return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || fallback;
}

function rgba(color: string, alpha: number, fallback: string) {
	const hex = color.match(/^#([0-9a-f]{6})$/i);
	if (hex) {
		const value = hex[1];
		const red = Number.parseInt(value.slice(0, 2), 16);
		const green = Number.parseInt(value.slice(2, 4), 16);
		const blue = Number.parseInt(value.slice(4, 6), 16);

		return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
	}

	const rgb = color.match(/^rgba?\(([^)]+)\)$/i);
	if (rgb) {
		const [red, green, blue] = rgb[1].split(",").map(part => part.trim());
		if (red && green && blue) {
			return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
		}
	}

	return fallback;
}

export function chartFontFamily() {
	if (typeof document === "undefined") {
		return fallbackSansFont;
	}

	return getComputedStyle(document.documentElement).getPropertyValue(sansFontToken).trim() || fallbackSansFont;
}

export function chartTheme(): ChartTheme {
	const primary = tokenValue("--ns-primary", "#fb923c");
	const secondary = tokenValue("--ns-secondary", "#38bdf8");
	const success = tokenValue("--ns-success", "#34c77b");
	const warning = tokenValue("--ns-warning", "#facc15");
	const critical = tokenValue("--ns-critical", "#ff6b63");
	const metal = tokenValue("--ns-metal", "#c4ccd9");
	const textMuted = tokenValue("--ns-text-muted", "#d8dce8");
	const text = tokenValue("--ns-text", "#f8fafc");
	const surface = tokenValue("--ns-surface", "#070707");
	const border = tokenValue("--ns-border", "#2d3035");

	return {
		axisLine: rgba(metal, 0.16, "rgba(196, 204, 217, 0.16)"),
		baselineLine: rgba(metal, 0.52, "rgba(196, 204, 217, 0.52)"),
		critical,
		criticalBand: rgba(critical, 0.14, "rgba(255, 107, 99, 0.14)"),
		criticalBandStrong: rgba(critical, 0.2, "rgba(255, 107, 99, 0.2)"),
		metal,
		primary,
		primaryBrush: rgba(primary, 0.12, "rgba(251, 146, 60, 0.12)"),
		primaryBrushBorder: rgba(primary, 0.72, "rgba(251, 146, 60, 0.72)"),
		primaryFill: rgba(primary, 0.12, "rgba(251, 146, 60, 0.12)"),
		secondary,
		seriesPalette: [primary, secondary, success, warning, metal, critical, tokenValue("--ns-primary-hover", "#fdba74"), tokenValue("--ns-secondary-hover", "#7dd3fc")],
		spreadFill: rgba(metal, 0.08, "rgba(196, 204, 217, 0.08)"),
		splitLine: rgba(metal, 0.18, "rgba(196, 204, 217, 0.18)"),
		success,
		textMuted,
		tooltipBackground: rgba(surface, 0.98, "rgba(7, 7, 7, 0.98)"),
		tooltipBorder: rgba(border, 0.92, "rgba(45, 48, 53, 0.92)"),
		tooltipText: text,
		warning,
		warningBand: rgba(warning, 0.12, "rgba(250, 204, 21, 0.12)")
	};
}

export function chartAxisLabel(fontSize = 10) {
	return { color: chartTheme().textMuted, fontFamily: chartFontFamily(), fontSize };
}

export function chartMutedTextStyle(fontSize = 10) {
	return { color: chartTheme().textMuted, fontFamily: chartFontFamily(), fontSize };
}

export function chartTooltipTextStyle(fontSize?: number) {
	return {
		color: chartTheme().tooltipText,
		fontFamily: chartFontFamily(),
		...(fontSize ? { fontSize } : {})
	};
}
