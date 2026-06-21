import { BellRinging, Broadcast, ChartLineUp, CheckCircle, Gauge, GearSix, Globe, Tag, UsersThree, type Icon } from "@phosphor-icons/react";
import type { AppRoute } from "./routeTypes";

export interface SidebarItem {
	label: string;
	route: AppRoute;
	icon: Icon;
}

export const sidebarItems: SidebarItem[] = [
	{ label: "Dashboard", route: "dashboard", icon: Gauge },
	{ label: "Probes", route: "probes", icon: Broadcast },
	{ label: "Checks", route: "checks", icon: CheckCircle },
	{ label: "Alerts", route: "alerts", icon: BellRinging },
	{ label: "Status", route: "statusPages", icon: Globe },
	{ label: "Labels", route: "labels", icon: Tag },
	{ label: "Insight", route: "insight", icon: ChartLineUp },
	{ label: "Members", route: "members", icon: UsersThree },
	{ label: "Settings", route: "projectSettings", icon: GearSix }
];
