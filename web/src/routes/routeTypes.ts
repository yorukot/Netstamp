export type AppRoute = "dashboard" | "probes" | "newProbe" | "insight" | "checks" | "alerts" | "project" | "settings";
export type PublicRoute = "landing" | "login" | "register" | "onboarding";
export type Route = AppRoute | PublicRoute;
export type Navigate = (route: Route) => void;
