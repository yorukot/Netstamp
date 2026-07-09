import { Spinner } from "@netstamp/ui";
import { Suspense, type ReactNode } from "react";

export function lazyRoute(element: ReactNode) {
	return <Suspense fallback={<Spinner label="Loading route" layout="page" size="lg" />}>{element}</Suspense>;
}
