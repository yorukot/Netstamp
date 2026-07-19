import { Suspense, type ReactNode } from "react";
import { RouteLoadingSpinner } from "./RouteLoadingSpinner";

export function lazyRoute(element: ReactNode) {
	return <Suspense fallback={<RouteLoadingSpinner />}>{element}</Suspense>;
}
