import { useAuth } from "@/features/auth/hooks/useAuth";
import { type Navigate } from "@/routes/routeTypes";
import { apiProblemCode } from "@/shared/api/client";
import { useAcceptProjectInviteMutation, useCreateProjectInviteForRefMutation, useCreateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite } from "@/shared/api/types";
import { useProjectSelection } from "@/shared/api/useCurrentProject";
import { appFeatures } from "@/shared/config/features";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, Input, PageShell, Spinner } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent, KeyboardEvent as ReactKeyboardEvent } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { Helmet } from "react-helmet-async";
import styles from "./OnboardingPage.module.css";

interface OnboardingPageProps {
	navigate: Navigate;
}

interface ScriptStep {
	prompt: string;
	text: string;
	autoAdvanceAfter?: number;
}

const typeDelayMs = 34;
const maxProjectSlugLength = 64;
const maxSlugAttempts = 20;
const randomSlugTokenLength = 6;
const randomSlugTokenAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789";

function slugifyProjectName(name: string) {
	return (
		name
			.toLowerCase()
			.trim()
			.replace(/[^a-z0-9]+/g, "-")
			.replace(/^-|-$/g, "") || "yoru-project"
	);
}

function projectSlugCandidate(baseSlug: string, attempt: number) {
	if (attempt === 0) {
		return baseSlug.slice(0, maxProjectSlugLength);
	}

	const suffix = `-${randomSlugToken(randomSlugTokenLength)}`;
	const baseLength = maxProjectSlugLength - suffix.length;
	return `${baseSlug.slice(0, Math.max(1, baseLength)).replace(/-$/g, "")}${suffix}`;
}

function randomSlugToken(length: number) {
	const values = new Uint8Array(length);
	globalThis.crypto.getRandomValues(values);

	return Array.from(values, value => randomSlugTokenAlphabet[value % randomSlugTokenAlphabet.length]).join("");
}

function isProjectSlugConflict(error: unknown) {
	return apiProblemCode(error) === "PROJECT_SLUG_ALREADY_EXISTS";
}

function projectRefFromInvite(invite: ApiProjectInvite) {
	return invite.project.slug || invite.project.id;
}

function roleLabel(role: string) {
	return role.charAt(0).toUpperCase() + role.slice(1);
}

function inviteSummary(invites: ApiProjectInvite[]) {
	if (invites.length === 1) {
		const invite = invites[0];
		return `You have pending ${roleLabel(invite.role)} access to ${invite.project.name}.`;
	}

	return `You have ${invites.length} pending project invites.`;
}

function shouldCreateProjectFromDecision(value: string) {
	const normalized = value.trim().toLowerCase();

	if (normalized === "" || normalized === "y" || normalized === "yes") {
		return true;
	}

	if (normalized === "n" || normalized === "no") {
		return false;
	}

	return null;
}

