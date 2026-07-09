import { BellRingingIcon } from "@phosphor-icons/react/dist/csr/BellRinging";
import { BroadcastIcon } from "@phosphor-icons/react/dist/csr/Broadcast";
import { ChartLineUpIcon } from "@phosphor-icons/react/dist/csr/ChartLineUp";
import { CheckCircleIcon } from "@phosphor-icons/react/dist/csr/CheckCircle";
import { GaugeIcon } from "@phosphor-icons/react/dist/csr/Gauge";
import { GearSixIcon } from "@phosphor-icons/react/dist/csr/GearSix";
import { GlobeIcon } from "@phosphor-icons/react/dist/csr/Globe";
import { ShieldCheckIcon } from "@phosphor-icons/react/dist/csr/ShieldCheck";
import { TagIcon } from "@phosphor-icons/react/dist/csr/Tag";
import { UsersThreeIcon } from "@phosphor-icons/react/dist/csr/UsersThree";
import type { Icon } from "@phosphor-icons/react/dist/lib/types";
import type { AppRoute } from "./routeTypes";

export interface SidebarItem {
	label: string;
	route: AppRoute;
	icon: Icon;
	systemAdminOnly?: boolean;
}

export const sidebarItems: SidebarItem[] = [
	{ label: "Dashboard", route: "dashboard", icon: GaugeIcon },
	{ label: "Probes", route: "probes", icon: BroadcastIcon },
	{ label: "Checks", route: "checks", icon: CheckCircleIcon },
	{ label: "Alerts", route: "alerts", icon: BellRingingIcon },
	{ label: "Status", route: "statusPages", icon: GlobeIcon },
	{ label: "Labels", route: "labels", icon: TagIcon },
	{ label: "Insight", route: "insight", icon: ChartLineUpIcon },
	{ label: "Members", route: "members", icon: UsersThreeIcon },
	{ label: "Settings", route: "projectSettings", icon: GearSixIcon },
	{ label: "Admin", route: "adminSettings", icon: ShieldCheckIcon, systemAdminOnly: true }
];
