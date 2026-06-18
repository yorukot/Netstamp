import { TrackingConsentBanner } from "@/shared/tracking/TrackingConsentBanner";
import { Outlet } from "react-router-dom";

export function RouteFrame() {
	return (
		<>
			<Outlet />
			<TrackingConsentBanner />
		</>
	);
}
