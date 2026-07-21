import { BellRingingIcon } from "@phosphor-icons/react/dist/csr/BellRinging";
import { BroadcastIcon } from "@phosphor-icons/react/dist/csr/Broadcast";
import { ChartLineUpIcon } from "@phosphor-icons/react/dist/csr/ChartLineUp";
import { CheckCircleIcon } from "@phosphor-icons/react/dist/csr/CheckCircle";
import { GaugeIcon } from "@phosphor-icons/react/dist/csr/Gauge";
import { GearSixIcon } from "@phosphor-icons/react/dist/csr/GearSix";
import { GlobeIcon } from "@phosphor-icons/react/dist/csr/Globe";
import { TagIcon } from "@phosphor-icons/react/dist/csr/Tag";
import { UsersThreeIcon } from "@phosphor-icons/react/dist/csr/UsersThree";
import type { Icon } from "@phosphor-icons/react/dist/lib/types";
import type { ProjectAppRoute } from "./routeTypes";

export interface SidebarItem {
	labelKey: "alerts" | "checks" | "insight" | "labels" | "members" | "overview" | "probes" | "settings" | "status";
	route: ProjectAppRoute;
	icon: Icon;
}

export const sidebarItems: SidebarItem[] = [
	{ labelKey: "overview", route: "dashboard", icon: GaugeIcon },
	{ labelKey: "probes", route: "probes", icon: BroadcastIcon },
	{ labelKey: "checks", route: "checks", icon: CheckCircleIcon },
	{ labelKey: "alerts", route: "alerts", icon: BellRingingIcon },
	{ labelKey: "status", route: "statusPages", icon: GlobeIcon },
	{ labelKey: "labels", route: "labels", icon: TagIcon },
	{ labelKey: "insight", route: "insight", icon: ChartLineUpIcon },
	{ labelKey: "members", route: "members", icon: UsersThreeIcon },
	{ labelKey: "settings", route: "projectSettings", icon: GearSixIcon }
];
