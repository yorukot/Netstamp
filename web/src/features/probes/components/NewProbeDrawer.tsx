import { pathForRoute } from "@/routes/routePaths";
import { installAssetPaths, installAssetUrl, probeInstallCommand } from "@/shared/api/installAssets";
import { useCreateProjectProbeMutation } from "@/shared/api/mutations";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Terminal, TextField } from "@netstamp/ui";
import { type FormEvent, type MouseEvent, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./NewProbeDrawer.module.css";
import { ProbeWizardTimeline } from "./ProbeWizardTimeline";
import { TagPicker } from "./TagPicker";

const defaultProbeTags = ["Edge", "Home", "VPS", "Bare metal", "IPv6", "Web3", "Lab"];
const drawerCloseDurationMs = 180;
const installDetectionDurationMs = 1600;
const createProbeSteps = [
	{ number: "01", title: "Name", copy: "Probe identity" },
	{ number: "02", title: "Install", copy: "Run command" },
	{ number: "03", title: "Details", copy: "Optional metadata" }
];

export function NewProbeDrawer() {
	const navigate = useNavigate();
	const { projectRef } = useCurrentProject();
	const closeTimeoutRef = useRef<number | null>(null);
	const detectTimeoutRef = useRef<number | null>(null);
	const [closing, setClosing] = useState(false);
	const [currentStep, setCurrentStep] = useState(0);
	const [installStatus, setInstallStatus] = useState<"idle" | "detecting" | "detected">("idle");
	const [probeName, setProbeName] = useState("");
	const [probeLocation, setProbeLocation] = useState("");
	const [asn, setAsn] = useState("");
	const [tagOptions, setTagOptions] = useState(defaultProbeTags);
	const [selectedTags, setSelectedTags] = useState(["Edge"]);
	const [newTag, setNewTag] = useState("");
	const [registrationSecret, setRegistrationSecret] = useState("");
	const [registeredProbeId, setRegisteredProbeId] = useState("");
	const createProbeMutation = useCreateProjectProbeMutation(projectRef);
	const canCreate = probeName.trim().length > 0 && Boolean(projectRef);
	const token = registrationSecret || "waiting-for-controller";
	const installerUrl = installAssetUrl(installAssetPaths.agentInstaller);
	const uninstallerUrl = installAssetUrl(installAssetPaths.agentUninstaller);
	const binaryUrl = installAssetUrl(installAssetPaths.linuxAmd64Binary);
	const installCommand =
		registeredProbeId && registrationSecret
			? probeInstallCommand({ probeId: registeredProbeId, probeSecret: registrationSecret })
			: `curl -fsSL '${installerUrl}' | sudo sh -s -- --probe-id waiting-for-controller --probe-secret ${token}`;

	useEffect(() => {
		return () => {
			if (closeTimeoutRef.current) {
				window.clearTimeout(closeTimeoutRef.current);
			}

			if (detectTimeoutRef.current) {
				window.clearTimeout(detectTimeoutRef.current);
			}
		};
	}, []);

	useEffect(() => {
		if (currentStep !== 1 || installStatus !== "detecting") {
			return undefined;
		}

		detectTimeoutRef.current = window.setTimeout(() => {
			setInstallStatus("detected");
			detectTimeoutRef.current = null;
		}, installDetectionDurationMs);

		return () => {
			if (detectTimeoutRef.current) {
				window.clearTimeout(detectTimeoutRef.current);
				detectTimeoutRef.current = null;
			}
		};
	}, [currentStep, installStatus]);

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

	function toggleTag(tag: string) {
		setSelectedTags(current => (current.includes(tag) ? current.filter(value => value !== tag) : [...current, tag]));
	}

	function addTag() {
		const trimmedTag = newTag.trim();

		if (!trimmedTag) {
			return;
		}

		setTagOptions(current => (current.includes(trimmedTag) ? current : [...current, trimmedTag]));
		setSelectedTags(current => (current.includes(trimmedTag) ? current : [...current, trimmedTag]));
		setNewTag("");
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
						<p>Name the probe, install it on a host, then optionally annotate it with network metadata.</p>
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
								<Badge tone={installStatus === "detected" ? "success" : "warning"}>{installStatus === "detected" ? "Probe detected" : "Auto detecting"}</Badge>
								<h3>Install the probe</h3>
								<p>Run the controller-served installer on the host. The wizard watches for the first heartbeat and unlocks metadata when the probe is detected.</p>
							</div>

							<div className={classNames("ns-cut-frame", styles.registrationBlock)}>
								<div className={styles.tokenLine}>
									<span>Registration token</span>
									<strong>{token}</strong>
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
								<Badge tone={installStatus === "detected" ? "success" : "warning"}>{installStatus === "detected" ? "Heartbeat received" : "Listening for heartbeat"}</Badge>
								<strong>{installStatus === "detected" ? `${probeName.trim()} is online` : "Waiting for install to finish"}</strong>
								<p>{installStatus === "detected" ? "The controller accepted the first signed result stream." : "Waiting for the first signed probe heartbeat."}</p>
							</div>

							<div className={styles.actions}>
								<Button type="button" variant="ghost" disabled={currentStep !== 1} onClick={() => setCurrentStep(0)}>
									Back
								</Button>
								<Button type="button" disabled={installStatus !== "detected" || currentStep !== 1} onClick={() => setCurrentStep(2)}>
									Continue to details
								</Button>
							</div>
						</section>

						<section className={styles.workflowPanel} aria-hidden={currentStep !== 2}>
							<div className={styles.stepCopy}>
								<Badge tone="accent">Step 03</Badge>
								<h3>Add optional metadata</h3>
								<p>Customize location, AS, and tags or we can guess it by ourselves.</p>
							</div>

							<div className={styles.formGrid}>
								<TextField
									label="Location (optional)"
									value={probeLocation}
									placeholder="Taipei, Taiwan"
									disabled={currentStep !== 2}
									onChange={event => setProbeLocation(event.currentTarget.value)}
								/>
								<TextField label="AS (optional)" value={asn} placeholder="AS3462" disabled={currentStep !== 2} onChange={event => setAsn(event.currentTarget.value)} />
							</div>

							<TagPicker tagOptions={tagOptions} selectedTags={selectedTags} newTag={newTag} disabled={currentStep !== 2} onToggleTag={toggleTag} onNewTagChange={setNewTag} onAddTag={addTag} />

							<div className={styles.actions}>
								<Button type="button" variant="ghost" disabled={currentStep !== 2} onClick={() => setCurrentStep(1)}>
									Back
								</Button>
								<Button type="button" disabled={currentStep !== 2} onClick={closeDrawer}>
									Save details
								</Button>
								<Button type="button" variant="outline" disabled={currentStep !== 2} onClick={closeDrawer}>
									Skip
								</Button>
							</div>
						</section>
					</div>
				</div>
			</aside>
		</div>
	);
}
