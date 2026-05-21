import type { Route } from "./routeTypes";

export const routePaths = {
	landing: "/",
	login: "/login",
	register: "/register",
	onboarding: "/onboarding",
	dashboard: "/dashboard",
	probes: "/probes",
	newProbe: "/probes/new",
	insight: "/insight",
	checks: "/checks",
	alerts: "/alerts",
	project: "/project",
	settings: "/settings"
} satisfies Record<Route, string>;

export function pathForRoute(route: Route) {
	return routePaths[route];
}
