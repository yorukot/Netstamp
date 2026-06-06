import { useSession } from "@/features/auth/session/SessionContext";
import { pathForRoute } from "@/routes/routePaths";
import { useDeleteProjectMutation, useRemoveProjectMemberMutation, useUpdateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, Panel, Surface, TextField } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./ProjectPage.module.css";

interface ProjectDraft {
	projectId: string;
	name: string | null;
	slug: string | null;
}

export function ProjectPage() {
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

	async function deleteCurrentProject() {
		if (!project) {
			return;
		}

		const confirmed = await confirm({
			title: `Delete ${project.name}?`,
			message: "This deletes the project, disables future assignments, and revokes all probe registration tokens.",
			confirmLabel: "Delete project",
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
			title: `Leave ${project?.name ?? "project"}?`,
			message: "This removes your access to the project's probes, checks, alerts, and measurements.",
			confirmLabel: "Leave project",
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
			<ScreenHeader title="Project" />

			<Panel tone="glass" title="Project info">
				<div className={styles.projectInfoGrid}>
					<TextField label="Project name" value={activeProjectName} disabled={!projectRef} onChange={event => updateProjectNameDraft(event.currentTarget.value)} />
					<TextField label="Slug" value={activeProjectSlug} disabled={!projectRef} onChange={event => updateProjectSlugDraft(event.currentTarget.value)} />
				</div>
				<Button
					disabled={!projectRef || updateProjectMutation.isPending}
					onClick={() =>
						updateProjectMutation.mutate(
							{ name: activeProjectName, slug: activeProjectSlug },
							{
								onSuccess: data => {
									const nextProjectRef = data.project.slug || data.project.id;
									setProjectDraft({ projectId: data.project.id, name: null, slug: null });
									setSelectedProjectRef(nextProjectRef);
									navigate(pathForRoute("project", { projectRef: nextProjectRef }), { replace: true });
								}
							}
						)
					}
				>
					{updateProjectMutation.isPending ? "Saving" : "Save changes"}
				</Button>
			</Panel>

			<Panel tone="deep" title="Dangerous project actions">
				<div className={styles.dangerZoneGrid}>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Delete project</h3>
						<p className={styles.warningCopy}>Delete this project, disable future assignments, and revoke all probe registration tokens.</p>
						<Button variant="danger" disabled={!projectRef || deleteProjectMutation.isPending} onClick={() => void deleteCurrentProject()}>
							{deleteProjectMutation.isPending ? "Deleting" : "Delete project"}
						</Button>
					</Surface>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Leave project</h3>
						<p className={styles.warningCopy}>
							{isCurrentOwner ? "Owners cannot leave a project while they hold the owner role." : "Leave this project and remove your access to its probes, checks, alerts, and measurements."}
						</p>
						<Button variant="outline" disabled={!canLeaveProject} onClick={() => void leaveCurrentProject()}>
							{removeMemberMutation.isPending ? "Leaving" : "Leave project"}
						</Button>
					</Surface>
				</div>
			</Panel>
		</PageStack>
	);
}
