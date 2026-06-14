import type { AppRoute, NavigateOptions, ProjectAppRoute, PublicRoute, Route } from "./routeTypes";

export const publicRoutePaths = {
	landing: "/",
	login: "/login",
	register: "/register",
	onboarding: "/onboarding"
} satisfies Record<PublicRoute, string>;

export const legacyAppRoutePaths = {
	dashboard: "/dashboard",
	probes: "/probes",
	newProbe: "/probes/new",
	labels: "/labels",
	insight: "/insight",
	checks: "/checks",
	alerts: "/alerts",
	members: "/members",
	project: "/project",
	settings: "/settings"
} satisfies Record<AppRoute, string>;

export const projectRouteSegments = {
	dashboard: "dashboard",
	probes: "probes",
	newProbe: "probes/new",
	labels: "labels",
	insight: "insight",
	checks: "checks",
	alerts: "alerts",
	members: "members",
	project: "project"
} satisfies Record<ProjectAppRoute, string>;

export const routePaths = {
	...publicRoutePaths,
	...legacyAppRoutePaths
} satisfies Record<Route, string>;

function hasOwnKey<T extends object>(value: T, key: PropertyKey): key is keyof T {
	return Object.prototype.hasOwnProperty.call(value, key);
}

function encodePathParam(value: string) {
	return encodeURIComponent(value);
}

function projectBasePath(projectRef: string) {
	return `/projects/${encodePathParam(projectRef)}`;
}

export function isProjectAppRoute(route: Route): route is ProjectAppRoute {
	return hasOwnKey(projectRouteSegments, route);
}

export function projectRoutePath(route: ProjectAppRoute) {
	return projectRouteSegments[route];
}

export function pathForRoute(route: Route, options: NavigateOptions = {}) {
	if (hasOwnKey(publicRoutePaths, route)) {
		return publicRoutePaths[route];
	}

	if (isProjectAppRoute(route) && options.projectRef) {
		return `${projectBasePath(options.projectRef)}/${projectRouteSegments[route]}`;
	}

	return legacyAppRoutePaths[route];
}

export function pathForProbeDetail(projectRef: string | null | undefined, probeId: string) {
	if (!projectRef) {
		return `/probes/${encodePathParam(probeId)}`;
	}

	return `${projectBasePath(projectRef)}/probes/${encodePathParam(probeId)}`;
}

export function pathForLabelDetail(projectRef: string | null | undefined, labelId: string) {
	if (!projectRef) {
		return `/labels/${encodePathParam(labelId)}`;
	}

	return `${projectBasePath(projectRef)}/labels/${encodePathParam(labelId)}`;
}

export function pathForCheckDetail(projectRef: string | null | undefined, checkId: string) {
	if (!projectRef) {
		return `/checks/${encodePathParam(checkId)}`;
	}

	return `${projectBasePath(projectRef)}/checks/${encodePathParam(checkId)}`;
}

export function pathForProjectSwitch(pathname: string, projectRef: string) {
	const match = /^\/projects\/[^/]+(?:\/([^/]+))?/.exec(pathname);

	if (!match) {
		return null;
	}

	const segment = match[1] || projectRouteSegments.dashboard;
	const route = (Object.entries(projectRouteSegments).find(([, routeSegment]) => routeSegment.split("/")[0] === segment)?.[0] ?? "dashboard") as ProjectAppRoute;

	return pathForRoute(route, { projectRef });
}
