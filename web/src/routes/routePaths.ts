import type { Route } from "@/shared/utils/mockData";

export const routePaths = {
	landing: "/",
	login: "/login",
	register: "/register",
	onboarding: "/onboarding",
	dashboard: "/dashboard",
	probes: "/probes",
	insight: "/insight",
	checks: "/checks",
	alerts: "/alerts",
	team: "/team",
	settings: "/settings"
} satisfies Record<Route, string>;

export function pathForRoute(route: Route) {
	return routePaths[route];
}
