import { useSession } from "@/features/auth/session/SessionContext";
import { useCreateProjectInviteMutation, useDeleteProjectMutation, useRemoveProjectMemberMutation, useUpdateProjectMemberRoleMutation, useUpdateProjectMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite, ProjectMemberRole } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { Badge, Button, DataTable, Panel, SelectField, Surface, TextField, type DataColumn } from "@netstamp/ui";
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
	isCurrentUser: boolean;
}

interface InviteRow {
	id: string;
	name: string;
	email: string;
	role: ProjectMemberRole;
	invitedBy: string;
	createdAt: string;
	status: string;
}

interface ProjectDraft {
	projectId: string;
	name: string | null;
	slug: string | null;
}

function formatDateTime(value: string) {
	return new Date(value).toLocaleString();
}

function roleLabel(role: string) {
	return role.charAt(0).toUpperCase() + role.slice(1);
}

function canRemoveMember(actorRole: string | undefined, memberRole: string, isCurrentUser: boolean) {
	if (isCurrentUser) {
		return false;
	}

	if (actorRole === "owner") {
		return true;
	}

	if (actorRole === "admin") {
		return memberRole === "editor" || memberRole === "viewer";
	}

	return false;
}

function blockedRemoveLabel(actorRole: string | undefined, memberRole: string) {
	if (memberRole === "owner") {
		return "Owner protected";
	}

	if (actorRole === "admin" && memberRole === "admin") {
		return "Admin protected";
	}

	return "No access";
}

function canUpdateMemberRole(actorRole: string | undefined, memberRole: string) {
	if (actorRole === "owner") {
		return true;
	}

	if (actorRole === "admin") {
		return memberRole !== "owner";
	}

	return false;
}

