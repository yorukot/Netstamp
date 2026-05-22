import { useAddProjectMemberMutation, useDeleteProjectMutation, useRemoveProjectMemberMutation, useUpdateProjectMemberRoleMutation, useUpdateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ProjectMemberRole } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, DataTable, Panel, SelectField, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import styles from "./ProjectPage.module.css";
import { RoleSelect } from "./RoleSelect";

interface MemberRow {
	id: string;
	userId: string;
	name: string;
	email: string;
	role: string;
	lastActive: string;
}

export function ProjectPage() {
	const { project, projectRef, setSelectedProjectRef } = useCurrentProject();
	const confirm = useConfirm();
	const updateProjectMutation = useUpdateProjectMutation(projectRef);
	const addMemberMutation = useAddProjectMemberMutation(projectRef);
	const removeMemberMutation = useRemoveProjectMemberMutation(projectRef);
	const updateMemberRoleMutation = useUpdateProjectMemberRoleMutation(projectRef);
	const deleteProjectMutation = useDeleteProjectMutation(projectRef);
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const [projectName, setProjectName] = useState("");
	const [projectSlug, setProjectSlug] = useState("");
	const [memberEmail, setMemberEmail] = useState("");
	const [memberRole, setMemberRole] = useState<ProjectMemberRole>("viewer");
	const activeProjectName = projectName || project?.name || "";
	const activeProjectSlug = projectSlug || project?.slug || "";
	function addCurrentMember() {
		addMemberMutation.mutate(
			{ email: memberEmail, role: memberRole },
			{
				onSuccess: () => {
					setMemberEmail("");
				}
			}
		);
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
			}
		});
	}
	const memberRows: MemberRow[] = (membersQuery.data?.members ?? []).map(member => ({
		id: member.id,
		userId: member.userId,
		name: member.user.displayName,
		email: member.user.email,
		role: member.role,
		lastActive: new Date(member.updatedAt).toLocaleString()
	}));
	const memberColumns: DataColumn<MemberRow>[] = [
		{ key: "name", label: "User ID" },
		{ key: "email", label: "Account" },
		{ key: "role", label: "Role", render: row => <RoleSelect role={row.role} name={row.name} onRoleChange={role => updateMemberRoleMutation.mutate({ userId: row.userId, role })} /> },
		{ key: "lastActive", label: "Last active" },
		{
			key: "delete",
			label: "Delete",
			render: row => (
				<Button variant="danger" size="sm" onClick={() => removeMemberMutation.mutate(row.userId)}>
					Delete
				</Button>
			)
		}
	];

	return (
		<PageStack>
			<ScreenHeader eyebrow="Project settings" title="Project" copy="Project profile, member management, and destructive project actions." />

			<Panel tone="glass" eyebrow="Project" title="Project info">
				<div className={styles.projectInfoGrid}>
					<TextField label="Project name" value={activeProjectName} disabled={!projectRef} onChange={event => setProjectName(event.currentTarget.value)} />
					<TextField label="Slug" value={activeProjectSlug} disabled={!projectRef} onChange={event => setProjectSlug(event.currentTarget.value)} />
				</div>
				<Button disabled={!projectRef || updateProjectMutation.isPending} onClick={() => updateProjectMutation.mutate({ name: activeProjectName, slug: activeProjectSlug })}>
					{updateProjectMutation.isPending ? "Saving" : "Save changes"}
				</Button>
			</Panel>

			<Panel tone="glass" eyebrow="Members" title="Member management">
				<div className={styles.formGridThree}>
					<TextField label="Email" value={memberEmail} onChange={event => setMemberEmail(event.currentTarget.value)} />
					<SelectField
						label="Role"
						value={memberRole}
						onChange={event => setMemberRole(event.currentTarget.value as ProjectMemberRole)}
						options={[
							{ value: "admin", label: "Admin" },
							{ value: "editor", label: "Editor" },
							{ value: "viewer", label: "Viewer" }
						]}
					/>
					<Button disabled={!projectRef || !memberEmail || addMemberMutation.isPending} onClick={addCurrentMember}>
						{addMemberMutation.isPending ? "Adding" : "Add member"}
					</Button>
				</div>
				<DataTable columns={memberColumns} rows={memberRows} getRowKey={row => row.id} />
			</Panel>

			<Panel tone="deep" eyebrow="Danger zone" title="Dangerous project actions">
				<div className={styles.dangerZoneGrid}>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Delete project</h3>
						<p className={styles.warningCopy}>Delete this project, disable future assignments, and revoke all probe registration tokens.</p>
						<Button variant="danger" disabled={!projectRef || deleteProjectMutation.isPending} onClick={() => void deleteCurrentProject()}>
							{deleteProjectMutation.isPending ? "Deleting" : "Delete project"}
						</Button>
					</Surface>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Exit project</h3>
						<p className={styles.warningCopy}>Leave this project and remove your access to its probes, checks, alerts, and measurements.</p>
						<Button variant="outline">Exit project</Button>
					</Surface>
				</div>
			</Panel>
		</PageStack>
	);
}
