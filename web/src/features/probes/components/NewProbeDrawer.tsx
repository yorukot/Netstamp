import {
	coordinateInputError,
	formatCoordinate,
	parseCoordinateInput,
	searchNominatimLocation,
	type CoordinateInputMode,
	type GeocodeStatus,
	type ProbeCoordinates
} from "@/features/probes/data/probeLocation";
import { pathForRoute } from "@/routes/routePaths";
import { installAssetPaths, installAssetUrl, probeInstallCommand } from "@/shared/api/installAssets";
import { useCreateProjectProbeMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { CloseButton } from "@/shared/components/CloseButton";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Terminal, TextField } from "@netstamp/ui";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState, type FormEvent, type MouseEvent } from "react";
import { useNavigate } from "react-router-dom";
import { LocationPreviewMap } from "./LocationPreviewMap";
import styles from "./NewProbeDrawer.module.css";
import { ProbeWizardTimeline } from "./ProbeWizardTimeline";

const drawerCloseDurationMs = 180;
const createProbeSteps = [
	{ number: "01", title: "Name", copy: "Probe identity" },
	{ number: "02", title: "Install", copy: "Run command" }
];

async function writeClipboardText(value: string) {
	if (navigator.clipboard?.writeText) {
		await navigator.clipboard.writeText(value);
		return;
	}

	const textarea = document.createElement("textarea");
	textarea.value = value;
	textarea.setAttribute("readonly", "");
	textarea.style.position = "fixed";
	textarea.style.left = "-9999px";
	document.body.appendChild(textarea);
	textarea.select();
	document.execCommand("copy");
	textarea.remove();
}

