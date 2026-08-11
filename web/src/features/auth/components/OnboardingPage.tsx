import { useAuth } from "@/features/auth/hooks/useAuth";
import { type Navigate } from "@/routes/routeTypes";
import { apiProblemCode } from "@/shared/api/client";
import { useAcceptProjectInviteMutation, useCreateProjectInviteForRefMutation, useCreateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite, ProjectMemberRole } from "@/shared/api/types";
import { useProjectSelection } from "@/shared/api/useCurrentProject";
import { useRuntimeFeatures } from "@/shared/config/features";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, DisclosureToggle, IconButton, Spinner, TextField } from "@netstamp/ui";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { type FormEvent, useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { AuthLayout } from "./AuthLayout";
import styles from "./OnboardingPage.module.css";

interface OnboardingPageProps {
	navigate: Navigate;
}

const maxProjectSlugLength = 64;
const maxSlugAttempts = 20;
const randomSlugTokenLength = 6;
const randomSlugTokenAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789";

const randomSlugToken = (length: number) => {
	const values = new Uint8Array(length);
	globalThis.crypto.getRandomValues(values);

	return Array.from(values, value => randomSlugTokenAlphabet[value % randomSlugTokenAlphabet.length]).join("");
};

const slugifyProjectName = (name: string) =>
	name
		.toLowerCase()
		.trim()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-|-$/g, "")
		.slice(0, maxProjectSlugLength);

const projectSlugCandidate = (baseSlug: string, attempt: number) => {
	if (attempt === 0) {
		return baseSlug.slice(0, maxProjectSlugLength);
	}

	const suffix = `-${randomSlugToken(randomSlugTokenLength)}`;
	const baseLength = maxProjectSlugLength - suffix.length;
	return `${baseSlug.slice(0, Math.max(1, baseLength)).replace(/-$/g, "")}${suffix}`;
};

const isProjectSlugConflict = (error: unknown) => apiProblemCode(error) === "PROJECT_SLUG_ALREADY_EXISTS";

const projectRefFromInvite = (invite: ApiProjectInvite) => invite.project.slug || invite.project.id;

