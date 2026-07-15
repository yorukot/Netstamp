import { useSession } from "@/features/auth/session/SessionContext";
import { useCancelProjectInviteMutation, useCreateProjectInviteMutation, useRemoveProjectMemberMutation, useUpdateProjectMemberRoleMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite, ProjectMemberRole } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { createGravatarUrl } from "@/shared/utils/gravatar";
import { Badge, Button, DataTable, Panel, SelectField, SignalAvatar, Spinner, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import styles from "./MembersPage.module.css";
import { RoleSelect } from "./RoleSelect";

interface MemberRow {
	id: string;
	userId: string;
	name: string;
	email: string;
	avatarUrl: string | undefined;
	initials: string;
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

function formatDateTime(value: string) {
	return new Date(value).toLocaleString();
}

function roleLabel(role: string) {
	return role.charAt(0).toUpperCase() + role.slice(1);
}

function roleRank(role: string) {
	switch (role) {
		case "owner":
			return 0;
		case "admin":
			return 1;
		case "editor":
			return 2;
		case "viewer":
			return 3;
		default:
			return 4;
	}
}

function memberInitials(name: string, email: string) {
	const nameParts = name.trim().split(/\s+/).filter(Boolean);
	const sourceParts = nameParts.length > 0 ? nameParts : [email.trim()];
	const first = sourceParts[0]?.[0] ?? "?";
	const last = sourceParts.length > 1 ? (sourceParts.at(-1)?.[0] ?? "") : (sourceParts[0]?.[1] ?? "");

	return `${first}${last}`.toUpperCase();
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

export function MembersPage() {
	const { projectRef } = useCurrentProject();
	const { session } = useSession();
	const createInviteMutation = useCreateProjectInviteMutation(projectRef);
	const cancelInviteMutation = useCancelProjectInviteMutation(projectRef);
	const removeMemberMutation = useRemoveProjectMemberMutation(projectRef);
	const updateMemberRoleMutation = useUpdateProjectMemberRoleMutation(projectRef);
	const [memberEmail, setMemberEmail] = useState("");
	const [memberRole, setMemberRole] = useState<ProjectMemberRole>("viewer");
	const [memberAvatarUrls, setMemberAvatarUrls] = useState<Record<string, string>>({});
	const currentUserId = session?.user.id ?? "";
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const rawMembers = membersQuery.data?.members;
	const members = rawMembers ?? [];
	const currentMember = members.find(member => member.userId === currentUserId);
	const currentMemberRole = currentMember?.role;
	const canManageMembers = currentMember?.role === "owner" || currentMember?.role === "admin";
	const invitesQuery = useQuery({
		...projectQueries.invites(projectRef || ""),
		enabled: Boolean(projectRef && canManageMembers)
	});

	useEffect(() => {
		let cancelled = false;
		const emails = Array.from(new Set((rawMembers ?? []).map(member => member.user.email).filter(Boolean)));

		if (emails.length === 0) {
			return () => {
				cancelled = true;
			};
		}

		void Promise.all(
			emails.map(async email => {
				try {
					return [email, await createGravatarUrl(email, 80)] as const;
				} catch {
					return [email, ""] as const;
				}
			})
		).then(entries => {
			if (cancelled) {
				return;
			}

			setMemberAvatarUrls(Object.fromEntries(entries.filter(([, avatarUrl]) => avatarUrl)));
		});

		return () => {
			cancelled = true;
		};
	}, [rawMembers]);

	function createCurrentInvite() {
		createInviteMutation.mutate(
			{ email: memberEmail, role: memberRole },
			{
				onSuccess: data => {
					setMemberEmail("");
					pushToast({
						title: "Invite sent",
						message: `${data.invite.invitedEmail} can now accept access to ${data.invite.project.name}.`,
						tone: "success"
					});
				}
			}
		);
	}

	function cancelInvite(row: InviteRow) {
		cancelInviteMutation.mutate(row.id, {
			onSuccess: () => {
				pushToast({
					title: "Invite canceled",
					message: `${row.email} no longer has a pending invite.`,
					tone: "success"
				});
			}
		});
	}

	const memberRows: MemberRow[] = members
		.map(member => ({
			id: member.id,
			userId: member.userId,
			name: member.user.displayName,
			email: member.user.email,
			avatarUrl: memberAvatarUrls[member.user.email],
			initials: memberInitials(member.user.displayName, member.user.email),
			role: member.role,
			lastActive: formatDateTime(member.updatedAt),
			isCurrentUser: member.userId === currentUserId
		}))
		.sort((left, right) => {
			const rankDelta = roleRank(left.role) - roleRank(right.role);

			if (rankDelta !== 0) {
				return rankDelta;
			}

			return left.name.localeCompare(right.name) || left.email.localeCompare(right.email);
		});
	const inviteRows: InviteRow[] = (invitesQuery.data?.invites ?? []).map((invite: ApiProjectInvite) => ({
		id: invite.id,
		name: invite.invitedUser.displayName || invite.invitedEmail,
		email: invite.invitedEmail,
		role: invite.role,
		invitedBy: invite.invitedByUser.displayName,
		createdAt: formatDateTime(invite.createdAt),
		status: invite.status
	}));
	const memberColumns: DataColumn<MemberRow>[] = [
		{
			key: "name",
			label: "User ID",
			render: row => (
				<span className={styles.memberCell}>
					{row.avatarUrl ? (
						<SignalAvatar className={styles.memberAvatar} size="sm" src={row.avatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
					) : (
						<span className={styles.memberAvatarFallback} aria-hidden="true">
							{row.initials}
						</span>
					)}
					<span className={styles.memberName}>{row.name}</span>
				</span>
			)
		},
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
		{ key: "status", label: "Status", render: row => <Badge tone="warning">{roleLabel(row.status)}</Badge> },
		{
			key: "actions",
			label: "Actions",
			render: row => {
				const canceling = cancelInviteMutation.isPending && cancelInviteMutation.variables === row.id;

				return (
					<Button variant="danger" size="sm" disabled={cancelInviteMutation.isPending} onClick={() => cancelInvite(row)}>
						{canceling ? "Canceling" : "Cancel"}
					</Button>
				);
			}
		}
	];

	return (
		<PageStack>
			<ScreenHeader title="Members" />

			{canManageMembers ? (
				<Panel tone="glass" title="Invite Members" summary={`${inviteRows.length} pending invites`}>
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
					<DataTable
						columns={inviteColumns}
						rows={inviteRows}
						density="compact"
						minWidth="46rem"
						emptyLabel={invitesQuery.isLoading ? <Spinner label="Loading pending invites" layout="compact" size="lg" /> : "No pending invites"}
						getRowKey={row => row.id}
					/>
				</Panel>
			) : null}

			<Panel tone="glass" title="Member Access" padded={false} bodySurface="transparent">
				<DataTable columns={memberColumns} rows={memberRows} getRowKey={row => row.id} />
			</Panel>
		</PageStack>
	);
}