export function NewProbeDrawer() {
	const navigate = useNavigate();
	const { projectRef } = useCurrentProject();
	const queryClient = useQueryClient();
	const closeTimeoutRef = useRef<number | null>(null);
	const copyTimeoutRef = useRef<number | null>(null);
	const geocodeAbortRef = useRef<AbortController | null>(null);
	const [closing, setClosing] = useState(false);
	const [currentStep, setCurrentStep] = useState(0);
	const [installStatus, setInstallStatus] = useState<"idle" | "detecting">("idle");
	const [installCommandCopied, setInstallCommandCopied] = useState(false);
	const [probeName, setProbeName] = useState("");
	const [coordinateInputMode, setCoordinateInputMode] = useState<CoordinateInputMode>("search");
	const [locationSearch, setLocationSearch] = useState("");
	const [locationSearchEdited, setLocationSearchEdited] = useState(false);
	const [locationName, setLocationName] = useState("");
	const [latitudeInput, setLatitudeInput] = useState("");
	const [longitudeInput, setLongitudeInput] = useState("");
	const [geocodeStatus, setGeocodeStatus] = useState<GeocodeStatus>("idle");
	const [geocodeError, setGeocodeError] = useState("");
	const [registrationSecret, setRegistrationSecret] = useState("");
	const [registeredProbeId, setRegisteredProbeId] = useState("");
	const createProbeMutation = useCreateProjectProbeMutation(projectRef);
	const createdProbeQuery = useQuery({
		...projectQueries.probeDetail(projectRef || "", registeredProbeId),
		enabled: Boolean(projectRef && registeredProbeId && currentStep === 1 && installStatus === "detecting"),
		refetchInterval: query => {
			const status = query.state.data?.probe.status;

			return status?.state === "online" || status?.lastSeenAt ? false : 2500;
		},
		refetchIntervalInBackground: true
	});
	const heartbeatReceived = Boolean(createdProbeQuery.data?.probe.status?.state === "online" || createdProbeQuery.data?.probe.status?.lastSeenAt);
	const latitude = parseCoordinateInput(latitudeInput);
	const longitude = parseCoordinateInput(longitudeInput);
	const latitudeError = coordinateInputMode === "manual" ? coordinateInputError("Latitude", latitudeInput, -90, 90) : "";
	const longitudeError = coordinateInputMode === "manual" ? coordinateInputError("Longitude", longitudeInput, -180, 180) : "";
	const visibleLatitudeError = latitudeInput.trim() ? latitudeError : "";
	const visibleLongitudeError = longitudeInput.trim() ? longitudeError : "";
	const searchCoordinatesReady = coordinateInputMode === "search" && geocodeStatus === "resolved" && latitude !== null && longitude !== null;
	const manualCoordinatesReady = coordinateInputMode === "manual" && latitude !== null && longitude !== null && !latitudeError && !longitudeError;
	const coordinatesReady = searchCoordinatesReady || manualCoordinatesReady;
	const selectedCoordinates: ProbeCoordinates | null = coordinatesReady && latitude !== null && longitude !== null ? { latitude, longitude } : null;
	const canSearchLocation = coordinateInputMode === "search" && locationSearch.trim().length > 0 && geocodeStatus !== "searching";
	const canCreate = probeName.trim().length > 0 && Boolean(projectRef) && Boolean(selectedCoordinates);
	const previewTitle = locationName.trim() || "Manual coordinates";
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);
	const uninstallerUrl = installAssetUrl(installAssetPaths.agentUninstaller);
	const binaryUrl = installAssetUrl(installAssetPaths.linuxAmd64Binary);
	const installCommand = registeredProbeId && registrationSecret ? probeInstallCommand({ probeId: registeredProbeId, probeSecret: registrationSecret }) : "";

	useEffect(() => {
		return () => {
			if (closeTimeoutRef.current) {
				window.clearTimeout(closeTimeoutRef.current);
			}
			if (copyTimeoutRef.current) {
				window.clearTimeout(copyTimeoutRef.current);
			}
			geocodeAbortRef.current?.abort();
		};
	}, []);

	useEffect(() => {
		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		return () => {
			document.body.style.overflow = previousOverflow;
		};
	}, []);

	useEffect(() => {
		if (heartbeatReceived) {
			void queryClient.invalidateQueries({ queryKey: projectQueries.probes(projectRef || "").queryKey });
		}
	}, [heartbeatReceived, projectRef, queryClient]);

	useEffect(() => {
		function handleKeyDown(event: KeyboardEvent) {
			if (event.key !== "Escape" || closing || closeTimeoutRef.current) {
				return;
			}

			setClosing(true);
			closeTimeoutRef.current = window.setTimeout(() => navigate(pathForRoute("probes")), drawerCloseDurationMs);
		}

		window.addEventListener("keydown", handleKeyDown);
		return () => window.removeEventListener("keydown", handleKeyDown);
	}, [closing, navigate]);

	function closeDrawer() {
		if (closing || closeTimeoutRef.current) {
			return;
		}

		setClosing(true);
		closeTimeoutRef.current = window.setTimeout(() => navigate(pathForRoute("probes")), drawerCloseDurationMs);
	}

	function clearResolvedCoordinates() {
		geocodeAbortRef.current?.abort();
		geocodeAbortRef.current = null;
		setLocationName("");
		setLatitudeInput("");
		setLongitudeInput("");
		setGeocodeStatus("idle");
		setGeocodeError("");
	}

	function updateProbeName(value: string) {
		setProbeName(value);
		setInstallStatus("idle");
		setRegisteredProbeId("");
		setRegistrationSecret("");

		if (!locationSearchEdited) {
			setLocationSearch(value);

			if (coordinateInputMode === "search") {
				clearResolvedCoordinates();
			}
		}
	}

	function updateLocationSearch(value: string) {
		setLocationSearch(value);
		setLocationSearchEdited(true);
		clearResolvedCoordinates();
	}

	function updateCoordinateInputMode(nextMode: CoordinateInputMode) {
		if (nextMode === coordinateInputMode) {
			return;
		}

		setCoordinateInputMode(nextMode);
		setGeocodeError("");
		setGeocodeStatus("idle");

		if (nextMode === "search") {
			setLocationName("");
			setLatitudeInput("");
			setLongitudeInput("");
		}
	}

	function startInstallDetection() {
		setInstallStatus("detecting");
		setCurrentStep(1);
	}

	async function copyInstallCommand() {
		if (!installCommand) {
			return;
		}

		try {
			await writeClipboardText(installCommand);
		} catch {
			setInstallCommandCopied(false);
			return;
		}

		setInstallCommandCopied(true);

		if (copyTimeoutRef.current) {
			window.clearTimeout(copyTimeoutRef.current);
		}

		copyTimeoutRef.current = window.setTimeout(() => {
			setInstallCommandCopied(false);
			copyTimeoutRef.current = null;
		}, 1800);
	}

	async function searchLocation() {
		const query = locationSearch.trim();

		if (!query || geocodeStatus === "searching") {
			return;
		}

		geocodeAbortRef.current?.abort();
		const abortController = new AbortController();
		geocodeAbortRef.current = abortController;

		setGeocodeStatus("searching");
		setGeocodeError("");
		setLocationName("");
		setLatitudeInput("");
		setLongitudeInput("");

		try {
			const result = await searchNominatimLocation(query, abortController.signal);

			setLocationName(result.locationName);
			setLatitudeInput(formatCoordinate(result.coordinates.latitude));
			setLongitudeInput(formatCoordinate(result.coordinates.longitude));
			setGeocodeStatus("resolved");
		} catch (error) {
			if (error instanceof DOMException && error.name === "AbortError") {
				return;
			}

			setGeocodeStatus("error");
			setGeocodeError(error instanceof Error ? error.message : "Location search failed.");
		} finally {
			if (geocodeAbortRef.current === abortController) {
				geocodeAbortRef.current = null;
			}
		}
	}

	async function handleNameSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!canCreate || !selectedCoordinates) {
			return;
		}

		const data = await createProbeMutation.mutateAsync({
			enabled: true,
			name: probeName.trim(),
			...(locationName.trim() ? { locationName: locationName.trim() } : {}),
			latitude: selectedCoordinates.latitude,
			longitude: selectedCoordinates.longitude
		});
		setRegisteredProbeId(data.probe.id);
		setRegistrationSecret(data.secret);
		startInstallDetection();
	}

	function handleBackdropClick(event: MouseEvent<HTMLDivElement>) {
		if (event.target === event.currentTarget) {
			closeDrawer();
		}
	}

	return (
		<div className={classNames(styles.backdrop, closing && styles.backdropClosing)} onClick={handleBackdropClick}>
			<aside className={classNames(styles.drawer, closing && styles.drawerClosing)} aria-label="New probe wizard">
				<div className={styles.header}>
					<div>
						<button type="button" className={styles.backLink} onClick={closeDrawer}>
							back to probes
						</button>
						<h2>Create probe</h2>
						<p>Name the probe, install it on a host, then wait for the controller to receive its first heartbeat.</p>
					</div>
					<CloseButton ariaLabel="Close probe wizard" onClick={closeDrawer} />
				</div>

				<ProbeWizardTimeline steps={createProbeSteps} currentStep={currentStep} />

				<div className={styles.workflowViewport}>
					<div className={styles.workflowTrack} style={{ transform: `translateX(-${currentStep * 100}%)` }}>
						<form className={classNames("ns-scrollbar", styles.workflowPanel)} aria-hidden={currentStep !== 0} onSubmit={handleNameSubmit}>
							<div className={styles.stepCopy}>
								<Badge tone="accent">Step 01</Badge>
								<h3>Enter probe identity</h3>
								<p>The probe name and coordinates are stored before the install command is generated.</p>
							</div>

							<TextField label="Probe name" value={probeName} placeholder="taipei-home-01" required disabled={currentStep !== 0} onChange={event => updateProbeName(event.currentTarget.value)} />

							<div className={styles.locationMode} role="group" aria-label="Coordinate input mode">
								<Button
									type="button"
									variant={coordinateInputMode === "search" ? "secondary" : "ghost"}
									size="sm"
									aria-pressed={coordinateInputMode === "search"}
									disabled={currentStep !== 0}
									onClick={() => updateCoordinateInputMode("search")}
								>
									Search name
								</Button>
								<Button
									type="button"
									variant={coordinateInputMode === "manual" ? "secondary" : "ghost"}
									size="sm"
									aria-pressed={coordinateInputMode === "manual"}
									disabled={currentStep !== 0}
									onClick={() => updateCoordinateInputMode("manual")}
								>
									Manual coordinates
								</Button>
							</div>

							{coordinateInputMode === "search" ? (
								<div className={styles.locationSearch}>
									<TextField
										label="Location search"
										value={locationSearch}
										placeholder="Taipei 101"
										disabled={currentStep !== 0 || geocodeStatus === "searching"}
										error={geocodeStatus === "error" ? geocodeError : undefined}
										onChange={event => updateLocationSearch(event.currentTarget.value)}
									/>
									<Button type="button" variant="outline" disabled={currentStep !== 0 || !canSearchLocation} onClick={() => void searchLocation()}>
										{geocodeStatus === "searching" ? "Searching" : "Search"}
									</Button>
								</div>
							) : (
								<div className={styles.manualLocationFields}>
									<TextField label="Location name" value={locationName} placeholder="Taipei, Taiwan" disabled={currentStep !== 0} onChange={event => setLocationName(event.currentTarget.value)} />
									<div className={styles.coordinateGrid}>
										<TextField
											label="Latitude"
											type="number"
											inputMode="decimal"
											step="any"
											value={latitudeInput}
											placeholder="25.033964"
											disabled={currentStep !== 0}
											error={visibleLatitudeError || undefined}
											onChange={event => setLatitudeInput(event.currentTarget.value)}
										/>
										<TextField
											label="Longitude"
											type="number"
											inputMode="decimal"
											step="any"
											value={longitudeInput}
											placeholder="121.564468"
											disabled={currentStep !== 0}
											error={visibleLongitudeError || undefined}
											onChange={event => setLongitudeInput(event.currentTarget.value)}
										/>
									</div>
								</div>
							)}

							{selectedCoordinates ? (
								<LocationPreviewMap coordinates={selectedCoordinates} locationName={previewTitle} probeName={probeName.trim() || previewTitle} />
							) : (
								<p className={styles.locationStatus} aria-live="polite">
									{coordinateInputMode === "search" ? "Search for a place before continuing." : "Enter valid decimal coordinates before continuing."}
								</p>
							)}

							<div className={styles.actions}>
								<Button type="submit" disabled={!canCreate || currentStep !== 0 || createProbeMutation.isPending}>
									{createProbeMutation.isPending ? "Creating probe" : "Continue to install"}
								</Button>
								<p className={styles.hint}>Use a stable hostname-style label so results are easy to scan later.</p>
							</div>
						</form>

						<section className={classNames("ns-scrollbar", styles.workflowPanel)} aria-hidden={currentStep !== 1}>
							<div className={styles.stepCopy}>
								<Badge tone={heartbeatReceived ? "success" : "warning"}>{heartbeatReceived ? "Probe detected" : "Listening"}</Badge>
								<h3>Install the probe</h3>
								<p>Run the controller-served installer on the host. The wizard polls the controller and only completes after the probe heartbeat is recorded by the API.</p>
							</div>

							<div className={classNames("ns-cut-frame", styles.registrationBlock)}>
								<div className={styles.tokenLine}>
									<span>Registration token</span>
									<strong>{registrationSecret || "-"}</strong>
								</div>
								<div className={styles.assetLinks}>
									<Button asChild variant="outline" size="sm">
										<a href={installerUrl}>Installer</a>
									</Button>
									<Button asChild variant="outline" size="sm">
										<a href={binaryUrl}>Linux binary</a>
									</Button>
									<Button asChild variant="ghost" size="sm">
										<a href={uninstallerUrl}>Uninstaller</a>
									</Button>
								</div>
								<Terminal
									title="install command"
									meta={
										<Button type="button" variant="ghost" size="sm" className={styles.copyCommandButton} disabled={!installCommand} onClick={() => void copyInstallCommand()}>
											{installCommandCopied ? "Copied" : "Copy to clipboard"}
										</Button>
									}
								>
									{installCommand}
								</Terminal>
							</div>

							<div className={classNames("ns-cut-frame", styles.detectCard)}>
								<Badge tone={heartbeatReceived ? "success" : "warning"}>{heartbeatReceived ? "Heartbeat received" : "Listening for heartbeat"}</Badge>
								<strong>{heartbeatReceived ? `${probeName.trim()} is online` : "Waiting for install to finish"}</strong>
								<p>
									{heartbeatReceived
										? "The controller API reports a runtime heartbeat for this probe."
										: createdProbeQuery.isError
											? "The controller could not confirm heartbeat status yet. Polling will continue."
											: "Waiting for the first signed probe heartbeat from the controller API."}
								</p>
							</div>

							<div className={styles.actions}>
								<Button type="button" variant="ghost" disabled={currentStep !== 1} onClick={() => setCurrentStep(0)}>
									Back
								</Button>
								<Button type="button" disabled={!heartbeatReceived || currentStep !== 1} onClick={closeDrawer}>
									Finish
								</Button>
							</div>
						</section>
					</div>
				</div>
			</aside>
		</div>
	);
}
