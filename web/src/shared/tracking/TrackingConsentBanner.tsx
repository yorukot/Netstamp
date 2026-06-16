import { Button } from "@netstamp/ui";
import { hasEnabledTrackers, loadConfiguredTrackers, readTrackingConsent, shouldRequestTrackingConsent, trackConfiguredPageView, writeTrackingConsent } from "@netstamp/ui/tracking";
import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import styles from "./TrackingConsentBanner.module.css";
import { trackerConfig } from "./trackerConfig";

export function TrackingConsentBanner() {
	const location = useLocation();
	const [visible, setVisible] = useState(false);
	const [trackingReady, setTrackingReady] = useState(false);
	const lastTrackedLocation = useRef("");
	const enabled = hasEnabledTrackers(trackerConfig);

	const activateTracking = useCallback(async () => {
		await loadConfiguredTrackers(trackerConfig);
		setTrackingReady(true);
	}, []);

	useEffect(() => {
		let cancelled = false;

		async function syncTrackingConsent() {
			if (!enabled) {
				return;
			}

			const storedConsent = readTrackingConsent(trackerConfig);

			if (storedConsent === "accepted") {
				await activateTracking();
				return;
			}

			if (storedConsent === "declined") {
				return;
			}

			if (await shouldRequestTrackingConsent(trackerConfig)) {
				if (!cancelled) {
					setVisible(true);
				}
				return;
			}

			await activateTracking();
		}

		void syncTrackingConsent();

		return () => {
			cancelled = true;
		};
	}, [activateTracking, enabled]);

	useEffect(() => {
		if (!trackingReady) {
			return;
		}

		const pageLocation = window.location.href;

		if (pageLocation === lastTrackedLocation.current) {
			return;
		}

		lastTrackedLocation.current = pageLocation;
		trackConfiguredPageView(trackerConfig);
	}, [location.hash, location.pathname, location.search, trackingReady]);

	async function acceptTracking() {
		writeTrackingConsent(trackerConfig, "accepted");
		setVisible(false);
		await activateTracking();
	}

	function declineTracking() {
		writeTrackingConsent(trackerConfig, "declined");
		setVisible(false);
	}

	if (!enabled || !visible) {
		return null;
	}

	return (
		<section className={styles.banner} aria-label="Analytics consent">
			<div className={styles.copy}>
				<strong>Analytics consent</strong>
				<p>Netstamp can load optional analytics to understand product and docs usage. Declining keeps those trackers disabled in this browser.</p>
			</div>
			<div className={styles.actions}>
				<Button type="button" variant="secondary" size="sm" onClick={declineTracking}>
					Decline
				</Button>
				<Button type="button" size="sm" onClick={() => void acceptTracking()}>
					Accept
				</Button>
			</div>
		</section>
	);
}
