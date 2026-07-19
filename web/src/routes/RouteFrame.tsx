import { TrackingConsentBanner } from "@/shared/tracking/TrackingConsentBanner";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Outlet, useMatches } from "react-router-dom";
import { pageTitleFromMatches } from "./pageTitles";

export function RouteFrame() {
	const { t } = useTranslation("navigation");
	const matches = useMatches();
	const title = pageTitleFromMatches(matches, t);

	useEffect(() => {
		document.title = title;
	}, [title]);

	return (
		<>
			<Outlet />
			<TrackingConsentBanner />
		</>
	);
}
