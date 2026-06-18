import { useNavigate } from "react-router-dom";
import { pathForRoute } from "./routePaths";
import type { Navigate } from "./routeTypes";

export function useRouteNavigate(defaultProjectRef?: string | null): Navigate {
	const navigate = useNavigate();

	return (route, options) => navigate(pathForRoute(route, { projectRef: options?.projectRef ?? defaultProjectRef }));
}