export function OnboardingPage({ navigate }: OnboardingPageProps) {
	const { session, loading, submitting, logout } = useAuth();
	const createProjectMutation = useCreateProjectMutation({ suppressGlobalErrorToast: true });
	const createInviteMutation = useCreateProjectInviteForRefMutation({ suppressGlobalErrorToast: true });
	const acceptInviteMutation = useAcceptProjectInviteMutation();
	const pendingInvitesQuery = useQuery(projectQueries.currentUserInvites());
	const { setSelectedProjectRef } = useProjectSelection();
	const [activeStep, setActiveStep] = useState(0);
	const [typedText, setTypedText] = useState("");
	const [createProjectDecision, setCreateProjectDecision] = useState("");
	const [projectName, setProjectName] = useState("");
	const [invites, setInvites] = useState([""]);
	const [acceptedInviteProject, setAcceptedInviteProject] = useState<{ name: string; ref: string } | null>(null);
	const [createdProject, setCreatedProject] = useState("");
	const [createdProjectRef, setCreatedProjectRef] = useState("");
	const [creatingProject, setCreatingProject] = useState(false);
	const [resolvingInvites, setResolvingInvites] = useState(false);
	const decisionInputRef = useRef<HTMLInputElement | null>(null);
	const projectInputRef = useRef<HTMLInputElement | null>(null);
	const inviteRefs = useRef<Array<HTMLInputElement | null>>([]);
	const displayName = session?.user.name.trim() || "there";
	const pendingInvites = pendingInvitesQuery.data?.invites ?? [];
	const shouldAskCreateProject = pendingInvites.length > 0 && !acceptedInviteProject;
	const projectFlowReady = appFeatures.projectCreation && !pendingInvitesQuery.isPending && !shouldAskCreateProject;
	const acceptingInvites = resolvingInvites || acceptInviteMutation.isPending;
	const scriptSteps = useMemo<ScriptStep[]>(
		() => [
			{ prompt: "netstamp", text: `Nice to meet you, ${displayName}`, autoAdvanceAfter: 180 },
			{ prompt: "netstamp", text: acceptedInviteProject ? "Let's create a new project too." : "Let's create your first project.", autoAdvanceAfter: 760 },
			{ prompt: "project", text: "How should we call your project?" },
			{ prompt: "members", text: "Invite project members?" }
		],
		[acceptedInviteProject, displayName]
	);

	const activeScript = scriptSteps[activeStep];
	const projectPromptReady = activeStep > 2 || (activeStep === 2 && typedText.length === scriptSteps[2].text.length);
	const membersPromptReady = activeStep > 3 || (activeStep === 3 && typedText.length === scriptSteps[3].text.length);

	useEffect(() => {
		if (!activeScript || loading || !projectFlowReady) {
			return undefined;
		}

		if (typedText.length < activeScript.text.length) {
			const timeout = window.setTimeout(() => {
				setTypedText(activeScript.text.slice(0, typedText.length + 1));
			}, typeDelayMs);

			return () => window.clearTimeout(timeout);
		}

		if (typeof activeScript.autoAdvanceAfter === "number") {
			const timeout = window.setTimeout(() => {
				setActiveStep(current => Math.min(current + 1, scriptSteps.length - 1));
				setTypedText("");
			}, activeScript.autoAdvanceAfter);

			return () => window.clearTimeout(timeout);
		}

		return undefined;
	}, [activeScript, loading, projectFlowReady, scriptSteps.length, typedText]);

	useEffect(() => {
		if (!shouldAskCreateProject) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => decisionInputRef.current?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [shouldAskCreateProject]);

	useEffect(() => {
		if (!projectFlowReady || !projectPromptReady || activeStep !== 2) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => projectInputRef.current?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, projectFlowReady, projectPromptReady]);

	useEffect(() => {
		if (!projectFlowReady || !membersPromptReady || activeStep !== 3) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => inviteRefs.current[0]?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, membersPromptReady, projectFlowReady]);

	function focusInvite(index: number) {
		window.requestAnimationFrame(() => inviteRefs.current[index]?.focus());
	}

	function updateInvite(index: number, value: string) {
		setInvites(current => current.map((invite, currentIndex) => (currentIndex === index ? value : invite)));
	}

	function addInvite() {
		const nextIndex = invites.length;
		setInvites(current => [...current, ""]);
		focusInvite(nextIndex);
	}

	function removeInvite(index: number) {
		setInvites(current => current.filter((_, currentIndex) => currentIndex !== index));
	}

	function advanceToMembers() {
		if (!projectPromptReady || activeStep !== 2) {
			return;
		}

		setActiveStep(3);
		setTypedText("");
	}

	function handleProjectKeyDown(event: ReactKeyboardEvent<HTMLInputElement>) {
		if (event.key !== "Enter") {
			return;
		}

		event.preventDefault();
		advanceToMembers();
	}

	function handleInviteKeyDown(event: ReactKeyboardEvent<HTMLInputElement>, index: number) {
		if (event.key === "Backspace" && invites[index] === "" && invites.length > 1) {
			event.preventDefault();
			setInvites(current => current.filter((_, currentIndex) => currentIndex !== index));
			focusInvite(Math.max(0, index - 1));
			return;
		}

		if (event.key !== "Enter") {
			return;
		}

		event.preventDefault();
		const nextIndex = index + 1;

		if (index === invites.length - 1) {
			setInvites(current => [...current, ""]);
		}

		focusInvite(nextIndex);
	}

	function normalizedInviteEmails() {
		return Array.from(new Set(invites.map(invite => invite.trim()).filter(Boolean)));
	}

	async function acceptPendingInvites() {
		setResolvingInvites(true);

		try {
			const inviteResults = await Promise.allSettled(pendingInvites.map(invite => acceptInviteMutation.mutateAsync(invite.id)));
			const acceptedInvites = inviteResults.flatMap(result => (result.status === "fulfilled" ? [result.value.invite] : []));
			const failedInviteCount = inviteResults.length - acceptedInvites.length;

			if (failedInviteCount) {
				pushErrorToast(`${failedInviteCount} project invite${failedInviteCount === 1 ? "" : "s"} could not be accepted.`);
			}

			const acceptedInvite = acceptedInvites[0];

			if (!acceptedInvite) {
				throw new Error("Project invite could not be accepted.");
			}

			const projectRef = projectRefFromInvite(acceptedInvite);
			const project = {
				name: acceptedInvite.project.name,
				ref: projectRef
			};

			setSelectedProjectRef(projectRef);
			setAcceptedInviteProject(project);

			return project;
		} finally {
			setResolvingInvites(false);
		}
	}

	async function handleCreateProjectDecisionSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		const shouldCreateProject = shouldCreateProjectFromDecision(createProjectDecision);

		if (shouldCreateProject === null) {
			pushErrorToast("Answer y or n.");
			return;
		}

		try {
			const project = await acceptPendingInvites();

			if (!shouldCreateProject || !appFeatures.projectCreation) {
				navigate("dashboard", { projectRef: project.ref });
				return;
			}

			setCreateProjectDecision("");
			setActiveStep(0);
			setTypedText("");
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, "Project invite could not be accepted."));
		}
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!membersPromptReady) {
			return;
		}

		const normalizedProjectName = projectName.trim() || "Yoru Labs";
		const baseSlug = slugifyProjectName(normalizedProjectName);

		setCreatingProject(true);
		try {
			for (let attempt = 0; attempt < maxSlugAttempts; attempt += 1) {
				try {
					const data = await createProjectMutation.mutateAsync({
						name: normalizedProjectName,
						slug: projectSlugCandidate(baseSlug, attempt)
					});
					const projectRef = data.project.slug || data.project.id;
					const inviteEmails = normalizedInviteEmails();

					if (inviteEmails.length) {
						const inviteResults = await Promise.allSettled(inviteEmails.map(email => createInviteMutation.mutateAsync({ projectRef, body: { email, role: "viewer" } })));
						const failedInviteCount = inviteResults.filter(result => result.status === "rejected").length;

						if (failedInviteCount) {
							pushErrorToast(`${failedInviteCount} project invite${failedInviteCount === 1 ? "" : "s"} could not be sent.`);
						}
					}

					setSelectedProjectRef(projectRef);
					setCreatedProject(normalizedProjectName);
					setCreatedProjectRef(projectRef);
					return;
				} catch (error) {
					if (!isProjectSlugConflict(error)) {
						throw error;
					}
				}
			}

			pushErrorToast("Project slug is already in use. Try a different project name.");
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, "Project could not be created."));
		} finally {
			setCreatingProject(false);
		}
	}

	if (loading || pendingInvitesQuery.isPending) {
		return (
			<PageShell variant="constellation" center className={styles.shell}>
				<Spinner label="Loading onboarding" layout="panel" size="lg" />
			</PageShell>
		);
	}

	if (!appFeatures.projectCreation && !shouldAskCreateProject) {
		return (
			<PageShell variant="constellation" center className={styles.shell}>
				<Helmet>
					<title>No Project Access - Netstamp</title>
				</Helmet>

				<section className={styles.console} aria-label="Project access console">
					<div className={styles.consoleBar}>
						<span aria-hidden="true" />
						<span aria-hidden="true" />
						<span aria-hidden="true" />
						<strong>yoru://project-access</strong>
					</div>
					<div className={styles.consoleBody}>
						<div className={styles.successView}>
							<ScriptLine prompt="access" text="No projects are assigned to this account." />
							<p>Use an invited demo account or ask an operator for project access.</p>
							<Button variant="plain" className={styles.tuiButton} type="button" onClick={logout}>
								[ log out ]
							</Button>
						</div>
					</div>
				</section>
			</PageShell>
		);
	}

	return (
		<PageShell variant="constellation" center className={styles.shell}>
			<Helmet>
				<title>{shouldAskCreateProject ? "Welcome" : "Create Project"} - Netstamp</title>
			</Helmet>

			<section className={styles.console} aria-label="First contact onboarding console">
				<div className={styles.consoleBar}>
					<span aria-hidden="true" />
					<span aria-hidden="true" />
					<span aria-hidden="true" />
					<strong>yoru://first-contact</strong>
				</div>
				<div className={styles.consoleBody}>
					{shouldAskCreateProject ? (
						<>
							<div className={styles.scriptLog}>
								<ScriptLine prompt="netstamp" text={`Nice to meet you, ${displayName}`} />
								<ScriptLine prompt="invite" text={inviteSummary(pendingInvites)} />
								<ScriptLine prompt="project" text={appFeatures.projectCreation ? "Create a project? [Y/n]" : "Open invited project? [Y/n]"} />
							</div>

							<form className={styles.tuiForm} onSubmit={handleCreateProjectDecisionSubmit}>
								<label className={styles.answerRow}>
									<span className={styles.answerPrompt}>answer</span>
									<Input
										variant="bare"
										ref={decisionInputRef}
										name="create-project"
										type="text"
										value={createProjectDecision}
										placeholder="Y"
										onChange={event => setCreateProjectDecision(event.currentTarget.value)}
										autoComplete="off"
										disabled={acceptingInvites}
									/>
									<small>Press Enter for yes. Type n to use invited project access only.</small>
								</label>

								<Button variant="plain" className={styles.tuiButton} type="submit" disabled={acceptingInvites}>
									{acceptingInvites ? "[ accepting invites… ]" : "[ continue ]"}
								</Button>
							</form>
						</>
					) : createdProject ? (
						<div className={styles.successView} aria-live="polite">
							<ScriptLine prompt="success" text={`Project ${createdProject} created.`} />
							<p>Nice, let's bring {createdProject} online. Next we will open the probe fleet and start the new probe wizard.</p>
							<Button variant="plain" className={styles.tuiButton} type="button" onClick={() => navigate("newProbe", { projectRef: createdProjectRef })}>
								[ open probe fleet / create probe ]
							</Button>
						</div>
					) : (
						<>
							<div className={styles.scriptLog}>
								{scriptSteps.slice(0, Math.min(activeStep, 3)).map(step => (
									<ScriptLine key={step.prompt + step.text} prompt={step.prompt} text={step.text} />
								))}
								{activeStep < 3 && activeScript ? <ScriptLine prompt={activeScript.prompt} text={typedText} cursor={typedText.length < activeScript.text.length} /> : null}
							</div>

							<form className={styles.tuiForm} onSubmit={handleSubmit}>
								{projectPromptReady ? (
									<label className={styles.answerRow}>
										<span className={styles.answerPrompt}>answer</span>
										<Input
											variant="bare"
											ref={projectInputRef}
											name="project"
											type="text"
											value={projectName}
											placeholder="Yoru Labs"
											onChange={event => setProjectName(event.currentTarget.value)}
											onKeyDown={handleProjectKeyDown}
											autoComplete="off"
										/>
										{activeStep === 2 ? <small>Press Enter to continue.</small> : null}
									</label>
								) : null}

								{activeStep >= 3 ? (
									<ScriptLine prompt="members" text={activeStep === 3 ? typedText : scriptSteps[3].text} cursor={activeStep === 3 && typedText.length < scriptSteps[3].text.length} />
								) : null}

								{membersPromptReady ? (
									<div className={styles.inviteSection}>
										<div className={styles.inviteHeader}>
											<p>Press Enter for next member email. Backspace on an empty row deletes it.</p>
											<Button variant="plain" className={styles.tuiMiniButton} type="button" onClick={addInvite}>
												+ add
											</Button>
										</div>

										<div className={styles.inviteList}>
											{invites.map((invite, index) => (
												<div className={styles.inviteRow} key={index}>
													<label className={styles.answerRow}>
														<span className={styles.answerPrompt}>{String(index + 1).padStart(2, "0")}</span>
														<Input
															variant="bare"
															ref={element => {
																inviteRefs.current[index] = element;
															}}
															name={`invite-${index}`}
															type="email"
															autoComplete="email"
															value={invite}
															placeholder="member@example.com"
															onChange={event => updateInvite(index, event.currentTarget.value)}
															onKeyDown={event => handleInviteKeyDown(event, index)}
														/>
													</label>
													<Button variant="plain" className={styles.tuiMiniButton} type="button" onClick={() => removeInvite(index)}>
														delete
													</Button>
												</div>
											))}
										</div>

										<Button variant="plain" className={styles.tuiButton} type="submit" disabled={submitting || creatingProject}>
											{submitting || creatingProject ? "[ creating project… ]" : "[ create project ]"}
										</Button>
									</div>
								) : null}
							</form>
						</>
					)}
				</div>
			</section>
		</PageShell>
	);
}

interface ScriptLineProps {
	prompt: string;
	text: string;
	cursor?: boolean;
}

function ScriptLine({ prompt, text, cursor = false }: ScriptLineProps) {
	return (
		<div className={styles.scriptLine}>
			<span className={styles.scriptPrompt}>{prompt}</span>
			<span className={styles.scriptText}>
				{text}
				{cursor ? <span className={styles.cursor} aria-hidden="true" /> : null}
			</span>
		</div>
	);
}
