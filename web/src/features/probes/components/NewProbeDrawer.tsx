import { pathForRoute } from "@/routes/routePaths";
import { installAssetPaths, installAssetUrl, probeInstallCommand } from "@/shared/api/installAssets";
import { useCreateProjectProbeMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Terminal, TextField } from "@netstamp/ui";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type FormEvent, type MouseEvent, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./NewProbeDrawer.module.css";
import { ProbeWizardTimeline } from "./ProbeWizardTimeline";

const drawerCloseDurationMs = 180;
const createProbeSteps = [
	{ number: "01", title: "Name", copy: "Probe identity" },
	{ number: "02", title: "Install", copy: "Run command" }
];

export function NewProbeDrawer() {
	const navigate = useNavigate();
	const { projectRef } = useCurrentProject();
	const queryClient = useQueryClient();
	const closeTimeoutRef = useRef<number | null>(null);
	const [closing, setClosing] = useState(false);
	const [currentStep, setCurrentStep] = useState(0);
	const [installStatus, setInstallStatus] = useState<"idle" | "detecting">("idle");
	const [probeName, setProbeName] = useState("");
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
	const canCreate = probeName.trim().length > 0 && Boolean(projectRef);
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);
	const uninstallerUrl = installAssetUrl(installAssetPaths.agentUninstaller);
	const binaryUrl = installAssetUrl(installAssetPaths.linuxAmd64Binary);
	const installCommand = registeredProbeId && registrationSecret ? probeInstallCommand({ probeId: registeredProbeId, probeSecret: registrationSecret }) : "";

	useEffect(() => {
		return () => {
			if (closeTimeoutRef.current) {
				window.clearTimeout(closeTimeoutRef.current);
			}
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

	function updateProbeName(value: string) {
		setProbeName(value);
		setInstallStatus("idle");
		setRegisteredProbeId("");
		setRegistrationSecret("");
	}

	function startInstallDetection() {
		setInstallStatus("detecting");
		setCurrentStep(1);
	}

	async function handleNameSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!canCreate) {
			return;
		}

		const data = await createProbeMutation.mutateAsync({ enabled: true, name: probeName.trim() });
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
						<Badge tone="accent">New probe wizard</Badge>
						<h2>Create probe</h2>
						<p>Name the probe, install it on a host, then wait for the controller to receive its first heartbeat.</p>
					</div>
					<Button type="button" variant="ghost" size="sm" onClick={closeDrawer}>
						Close
					</Button>
				</div>

				<ProbeWizardTimeline steps={createProbeSteps} currentStep={currentStep} />

				<div className={styles.workflowViewport}>
					<div className={styles.workflowTrack} style={{ transform: `translateX(-${currentStep * 100}%)` }}>
						<form className={styles.workflowPanel} aria-hidden={currentStep !== 0} onSubmit={handleNameSubmit}>
							<div className={styles.stepCopy}>
								<Badge tone="accent">Step 01</Badge>
								<h3>Enter probe name</h3>
								<p>This name is embedded in the registration command and shown in the probe fleet.</p>
							</div>

							<TextField label="Probe name" value={probeName} placeholder="taipei-home-01" required disabled={currentStep !== 0} onChange={event => updateProbeName(event.currentTarget.value)} />

							<div className={styles.actions}>
								<Button type="submit" disabled={!canCreate || currentStep !== 0 || createProbeMutation.isPending}>
									{createProbeMutation.isPending ? "Creating probe" : "Continue to install"}
								</Button>
								<p className={styles.hint}>Use a stable hostname-style label so results are easy to scan later.</p>
							</div>
						</form>

						<section className={styles.workflowPanel} aria-hidden={currentStep !== 1}>
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
								<Terminal title="install command" meta="copy to host">
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
