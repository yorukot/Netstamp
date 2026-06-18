import { Suspense, type ReactNode } from "react";

export function lazyRoute(element: ReactNode) {
	return <Suspense fallback={null}>{element}</Suspense>;
}
