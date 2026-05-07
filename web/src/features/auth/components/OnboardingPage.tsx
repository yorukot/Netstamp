import { type Navigate } from "@/routes/routeTypes";
import { classNames } from "@/shared/utils/classNames";
import { Button, Input, PageShell } from "@netstamp/ui";
import type { FormEvent, KeyboardEvent as ReactKeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { Helmet } from "react-helmet-async";
import { useAuthMock } from "../hooks/useAuthMock";
import styles from "./OnboardingPage.module.css";

interface OnboardingPageProps {
	navigate: Navigate;
}

interface ScriptStep {
	prompt: string;
	text: string;
	autoAdvanceAfter?: number;
}

const scriptSteps: ScriptStep[] = [
	{ prompt: "netstamp", text: "Nice to meet you, Yoru", autoAdvanceAfter: 180 },
	{ prompt: "netstamp", text: "Let's create our first team!", autoAdvanceAfter: 760 },
	{ prompt: "team", text: "How should we call your team?" },
	{ prompt: "friends", text: "Any friends?" }
];

const typeDelayMs = 34;

export function OnboardingPage({ navigate }: OnboardingPageProps) {
	const { submitting, createTeam } = useAuthMock();
	const [activeStep, setActiveStep] = useState(0);
	const [typedText, setTypedText] = useState("");
	const [teamName, setTeamName] = useState("");
	const [invites, setInvites] = useState([""]);
	const [createdTeam, setCreatedTeam] = useState("");
	const teamInputRef = useRef<HTMLInputElement | null>(null);
	const inviteRefs = useRef<Array<HTMLInputElement | null>>([]);

	const activeScript = scriptSteps[activeStep];
	const teamPromptReady = activeStep > 2 || (activeStep === 2 && typedText.length === scriptSteps[2].text.length);
	const friendsPromptReady = activeStep > 3 || (activeStep === 3 && typedText.length === scriptSteps[3].text.length);

	useEffect(() => {
		if (!activeScript) {
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
	}, [activeScript, typedText]);

	useEffect(() => {
		if (!teamPromptReady || activeStep !== 2) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => teamInputRef.current?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, teamPromptReady]);

	useEffect(() => {
		if (!friendsPromptReady || activeStep !== 3) {
			return undefined;
		}

		const frame = window.requestAnimationFrame(() => inviteRefs.current[0]?.focus());
		return () => window.cancelAnimationFrame(frame);
	}, [activeStep, friendsPromptReady]);

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

	function advanceToFriends() {
		if (!teamPromptReady || activeStep !== 2) {
			return;
		}

		setActiveStep(3);
		setTypedText("");
	}

	function handleTeamKeyDown(event: ReactKeyboardEvent<HTMLInputElement>) {
		if (event.key !== "Enter") {
			return;
		}

		event.preventDefault();
		advanceToFriends();
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

		if (!friendsPromptReady) {
			return;
		}

		const normalizedTeamName = teamName.trim() || "Yoru Labs";
		await createTeam({
			name: normalizedTeamName,
			slug:
				normalizedTeamName
					.toLowerCase()
					.trim()
					.replace(/[^a-z0-9]+/g, "-")
					.replace(/^-|-$/g, "") || "yoru-team"
		});
		setCreatedTeam(normalizedTeamName);
	}

	return (
		<PageShell variant="constellation" center className={styles.shell}>
			<Helmet>
				<title>Create Team - Netstamp</title>
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

					{createdTeam ? (
						<div className={styles.successView} aria-live="polite">
							<ScriptLine prompt="success" text={`Team ${createdTeam} created.`} />
							<p>Nice, let's bring {createdTeam} online. Next we will open the probe fleet and start the new probe wizard.</p>
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
								{teamPromptReady ? (
									<label className={styles.answerRow}>
										<span className={styles.answerPrompt}>answer</span>
										<Input
											variant="bare"
											ref={teamInputRef}
											name="team"
											value={teamName}
											placeholder="Yoru Labs"
											onChange={event => setTeamName(event.currentTarget.value)}
											onKeyDown={handleTeamKeyDown}
											autoComplete="organization"
										/>
										{activeStep === 2 ? <small>Press Enter to continue.</small> : null}
									</label>
								) : null}

								{activeStep >= 3 ? (
									<ScriptLine prompt="friends" text={activeStep === 3 ? typedText : scriptSteps[3].text} cursor={activeStep === 3 && typedText.length < scriptSteps[3].text.length} />
								) : null}

								{friendsPromptReady ? (
									<div className={styles.inviteSection}>
										<div className={styles.inviteHeader}>
											<p>Press Enter for next friend email. Backspace on an empty row deletes it.</p>
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
															placeholder="friend@example.com"
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

										<Button variant="plain" className={styles.tuiButton} type="submit" disabled={submitting}>
											{submitting ? "[ creating team... ]" : "[ create team ]"}
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
