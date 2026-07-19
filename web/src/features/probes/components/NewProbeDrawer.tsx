import {
	coordinateInputError,
	formatCoordinate,
	LocationSearchError,
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
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, CodeBlock, SegmentedControl, TextField } from "@netstamp/ui";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { LocationPreviewMap } from "./LocationPreviewMap";
import styles from "./NewProbeDrawer.module.css";
import { ProbeWizardTimeline } from "./ProbeWizardTimeline";

export function NewProbeDrawer() {
	const { t } = useTranslation(["probes", "common"]);
	const navigate = useNavigate();
	const { projectRef } = useCurrentProject();
	const queryClient = useQueryClient();
	const geocodeAbortRef = useRef<AbortController | null>(null);
	const [currentStep, setCurrentStep] = useState(0);
	const [installStatus, setInstallStatus] = useState<"idle" | "detecting">("idle");
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
	const latitudeError = coordinateInputMode === "manual" ? coordinateInputError(latitudeInput, -90, 90) : null;
	const longitudeError = coordinateInputMode === "manual" ? coordinateInputError(longitudeInput, -180, 180) : null;
	const visibleLatitudeError = latitudeInput.trim() && latitudeError ? t(`location.${latitudeError}`, { label: t("location.latitude"), min: -90, max: 90 }) : "";
	const visibleLongitudeError = longitudeInput.trim() && longitudeError ? t(`location.${longitudeError}`, { label: t("location.longitude"), min: -180, max: 180 }) : "";
	const searchCoordinatesReady = coordinateInputMode === "search" && geocodeStatus === "resolved" && latitude !== null && longitude !== null;
	const manualCoordinatesReady = coordinateInputMode === "manual" && latitude !== null && longitude !== null && !latitudeError && !longitudeError;
	const coordinatesReady = searchCoordinatesReady || manualCoordinatesReady;
	const selectedCoordinates: ProbeCoordinates | null = coordinatesReady && latitude !== null && longitude !== null ? { latitude, longitude } : null;
	const canSearchLocation = coordinateInputMode === "search" && locationSearch.trim().length > 0 && geocodeStatus !== "searching";
	const canCreate = probeName.trim().length > 0 && Boolean(projectRef) && Boolean(selectedCoordinates);
	const previewTitle = locationName.trim() || t("location.manual");
	const createProbeSteps = [
		{ number: "01", title: t("wizard.nameStep"), copy: t("wizard.nameStepCopy") },
		{ number: "02", title: t("wizard.installStep"), copy: t("wizard.installStepCopy") }
	];
	const coordinateModeOptions: Array<{ value: CoordinateInputMode; label: string }> = [
		{ value: "search", label: t("location.searchName") },
		{ value: "manual", label: t("location.manualCoordinates") }
	];
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);
	const uninstallerUrl = installAssetUrl(installAssetPaths.agentUninstaller);
	const binaryUrl = installAssetUrl(installAssetPaths.linuxAmd64Binary);
	const installCommand = registeredProbeId && registrationSecret ? probeInstallCommand({ probeId: registeredProbeId, probeSecret: registrationSecret }) : "";

	useEffect(() => {
		return () => {
			geocodeAbortRef.current?.abort();
		};
	}, []);

	useEffect(() => {
		if (heartbeatReceived) {
			void queryClient.invalidateQueries({ queryKey: projectQueries.probes(projectRef || "").queryKey });
		}
	}, [heartbeatReceived, projectRef, queryClient]);

	function closeDrawer() {
		navigate(pathForRoute("probes", { projectRef }));
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
			setGeocodeError(t(`location.${error instanceof LocationSearchError ? error.code : "searchFailed"}`));
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

	return (
		<EditorDrawer open title={t("wizard.title")} ariaLabel={t("wizard.aria")} contentClassName={styles.drawerContent} onClose={closeDrawer}>
			<p className={styles.drawerIntro}>{t("wizard.intro")}</p>

			<ProbeWizardTimeline steps={createProbeSteps} currentStep={currentStep} />

			<div className={styles.workflowViewport}>
				<div className={styles.workflowTrack} style={{ transform: `translateX(-${currentStep * 100}%)` }}>
					<form className={classNames("ns-scrollbar", styles.workflowPanel)} aria-hidden={currentStep !== 0} onSubmit={handleNameSubmit}>
						<div className={styles.stepCopy}>
							<Badge tone="accent">{t("wizard.step01")}</Badge>
							<h3>{t("wizard.identityTitle")}</h3>
							<p>{t("wizard.identityDescription")}</p>
						</div>

						<TextField
							label={t("wizard.name")}
							value={probeName}
							placeholder={t("wizard.namePlaceholder")}
							required
							disabled={currentStep !== 0}
							onChange={event => updateProbeName(event.currentTarget.value)}
						/>

						<SegmentedControl
							className={styles.locationMode}
							size="sm"
							ariaLabel={t("location.coordinateMode")}
							value={coordinateInputMode}
							options={coordinateModeOptions.map(option => ({ ...option, disabled: currentStep !== 0 }))}
							onValueChange={nextMode => updateCoordinateInputMode(nextMode as CoordinateInputMode)}
						/>

						{coordinateInputMode === "search" ? (
							<div className={styles.locationSearch}>
								<TextField
									label={t("location.search")}
									value={locationSearch}
									placeholder={t("location.searchPlaceholder")}
									disabled={currentStep !== 0 || geocodeStatus === "searching"}
									error={geocodeStatus === "error" ? geocodeError : undefined}
									onChange={event => updateLocationSearch(event.currentTarget.value)}
								/>
								<Button type="button" variant="outline" disabled={currentStep !== 0 || !canSearchLocation} onClick={() => void searchLocation()}>
									{geocodeStatus === "searching" ? t("location.searching") : t("location.searchAction")}
								</Button>
							</div>
						) : (
							<div className={styles.manualLocationFields}>
								<TextField
									label={t("location.name")}
									value={locationName}
									placeholder={t("location.namePlaceholder")}
									disabled={currentStep !== 0}
									onChange={event => setLocationName(event.currentTarget.value)}
								/>
								<div className={styles.coordinateGrid}>
									<TextField
										label={t("location.latitude")}
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
										label={t("location.longitude")}
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
								{coordinateInputMode === "search" ? t("location.searchBeforeContinue") : t("location.coordinatesBeforeContinue")}
							</p>
						)}

						<div className={styles.actions}>
							<Button type="submit" disabled={!canCreate || currentStep !== 0 || createProbeMutation.isPending}>
								{createProbeMutation.isPending ? t("wizard.createPending") : t("wizard.continueInstall")}
							</Button>
							<p className={styles.hint}>{t("wizard.stableNameHint")}</p>
						</div>
					</form>

					<section className={classNames("ns-scrollbar", styles.workflowPanel)} aria-hidden={currentStep !== 1}>
						<div className={styles.stepCopy}>
							<Badge tone={heartbeatReceived ? "success" : "warning"}>{heartbeatReceived ? t("wizard.detected") : t("wizard.listening")}</Badge>
							<h3>{t("wizard.installTitle")}</h3>
							<p>{t("wizard.installDescription")}</p>
						</div>

						<div className={styles.registrationBlock}>
							<div className={styles.tokenLine}>
								<span>{t("wizard.registrationToken")}</span>
								<strong>{registrationSecret || "-"}</strong>
							</div>
							<CodeBlock title={t("wizard.installCommand")} className={styles.installCommand} copyDisabled={!installCommand}>
								{installCommand}
							</CodeBlock>
							<div className={styles.assetLinks}>
								<Button asChild variant="outline" size="sm">
									<a href={installerUrl}>{t("wizard.installer")}</a>
								</Button>
								<Button asChild variant="outline" size="sm">
									<a href={binaryUrl}>{t("wizard.linuxBinary")}</a>
								</Button>
								<Button asChild variant="ghost" size="sm">
									<a href={uninstallerUrl}>{t("wizard.uninstaller")}</a>
								</Button>
							</div>
						</div>

						<div className={styles.detectCard}>
							<Badge tone={heartbeatReceived ? "success" : "warning"}>{heartbeatReceived ? t("wizard.heartbeatReceived") : t("wizard.listeningHeartbeat")}</Badge>
							<strong>{heartbeatReceived ? t("wizard.online", { name: probeName.trim() }) : t("wizard.waitingInstall")}</strong>
							<p>{heartbeatReceived ? t("wizard.heartbeatConfirmed") : createdProbeQuery.isError ? t("wizard.heartbeatError") : t("wizard.heartbeatWaiting")}</p>
						</div>

						<div className={styles.actions}>
							<Button type="button" variant="ghost" disabled={currentStep !== 1} onClick={() => setCurrentStep(0)}>
								{t("common:actions.back")}
							</Button>
							<Button type="button" disabled={!heartbeatReceived || currentStep !== 1} onClick={closeDrawer}>
								{t("common:actions.finish")}
							</Button>
						</div>
					</section>
				</div>
			</div>
		</EditorDrawer>
	);
}
