export type ProjectAppRoute = "dashboard" | "probes" | "newProbe" | "labels" | "insight" | "checks" | "alerts" | "members" | "project";
export type GlobalAppRoute = "settings";
export type AppRoute = ProjectAppRoute | GlobalAppRoute;
export type PublicRoute = "landing" | "login" | "register" | "onboarding";
export type Route = AppRoute | PublicRoute;
export interface NavigateOptions {
	projectRef?: string | null;
}
export type Navigate = (route: Route, options?: NavigateOptions) => void;
