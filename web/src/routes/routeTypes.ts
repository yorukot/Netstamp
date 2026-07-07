export type ProjectAppRoute = "dashboard" | "probes" | "newProbe" | "labels" | "insight" | "checks" | "alerts" | "statusPages" | "members" | "projectSettings";
export type GlobalAppRoute = "accountSettings" | "adminSettings";
export type AppRoute = ProjectAppRoute | GlobalAppRoute;
export type PublicRoute = "landing" | "login" | "register" | "forgotPassword" | "resetPassword" | "verifyEmail" | "onboarding";
export type Route = AppRoute | PublicRoute;
export interface NavigateOptions {
	projectRef?: string | null;
}
export type Navigate = (route: Route, options?: NavigateOptions) => void;
