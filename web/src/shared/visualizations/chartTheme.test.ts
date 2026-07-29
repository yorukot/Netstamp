// @vitest-environment jsdom

import { afterEach, describe, expect, it } from "vitest";
import { chartAxisLabel, chartTheme, chartTooltipTextStyle } from "./chartTheme";

afterEach(() => {
	document.body.replaceChildren();
});

describe("chartTheme", () => {
	it("resolves chart colors from the provided theme scope", () => {
		const scope = document.createElement("div");
		scope.style.setProperty("--ns-font-sans", "Scoped Sans");
		scope.style.setProperty("--ns-primary", "#123456");
		scope.style.setProperty("--ns-primary-hover", "#234567");
		scope.style.setProperty("--ns-secondary", "#345678");
		scope.style.setProperty("--ns-secondary-hover", "#456789");
		scope.style.setProperty("--ns-success", "#168a45");
		scope.style.setProperty("--ns-warning", "#b7791f");
		scope.style.setProperty("--ns-critical", "#c9362c");
		scope.style.setProperty("--ns-metal", "#64748b");
		scope.style.setProperty("--ns-text-muted", "#405168");
		scope.style.setProperty("--ns-text", "#172033");
		scope.style.setProperty("--ns-surface", "#ffffff");
		scope.style.setProperty("--ns-border", "#d7dbe2");
		document.body.append(scope);

		const theme = chartTheme(scope);

		expect(theme.primary).toBe("#123456");
		expect(theme.tooltipBackground).toBe("rgba(255, 255, 255, 0.98)");
		expect(theme.tooltipText).toBe("#172033");
		expect(theme.seriesPalette.slice(0, 2)).toEqual(["#123456", "#345678"]);
		expect(chartAxisLabel(theme)).toMatchObject({ color: "#405168", fontFamily: "Scoped Sans" });
		expect(chartTooltipTextStyle(theme)).toMatchObject({ color: "#172033", fontFamily: "Scoped Sans" });
	});
});
