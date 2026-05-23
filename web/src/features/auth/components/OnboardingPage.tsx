import { type Navigate } from "@/routes/routeTypes";
import { ApiError } from "@/shared/api/client";
import { createProject as createProjectRequest } from "@/shared/api/mutations";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { Button, Input, PageShell } from "@netstamp/ui";
import { useQueryClient } from "@tanstack/react-query";
import type { FormEvent, KeyboardEvent as ReactKeyboardEvent } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { Helmet } from "react-helmet-async";
import { useAuth } from "../hooks/useAuth";
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
	return error instanceof ApiError && error.status === 409 && /project slug already exists/i.test(error.message);
}

export function OnboardingPage({ navigate }: OnboardingPageProps) {
	const { session, loading, submitting } = useAuth();
	const queryClient = useQueryClient();
	const [activeStep, setActiveStep] = useState(0);
	const [typedText, setTypedText] = useState("");
	const [projectName, setProjectName] = useState("");
	const [invites, setInvites] = useState([""]);
	const [createdProject, setCreatedProject] = useState("");
	const [creatingProject, setCreatingProject] = useState(false);
	const projectInputRef = useRef<HTMLInputElement | null>(null);
	const inviteRefs = useRef<Array<HTMLInputElement | null>>([]);
	const displayName = session?.user.name.trim() || "there";
	const scriptSteps = useMemo<ScriptStep[]>(
		() => [
			{ prompt: "netstamp", text: `Nice to meet you, ${displayName}`, autoAdvanceAfter: 180 },
			{ prompt: "netstamp", text: "Let's create our first project!", autoAdvanceAfter: 760 },
			{ prompt: "project", text: "How should we call your project?" },
			{ prompt: "members", text: "Invite project members?" }
		],
		[displayName]
	);

	const activeScript = scriptSteps[activeStep];
	const projectPromptReady = activeStep > 2 || (activeStep === 2 && typedText.length === scriptSteps[2].text.length);
	const membersPromptReady = activeStep > 3 || (activeStep === 3 && typedText.length === scriptSteps[3].text.length);

	useEffect(() => {
		if (!activeScript || loading) {
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
	}, [activeScript, loading, scriptSteps.length, typedText]);

	useEffect(() => {
		if (!projectPromptReady || activeStep !== 2) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => projectInputRef.current?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, projectPromptReady]);

	useEffect(() => {
		if (!membersPromptReady || activeStep !== 3) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => inviteRefs.current[0]?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, membersPromptReady]);

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
					await createProjectRequest({
						name: normalizedProjectName,
						slug: projectSlugCandidate(baseSlug, attempt)
					});
					await queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.list() });
					setCreatedProject(normalizedProjectName);
					return;
				} catch (error) {
					if (!isProjectSlugConflict(error)) {
						throw error;
					}
				}
			}

			pushErrorToast("Project slug is already in use. Try a different project name.");
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Project could not be created.");
		} finally {
			setCreatingProject(false);
		}
	}

	if (loading) {
		return null;
	}

	return (
		<PageShell variant="constellation" center className={styles.shell}>
			<Helmet>
				<title>Create Project - Netstamp</title>
			</Helmet>

			<section className={classNames("ns-cut-frame", styles.console)} aria-label="First contact onboarding console">
				<div className={styles.consoleBar}>
					<span aria-hidden="true" />
					<span aria-hidden="true" />
					<span aria-hidden="true" />
					<strong>yoru://first-contact</strong>
				</div>
				<div className={styles.consoleBody}>
					<div className={styles.scanline} aria-hidden="true" />

					{createdProject ? (
						<div className={styles.successView} aria-live="polite">
							<ScriptLine prompt="success" text={`Project ${createdProject} created.`} />
							<p>Nice, let's bring {createdProject} online. Next we will open the probe fleet and start the new probe wizard.</p>
							<Button variant="plain" className={styles.tuiButton} type="button" onClick={() => navigate("newProbe")}>
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
											{submitting || creatingProject ? "[ creating project... ]" : "[ create project ]"}
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
