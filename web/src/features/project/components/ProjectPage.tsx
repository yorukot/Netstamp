import { useSession } from "@/features/auth/session/SessionContext";
import { pathForRoute } from "@/routes/routePaths";
import { useDeleteProjectMutation, useRemoveProjectMemberMutation, useUpdateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { Button, DangerAction, Panel, TextField } from "@netstamp/ui";
import { SignOutIcon } from "@phosphor-icons/react/dist/csr/SignOut";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import styles from "./ProjectPage.module.css";

interface ProjectDraft {
	projectId: string;
	name: string | null;
	slug: string | null;
}

export function ProjectPage() {
	const { t } = useTranslation("project");
	const { project, projectRef, setSelectedProjectRef } = useCurrentProject();
	const { session } = useSession();
	const confirm = useConfirm();
	const navigate = useNavigate();
	const updateProjectMutation = useUpdateProjectMutation(projectRef);
	const removeMemberMutation = useRemoveProjectMemberMutation(projectRef);
	const deleteProjectMutation = useDeleteProjectMutation(projectRef);
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const [projectDraft, setProjectDraft] = useState<ProjectDraft>({ projectId: "", name: null, slug: null });
	const activeProjectId = project?.id ?? "";
	const activeProjectDraft = projectDraft.projectId === activeProjectId ? projectDraft : null;
	const activeProjectName = activeProjectDraft?.name ?? project?.name ?? "";
	const activeProjectSlug = activeProjectDraft?.slug ?? project?.slug ?? "";
	const hasProjectChanges = Boolean(project && (activeProjectName !== project.name || activeProjectSlug !== project.slug));
	const currentUserId = session?.user.id ?? "";
	const members = membersQuery.data?.members ?? [];
	const currentMember = members.find(member => member.userId === currentUserId);
	const isCurrentOwner = currentMember?.role === "owner";
	const canLeaveProject = Boolean(projectRef && currentUserId && currentMember && !isCurrentOwner && !removeMemberMutation.isPending);

	function updateProjectNameDraft(name: string) {
		setProjectDraft(current => ({
			projectId: activeProjectId,
			name,
			slug: current.projectId === activeProjectId ? current.slug : null
		}));
	}

	function updateProjectSlugDraft(slug: string) {
		setProjectDraft(current => ({
			projectId: activeProjectId,
			name: current.projectId === activeProjectId ? current.name : null,
			slug
		}));
	}

	function resetProjectDraft() {
		setProjectDraft({ projectId: activeProjectId, name: null, slug: null });
	}

	function saveProjectSettings() {
		if (!projectRef) {
			return;
		}

		updateProjectMutation.mutate(
			{ name: activeProjectName, slug: activeProjectSlug },
			{
				onSuccess: data => {
					const nextProjectRef = data.project.slug || data.project.id;
					setProjectDraft({ projectId: data.project.id, name: null, slug: null });
					setSelectedProjectRef(nextProjectRef);
					navigate(pathForRoute("projectSettings", { projectRef: nextProjectRef }), { replace: true });
				}
			}
		);
	}

	async function deleteCurrentProject() {
		if (!project) {
			return;
		}

		const confirmed = await confirm({
			title: t("settings.deleteQuestion"),
			message: t("settings.deleteMessage"),
			confirmLabel: t("settings.delete"),
			confirmationText: project.name,
			confirmationLabel: t("settings.name"),
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteProjectMutation.mutate(undefined, {
			onSuccess: () => {
				setSelectedProjectRef("");
				navigate(pathForRoute("dashboard"));
			}
		});
	}

	async function leaveCurrentProject() {
		if (!canLeaveProject) {
			return;
		}

		const confirmed = await confirm({
			title: t("settings.leaveQuestion", { name: project?.name ?? t("settings.projectFallback") }),
			message: t("settings.leaveMessage"),
			confirmLabel: t("settings.leave"),
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		removeMemberMutation.mutate(currentUserId, {
			onSuccess: () => {
				setSelectedProjectRef("");
				navigate(pathForRoute("dashboard"));
			}
		});
	}

	return (
		<PageStack>
			<ScreenHeader title={t("settings.title")} />

			<Panel tone="glass" title={t("settings.info")}>
				<div className={styles.projectInfoGrid}>
					<TextField label={t("settings.name")} value={activeProjectName} disabled={!projectRef} onChange={event => updateProjectNameDraft(event.currentTarget.value)} />
					<TextField label={t("settings.slug")} value={activeProjectSlug} disabled={!projectRef} onChange={event => updateProjectSlugDraft(event.currentTarget.value)} />
				</div>
				<UnsavedChangesBar show={hasProjectChanges} saving={updateProjectMutation.isPending} disabled={!projectRef} onReset={resetProjectDraft} onSave={saveProjectSettings} />
			</Panel>

			<Panel tone="deep" title={t("settings.dangerous")} padded={false} bodySurface="transparent">
				<div className={styles.dangerActionList}>
					<DangerAction
						title={t("settings.delete")}
						description={t("settings.deleteDescription")}
						descriptionId="delete-project-description"
						action={
							<Button variant="danger" disabled={!projectRef || deleteProjectMutation.isPending} aria-describedby="delete-project-description" onClick={() => void deleteCurrentProject()}>
								<TrashIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
								{deleteProjectMutation.isPending ? t("settings.deleting") : t("settings.delete")}
							</Button>
						}
					/>
					<DangerAction
						title={t("settings.leave")}
						description={isCurrentOwner ? t("settings.ownerCannotLeave") : t("settings.leaveDescription")}
						descriptionId="leave-project-description"
						action={
							<Button variant="danger" disabled={!canLeaveProject} aria-describedby="leave-project-description" onClick={() => void leaveCurrentProject()}>
								<SignOutIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
								{removeMemberMutation.isPending ? t("settings.leaving") : t("settings.leave")}
							</Button>
						}
					/>
				</div>
			</Panel>
		</PageStack>
	);
}
