import type { AppRoute } from "./routeTypes";

export interface SidebarItem {
	label: string;
	route: AppRoute;
}

export const sidebarItems: SidebarItem[] = [
	{ label: "Dashboard", route: "dashboard" },
	{ label: "Probes", route: "probes" },
	{ label: "Labels", route: "labels" },
	{ label: "Checks", route: "checks" },
	{ label: "Insight", route: "insight" },
	{ label: "Public Pages", route: "publicPages" },
	{ label: "Members", route: "members" },
	{ label: "Project", route: "project" }
];
