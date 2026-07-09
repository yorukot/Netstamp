import { TrackingConsentBanner } from "@/shared/tracking/TrackingConsentBanner";
import { Helmet } from "react-helmet-async";
import { Outlet, useMatches } from "react-router-dom";
import { pageTitleFromMatches } from "./pageTitles";

export function RouteFrame() {
	const matches = useMatches();
	const title = pageTitleFromMatches(matches);

	return (
		<>
			<Helmet>
				<title>{title}</title>
			</Helmet>
			<Outlet />
			<TrackingConsentBanner />
		</>
	);
}
