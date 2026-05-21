import type { AppRoute } from "./routeTypes";

export interface SidebarItem {
	label: string;
	route: AppRoute;
}

export const sidebarItems: SidebarItem[] = [
	{ label: "Dashboard", route: "dashboard" },
	{ label: "Probes", route: "probes" },
	{ label: "Insight", route: "insight" },
	{ label: "Checks", route: "checks" },
	{ label: "Alerts", route: "alerts" },
	{ label: "Project", route: "project" }
];
