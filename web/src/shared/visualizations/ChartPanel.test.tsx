// @vitest-environment jsdom

import { ThemeContext, type AppTheme } from "@/shared/theme/themeContext";
import type { ChartOption } from "@/shared/visualizations/chartOptions";
import type { ChartTheme } from "@/shared/visualizations/chartTheme";
import { cleanup, render, waitFor } from "@testing-library/react";
import { useLayoutEffect, type ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ChartPanel } from "./ChartPanel";

const chartRuntime = vi.hoisted(() => {
	const instance = {
		dispatchAction: vi.fn(),
		dispose: vi.fn(),
		off: vi.fn(),
		on: vi.fn(),
		resize: vi.fn(),
		setOption: vi.fn()
	};

	return {
		init: vi.fn(() => instance),
		instance,
		use: vi.fn()
	};
});

vi.mock("echarts/core", () => ({
	init: chartRuntime.init,
	use: chartRuntime.use
}));

vi.mock("echarts/charts", () => ({
	BarChart: {},
	CustomChart: {},
	GraphChart: {},
	LineChart: {},
	ScatterChart: {}
}));

vi.mock("echarts/components", () => ({
	DataZoomComponent: {},
	GridComponent: {},
	LegendComponent: {},
	ToolboxComponent: {},
	TooltipComponent: {},
	VisualMapComponent: {}
}));

vi.mock("echarts/renderers", () => ({
	CanvasRenderer: {}
}));

const primaryByTheme: Record<AppTheme, string> = {
	dark: "dark-primary",
	light: "light-primary"
};

const themeContextValue = {
	theme: "dark" as const,
	setTheme: vi.fn(),
	toggleTheme: vi.fn()
};

const ScopedThemeBoundary = ({ children, theme }: { children: ReactNode; theme: AppTheme }) => {
	useLayoutEffect(() => {
		document.documentElement.dataset.theme = theme;
	}, [theme]);

	return children;
};

const renderChart = (theme: AppTheme, getOption: (chartTheme: ChartTheme) => ChartOption) => (
	<ThemeContext.Provider value={themeContextValue}>
		<ScopedThemeBoundary theme={theme}>
			<ChartPanel getOption={getOption} theme={theme} />
		</ScopedThemeBoundary>
	</ThemeContext.Provider>
);

beforeEach(() => {
	document.documentElement.dataset.theme = "dark";
	vi.stubGlobal(
		"getComputedStyle",
		vi.fn(() => {
			const scopedTheme = document.documentElement.dataset.theme === "light" ? "light" : "dark";

			return {
				getPropertyValue: (name: string) => (name === "--ns-primary" ? primaryByTheme[scopedTheme] : "")
			} as CSSStyleDeclaration;
		})
	);
	vi.stubGlobal(
		"ResizeObserver",
		class {
			disconnect = vi.fn();
			observe = vi.fn();
		}
	);
	vi.clearAllMocks();
});

afterEach(() => {
	cleanup();
	vi.unstubAllGlobals();
	delete document.documentElement.dataset.theme;
});

describe("ChartPanel", () => {
	it("rebuilds options after a scoped document theme is applied", async () => {
		const getOption = vi.fn((theme: ChartTheme) => ({ color: theme.primary }));
		const { rerender } = render(renderChart("light", getOption));

		expect(getOption.mock.calls[0]?.[0].primary).toBe(primaryByTheme.dark);
		await waitFor(() => expect(getOption.mock.lastCall?.[0].primary).toBe(primaryByTheme.light));
		await waitFor(() => expect(chartRuntime.instance.setOption).toHaveBeenLastCalledWith({ color: primaryByTheme.light }, { notMerge: true }));

		rerender(renderChart("dark", getOption));

		await waitFor(() => expect(getOption.mock.lastCall?.[0].primary).toBe(primaryByTheme.dark));
		await waitFor(() => expect(chartRuntime.instance.setOption).toHaveBeenLastCalledWith({ color: primaryByTheme.dark }, { notMerge: true }));
	});
});