export const OnboardingPage = ({ navigate }: OnboardingPageProps) => {
	const { t } = useTranslation(["auth", "project"]);
	const { loading, submitting, logout } = useAuth();
	const { appFeatures, readOnlyMode } = useRuntimeFeatures();
	const createProjectMutation = useCreateProjectMutation({ suppressGlobalErrorToast: true });
	const createInviteMutation = useCreateProjectInviteForRefMutation({ suppressGlobalErrorToast: true });
	const acceptInviteMutation = useAcceptProjectInviteMutation();
	const pendingInvitesQuery = useQuery(projectQueries.currentUserInvites());
	const { setSelectedProjectRef } = useProjectSelection();
	const inviteSectionId = useId();
	const [projectName, setProjectName] = useState("");
	const [invites, setInvites] = useState([""]);
	const [inviteSectionOpen, setInviteSectionOpen] = useState(false);
	const [inviteChoiceResolved, setInviteChoiceResolved] = useState(false);
	const [creatingProject, setCreatingProject] = useState(false);
	const [resolvingInvites, setResolvingInvites] = useState(false);
	const pendingInvites = pendingInvitesQuery.data?.invites ?? [];
	const showInviteChoice = pendingInvites.length > 0 && !inviteChoiceResolved;
	const noProjectAccess = !appFeatures.projectCreation && !showInviteChoice;
	const acceptingInvites = resolvingInvites || acceptInviteMutation.isPending;
	const busy = submitting || creatingProject || createProjectMutation.isPending || acceptingInvites;
	const projectNameTrimmed = projectName.trim();
	const canCreate = Boolean(projectNameTrimmed && !readOnlyMode && !busy);
	const pendingInviteSummary =
		pendingInvites.length === 1 && pendingInvites[0]
			? t("onboarding.singleInvite", { role: t(`project:roles.${pendingInvites[0].role as ProjectMemberRole}`), project: pendingInvites[0].project.name })
			: t("onboarding.inviteCount", { count: pendingInvites.length });
	const title = showInviteChoice ? t("onboarding.inviteTitle") : noProjectAccess ? t("onboarding.noAccessTitle") : t("onboarding.createTitle");
	const description = showInviteChoice ? t("onboarding.inviteDescription") : noProjectAccess ? t("onboarding.noProjectsHelp") : t("onboarding.createDescription");

	const normalizedInviteEmails = () => Array.from(new Set(invites.map(invite => invite.trim()).filter(Boolean)));

	const acceptPendingInvites = async () => {
		setResolvingInvites(true);

		try {
			const inviteResults = await Promise.allSettled(pendingInvites.map(invite => acceptInviteMutation.mutateAsync(invite.id)));
			const acceptedInvites = inviteResults.flatMap(result => (result.status === "fulfilled" ? [result.value.invite] : []));
			const failedInviteCount = inviteResults.length - acceptedInvites.length;

			if (failedInviteCount && acceptedInvites.length) {
				pushErrorToast(t("onboarding.inviteAcceptFailed", { count: failedInviteCount }));
			}

			const acceptedInvite = acceptedInvites[0];

			if (!acceptedInvite) {
				throw new Error(t("onboarding.inviteAcceptError"));
			}

			return {
				name: acceptedInvite.project.name,
				ref: projectRefFromInvite(acceptedInvite)
			};
		} finally {
			setResolvingInvites(false);
		}
	};

	const openInvitedProject = async () => {
		if (readOnlyMode || acceptingInvites) {
			return;
		}

		try {
			const project = await acceptPendingInvites();
			setSelectedProjectRef(project.ref);
			navigate("dashboard", { projectRef: project.ref });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, t("onboarding.inviteAcceptError")));
		}
	};

	const continueToProjectCreation = async () => {
		if (readOnlyMode || acceptingInvites || !appFeatures.projectCreation) {
			return;
		}

		try {
			await acceptPendingInvites();
			setInviteChoiceResolved(true);
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, t("onboarding.inviteAcceptError")));
		}
	};

	const updateInvite = (index: number, value: string) => {
		setInvites(current => current.map((invite, currentIndex) => (currentIndex === index ? value : invite)));
	};

	const addInvite = () => {
		setInvites(current => [...current, ""]);
	};

	const removeInvite = (index: number) => {
		setInvites(current => (current.length === 1 ? [""] : current.filter((_, currentIndex) => currentIndex !== index)));
	};

	const submitProject = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();

		if (!canCreate) {
			return;
		}

		const normalizedSlug = slugifyProjectName(projectNameTrimmed);
		const baseSlug = normalizedSlug || `project-${randomSlugToken(randomSlugTokenLength)}`;

		setCreatingProject(true);

		try {
			for (let attempt = 0; attempt < maxSlugAttempts; attempt += 1) {
				try {
					const data = await createProjectMutation.mutateAsync({
						name: projectNameTrimmed,
						slug: projectSlugCandidate(baseSlug, attempt)
					});
					const projectRef = data.project.slug || data.project.id;
					const inviteEmails = normalizedInviteEmails();

					if (inviteEmails.length) {
						const inviteResults = await Promise.allSettled(inviteEmails.map(email => createInviteMutation.mutateAsync({ projectRef, body: { email, role: "viewer" } })));
						const failedInviteCount = inviteResults.filter(result => result.status === "rejected").length;

						if (failedInviteCount) {
							pushErrorToast(t("onboarding.inviteSendFailed", { count: failedInviteCount }));
						}
					}

					setSelectedProjectRef(projectRef);
					pushToast({
						title: t("project:create.successTitle"),
						message: t("project:create.successMessage", { name: data.project.name }),
						tone: "success"
					});
					navigate("dashboard", { projectRef });
					return;
				} catch (error) {
					if (!isProjectSlugConflict(error)) {
						throw error;
					}
				}
			}

			pushErrorToast(t("onboarding.slugConflict"));
		} catch (error) {
			pushErrorToast(requestErrorMessage(error, t("onboarding.createError")));
		} finally {
			setCreatingProject(false);
		}
	};

	if (loading || pendingInvitesQuery.isPending) {
		return (
			<AuthLayout title={t("onboarding.welcomeTitle")} description={t("onboarding.loadingDescription")}>
				<Spinner label={t("onboarding.loading")} layout="panel" size="lg" />
			</AuthLayout>
		);
	}

	return (
		<AuthLayout title={title} description={description}>
			{showInviteChoice ? (
				<div className={styles.inviteChoice}>
					<div className={styles.inviteSummary}>
						<strong>{pendingInviteSummary}</strong>
						<span>{t("onboarding.inviteChoiceHelp")}</span>
					</div>

					<div className={styles.choiceActions}>
						<Button type="button" size="lg" disabled={readOnlyMode || acceptingInvites} onClick={() => void openInvitedProject()}>
							{acceptingInvites ? t("onboarding.accepting") : t("onboarding.openInvited")}
						</Button>
						{appFeatures.projectCreation ? (
							<Button type="button" variant="outline" size="lg" disabled={readOnlyMode || acceptingInvites} onClick={() => void continueToProjectCreation()}>
								{t("onboarding.createAnotherAction")}
							</Button>
						) : null}
					</div>
				</div>
			) : noProjectAccess ? (
				<div className={styles.noAccess}>
					<Button type="button" variant="outline" disabled={submitting} onClick={logout}>
						{t("onboarding.logOut")}
					</Button>
				</div>
			) : (
				<form className={styles.form} onSubmit={submitProject}>
					<TextField
						label={t("onboarding.projectName")}
						helper={t("onboarding.projectNameHelper")}
						value={projectName}
						onChange={event => setProjectName(event.currentTarget.value)}
						autoComplete="organization"
						autoFocus
						required
						disabled={busy || readOnlyMode}
					/>

					<section className={styles.inviteSection} aria-labelledby={`${inviteSectionId}-title`}>
						<div className={styles.inviteHeader}>
							<div>
								<strong id={`${inviteSectionId}-title`}>{t("onboarding.inviteMembers")}</strong>
								<span>{t("onboarding.inviteOptional")}</span>
							</div>
							<DisclosureToggle
								open={inviteSectionOpen}
								label={t(inviteSectionOpen ? "onboarding.hideInvites" : "onboarding.showInvites")}
								aria-controls={inviteSectionId}
								disabled={busy || readOnlyMode}
								onClick={() => setInviteSectionOpen(current => !current)}
							/>
						</div>

						{inviteSectionOpen ? (
							<div className={styles.inviteBody} id={inviteSectionId}>
								<div className={styles.inviteList}>
									{invites.map((invite, index) => (
										<div className={styles.inviteRow} key={index}>
											<TextField
												label={t("onboarding.memberEmail", { number: index + 1 })}
												type="email"
												value={invite}
												placeholder="member@example.com"
												onChange={event => updateInvite(index, event.currentTarget.value)}
												disabled={busy || readOnlyMode}
											/>
											{invites.length > 1 ? (
												<IconButton
													className={styles.inviteRemoveButton}
													aria-label={t("onboarding.removeInvite", { number: index + 1 })}
													danger
													size="md"
													disabled={busy || readOnlyMode}
													onClick={() => removeInvite(index)}
												>
													<TrashIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
												</IconButton>
											) : null}
										</div>
									))}
								</div>

								<Button type="button" variant="outline" size="sm" disabled={busy || readOnlyMode} onClick={addInvite}>
									{t("onboarding.addInvite")}
								</Button>
							</div>
						) : null}
					</section>

					<Button className={styles.createButton} type="submit" size="lg" disabled={!canCreate}>
						{creatingProject ? t("onboarding.creating") : t("onboarding.create")}
					</Button>
				</form>
			)}
		</AuthLayout>
	);
};