export function ProjectPage() {
	const { project, projectRef, setSelectedProjectRef } = useCurrentProject();
	const { session } = useSession();
	const confirm = useConfirm();
	const updateProjectMutation = useUpdateProjectMutation(projectRef);
	const createInviteMutation = useCreateProjectInviteMutation(projectRef);
	const removeMemberMutation = useRemoveProjectMemberMutation(projectRef);
	const updateMemberRoleMutation = useUpdateProjectMemberRoleMutation(projectRef);
	const deleteProjectMutation = useDeleteProjectMutation(projectRef);
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const [projectDraft, setProjectDraft] = useState<ProjectDraft>({ projectId: "", name: null, slug: null });
	const [memberEmail, setMemberEmail] = useState("");
	const [memberRole, setMemberRole] = useState<ProjectMemberRole>("viewer");
	const activeProjectId = project?.id ?? "";
	const activeProjectDraft = projectDraft.projectId === activeProjectId ? projectDraft : null;
	const activeProjectName = activeProjectDraft?.name ?? project?.name ?? "";
	const activeProjectSlug = activeProjectDraft?.slug ?? project?.slug ?? "";
	const currentUserId = session?.user.id ?? "";
	const members = membersQuery.data?.members ?? [];
	const currentMember = members.find(member => member.userId === currentUserId);
	const currentMemberRole = currentMember?.role;
	const isCurrentOwner = currentMember?.role === "owner";
	const canManageMembers = currentMember?.role === "owner" || currentMember?.role === "admin";
	const canLeaveProject = Boolean(projectRef && currentUserId && currentMember && !isCurrentOwner && !removeMemberMutation.isPending);
	const invitesQuery = useQuery({
		...projectQueries.invites(projectRef || ""),
		enabled: Boolean(projectRef && canManageMembers)
	});

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

	function createCurrentInvite() {
		createInviteMutation.mutate(
			{ email: memberEmail, role: memberRole },
			{
				onSuccess: data => {
					setMemberEmail("");
					pushToast({
						title: "Invite sent",
						message: `${data.invite.invitedUser.email} can now accept access to ${data.invite.project.name}.`,
						tone: "success"
					});
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
			}
		});
	}

	const memberRows: MemberRow[] = (membersQuery.data?.members ?? []).map(member => ({
		id: member.id,
		userId: member.userId,
		name: member.user.displayName,
		email: member.user.email,
		role: member.role,
		lastActive: new Date(member.updatedAt).toLocaleString(),
		isCurrentUser: member.userId === currentUserId
	}));
	const inviteRows: InviteRow[] = (invitesQuery.data?.invites ?? []).map((invite: ApiProjectInvite) => ({
		id: invite.id,
		name: invite.invitedUser.displayName,
		email: invite.invitedUser.email,
		role: invite.role,
		invitedBy: invite.invitedByUser.displayName,
		createdAt: formatDateTime(invite.createdAt),
		status: invite.status
	}));
	const memberColumns: DataColumn<MemberRow>[] = [
		{ key: "name", label: "User ID" },
		{ key: "email", label: "Account" },
		{
			key: "role",
			label: "Role",
			render: row => (
				<RoleSelect
					role={row.role}
					name={row.name}
					disabled={updateMemberRoleMutation.isPending || !canUpdateMemberRole(currentMemberRole, row.role)}
					onRoleChange={role => updateMemberRoleMutation.mutate({ userId: row.userId, role })}
				/>
			)
		},
		{ key: "lastActive", label: "Last active" },
		{
			key: "delete",
			label: "Delete",
			render: row => {
				const canDeleteMember = Boolean(projectRef) && canRemoveMember(currentMemberRole, row.role, row.isCurrentUser);

				if (row.isCurrentUser) {
					return (
						<Button variant="outline" size="sm" disabled title="Use Leave project when your role allows it.">
							Current user
						</Button>
					);
				}

				if (!canDeleteMember) {
					return (
						<Button variant="outline" size="sm" disabled>
							{blockedRemoveLabel(currentMemberRole, row.role)}
						</Button>
					);
				}

				return (
					<Button variant="danger" size="sm" disabled={removeMemberMutation.isPending} onClick={() => removeMemberMutation.mutate(row.userId)}>
						{removeMemberMutation.isPending ? "Deleting" : "Delete"}
					</Button>
				);
			}
		}
	];
	const inviteColumns: DataColumn<InviteRow>[] = [
		{
			key: "name",
			label: "Invited user",
			render: row => (
				<span className={styles.stackedCell}>
					<strong>{row.name}</strong>
					<span>{row.email}</span>
				</span>
			)
		},
		{ key: "role", label: "Role", render: row => <Badge tone="accent">{roleLabel(row.role)}</Badge> },
		{ key: "invitedBy", label: "Invited by" },
		{ key: "createdAt", label: "Sent" },
		{ key: "status", label: "Status", render: row => <Badge tone="warning">{roleLabel(row.status)}</Badge> }
	];

	return (
		<PageStack>
			<ScreenHeader title="Project" />

			<Panel tone="glass" eyebrow="Project" title="Project info">
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
								onSuccess: data => setProjectDraft({ projectId: data.project.id, name: null, slug: null })
							}
						)
					}
				>
					{updateProjectMutation.isPending ? "Saving" : "Save changes"}
				</Button>
			</Panel>

			{canManageMembers ? (
				<Panel tone="glass" eyebrow="Invites" title="Invite member">
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
						<Button disabled={!projectRef || !memberEmail || createInviteMutation.isPending} onClick={createCurrentInvite}>
							{createInviteMutation.isPending ? "Sending" : "Send invite"}
						</Button>
					</div>
				</Panel>
			) : null}

			{canManageMembers ? (
				<Panel tone="glass" eyebrow="Pending access" title={`${inviteRows.length} pending invites`}>
					<DataTable
						columns={inviteColumns}
						rows={inviteRows}
						density="compact"
						minWidth="46rem"
						emptyLabel={invitesQuery.isLoading ? "Loading pending invites" : "No pending invites"}
						getRowKey={row => row.id}
					/>
				</Panel>
			) : null}

			<Panel tone="glass" eyebrow="Members" title="Member access">
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
