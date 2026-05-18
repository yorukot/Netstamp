import {
	addProjectMember,
	createProjectLabel,
	deleteProject,
	deleteProjectLabel,
	projectQueries,
	removeProjectMember,
	updateProject,
	updateProjectLabel,
	updateProjectMemberRole
} from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import type { ApiLabel, ProjectMemberRole } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, DataTable, Panel, SelectField, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { RoleSelect } from "./RoleSelect";
import styles from "./TeamPage.module.css";

interface MemberRow {
	id: string;
	userId: string;
	name: string;
	email: string;
	role: string;
	lastActive: string;
}

interface LabelRow {
	id: string;
	key: string;
	value: string;
	updatedAt: string;
}

export function TeamPage() {
	const queryClient = useQueryClient();
	const { project, projectRef, setSelectedProjectRef } = useCurrentProject();
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const labelsQuery = useQuery({
		...projectQueries.labels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const [projectName, setProjectName] = useState("");
	const [projectSlug, setProjectSlug] = useState("");
	const [memberUserId, setMemberUserId] = useState("");
	const [memberRole, setMemberRole] = useState<ProjectMemberRole>("viewer");
	const [selectedLabelId, setSelectedLabelId] = useState("");
	const [labelKey, setLabelKey] = useState("");
	const [labelValue, setLabelValue] = useState("");
	const activeProjectName = projectName || project?.name || "";
	const activeProjectSlug = projectSlug || project?.slug || "";
	const selectedLabel = labelsQuery.data?.labels?.find(label => label.id === selectedLabelId) ?? null;
	const activeLabelKey = labelKey || selectedLabel?.key || "";
	const activeLabelValue = labelValue || selectedLabel?.value || "";
	const updateProjectMutation = useMutation({
		mutationFn: () => updateProject(projectRef || "", { name: activeProjectName, slug: activeProjectSlug }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const addMemberMutation = useMutation({
		mutationFn: () => addProjectMember(projectRef || "", { role: memberRole, userId: memberUserId }),
		onSuccess: () => {
			setMemberUserId("");

			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.members(projectRef) });
			}
		}
	});
	const removeMemberMutation = useMutation({
		mutationFn: (userId: string) => removeProjectMember(projectRef || "", userId),
		onSuccess: () => {
			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.members(projectRef) });
			}
		}
	});
	const updateMemberRoleMutation = useMutation({
		mutationFn: ({ userId, role }: { userId: string; role: ProjectMemberRole }) => updateProjectMemberRole(projectRef || "", userId, role),
		onSuccess: () => {
			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.members(projectRef) });
			}
		}
	});
	const saveLabelMutation = useMutation({
		mutationFn: () => {
			const body = { key: activeLabelKey, value: activeLabelValue };
			return selectedLabel ? updateProjectLabel(projectRef || "", selectedLabel.id, body) : createProjectLabel(projectRef || "", body);
		},
		onSuccess: data => {
			setSelectedLabelId(data.label.id);
			setLabelKey(data.label.key);
			setLabelValue(data.label.value);

			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.labels(projectRef) });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(projectRef) });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(projectRef) });
			}
		}
	});
	const deleteLabelMutation = useMutation({
		mutationFn: (labelId: string) => deleteProjectLabel(projectRef || "", labelId),
		onSuccess: () => {
			setSelectedLabelId("");
			setLabelKey("");
			setLabelValue("");

			if (projectRef) {
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.labels(projectRef) });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.probes(projectRef) });
				queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.checks(projectRef) });
			}
		}
	});
	const deleteProjectMutation = useMutation({
		mutationFn: () => deleteProject(projectRef || ""),
		onSuccess: () => {
			setSelectedProjectRef("");
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const memberRows: MemberRow[] = (membersQuery.data?.members ?? []).map(member => ({
		id: member.id,
		userId: member.userId,
		name: member.userId,
		email: member.userId,
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
	const labelRows: LabelRow[] = (labelsQuery.data?.labels ?? []).map(label => ({
		id: label.id,
		key: label.key,
		value: label.value,
		updatedAt: new Date(label.updatedAt).toLocaleString()
	}));
	const labelColumns: DataColumn<LabelRow>[] = [
		{ key: "key", label: "Key" },
		{ key: "value", label: "Value" },
		{ key: "updatedAt", label: "Updated" },
		{
			key: "delete",
			label: "Delete",
			render: row => (
				<Button variant="danger" size="sm" onClick={() => deleteLabelMutation.mutate(row.id)}>
					Delete
				</Button>
			)
		}
	];

	function selectLabel(label: ApiLabel) {
		setSelectedLabelId(label.id);
		setLabelKey(label.key);
		setLabelValue(label.value);
	}

	function startNewLabel() {
		setSelectedLabelId("");
		setLabelKey("");
		setLabelValue("");
	}

	function deleteCurrentProject() {
		if (!project || !window.confirm(`Delete organization ${project.name}?`)) {
			return;
		}

		deleteProjectMutation.mutate();
	}

	return (
		<PageStack>
			<ScreenHeader eyebrow="Team settings" title="Team" copy="Organization profile, member management, and destructive organization actions." />

			<Panel tone="glass" eyebrow="Organization" title="Org info">
				<div className={styles.orgInfoGrid}>
					<TextField label="Organization name" value={activeProjectName} disabled={!projectRef} onChange={event => setProjectName(event.currentTarget.value)} />
					<TextField label="Slug" value={activeProjectSlug} disabled={!projectRef} onChange={event => setProjectSlug(event.currentTarget.value)} />
				</div>
				<Button disabled={!projectRef || updateProjectMutation.isPending} onClick={() => updateProjectMutation.mutate()}>
					{updateProjectMutation.isPending ? "Saving" : "Save changes"}
				</Button>
			</Panel>

			<Panel tone="glass" eyebrow="Members" title="Member management">
				<div className={styles.formGridThree}>
					<TextField label="User ID" value={memberUserId} onChange={event => setMemberUserId(event.currentTarget.value)} />
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
					<Button disabled={!projectRef || !memberUserId || addMemberMutation.isPending} onClick={() => addMemberMutation.mutate()}>
						{addMemberMutation.isPending ? "Adding" : "Add member"}
					</Button>
				</div>
				<DataTable columns={memberColumns} rows={memberRows} getRowKey={row => row.id} />
			</Panel>

			<Panel tone="glass" eyebrow="Labels" title="Project labels">
				<div className={styles.formGridThree}>
					<TextField label="Key" value={activeLabelKey} onChange={event => setLabelKey(event.currentTarget.value)} />
					<TextField label="Value" value={activeLabelValue} onChange={event => setLabelValue(event.currentTarget.value)} />
					<Button disabled={!projectRef || !activeLabelKey || !activeLabelValue || saveLabelMutation.isPending} onClick={() => saveLabelMutation.mutate()}>
						{saveLabelMutation.isPending ? "Saving" : selectedLabel ? "Save label" : "Create label"}
					</Button>
				</div>
				<Button variant="outline" size="sm" onClick={startNewLabel}>
					New label
				</Button>
				<DataTable
					columns={labelColumns}
					rows={labelRows}
					getRowKey={row => row.id}
					selectedKey={selectedLabelId}
					onRowClick={row => selectLabel({ id: row.id, key: row.key, value: row.value, projectId: project?.id || "", createdAt: "", updatedAt: row.updatedAt })}
				/>
			</Panel>

			<Panel tone="deep" eyebrow="Danger zone" title="Dangerous organization actions">
				<div className={styles.dangerZoneGrid}>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Delete organization</h3>
						<p className={styles.warningCopy}>Delete this organization, disable future assignments, and revoke all probe registration tokens.</p>
						<Button variant="danger" disabled={!projectRef || deleteProjectMutation.isPending} onClick={deleteCurrentProject}>
							{deleteProjectMutation.isPending ? "Deleting" : "Delete organization"}
						</Button>
					</Surface>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Exit organization</h3>
						<p className={styles.warningCopy}>Leave this organization and remove your access to its probes, checks, alerts, and measurements.</p>
						<Button variant="outline">Exit organization</Button>
					</Surface>
				</div>
			</Panel>
		</PageStack>
	);
}
