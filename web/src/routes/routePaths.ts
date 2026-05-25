import type { Route } from "./routeTypes";

export const routePaths = {
	landing: "/",
	login: "/login",
	register: "/register",
	onboarding: "/onboarding",
	dashboard: "/dashboard",
	probes: "/probes",
	newProbe: "/probes/new",
	labels: "/labels",
	insight: "/insight",
	checks: "/checks",
	project: "/project",
	settings: "/settings"
} satisfies Record<Route, string>;

export function pathForRoute(route: Route) {
	return routePaths[route];
}
