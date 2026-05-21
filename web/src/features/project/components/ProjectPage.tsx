import {
	useAddProjectMemberMutation,
	useDeleteProjectLabelMutation,
	useDeleteProjectMutation,
	useRemoveProjectMemberMutation,
	useSaveProjectLabelMutation,
	useUpdateProjectMemberRoleMutation,
	useUpdateProjectMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiLabel, ProjectMemberRole } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, DataTable, Panel, SelectField, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { RoleSelect } from "./RoleSelect";
import styles from "./ProjectPage.module.css";

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

export function ProjectPage() {
	const { project, projectRef, setSelectedProjectRef } = useCurrentProject();
	const updateProjectMutation = useUpdateProjectMutation(projectRef);
	const addMemberMutation = useAddProjectMemberMutation(projectRef);
	const removeMemberMutation = useRemoveProjectMemberMutation(projectRef);
	const updateMemberRoleMutation = useUpdateProjectMemberRoleMutation(projectRef);
	const saveLabelMutation = useSaveProjectLabelMutation(projectRef);
	const deleteLabelMutation = useDeleteProjectLabelMutation(projectRef);
	const deleteProjectMutation = useDeleteProjectMutation(projectRef);
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
	const [memberEmail, setMemberEmail] = useState("");
	const [memberRole, setMemberRole] = useState<ProjectMemberRole>("viewer");
	const [selectedLabelId, setSelectedLabelId] = useState("");
	const [labelKey, setLabelKey] = useState("");
	const [labelValue, setLabelValue] = useState("");
	const activeProjectName = projectName || project?.name || "";
	const activeProjectSlug = projectSlug || project?.slug || "";
	const selectedLabel = labelsQuery.data?.labels?.find(label => label.id === selectedLabelId) ?? null;
	const activeLabelKey = labelKey || selectedLabel?.key || "";
	const activeLabelValue = labelValue || selectedLabel?.value || "";
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

	function saveCurrentLabel() {
		saveLabelMutation.mutate(
			{ labelId: selectedLabel?.id, body: { key: activeLabelKey, value: activeLabelValue } },
			{
				onSuccess: data => {
					setSelectedLabelId(data.label.id);
					setLabelKey(data.label.key);
					setLabelValue(data.label.value);
				}
			}
		);
	}

	function deleteCurrentLabel(labelId: string) {
		deleteLabelMutation.mutate(labelId, {
			onSuccess: () => {
				setSelectedLabelId("");
				setLabelKey("");
				setLabelValue("");
			}
		});
	}

	function deleteCurrentProject() {
		if (!project || !window.confirm(`Delete project ${project.name}?`)) {
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
				<Button variant="danger" size="sm" onClick={() => deleteCurrentLabel(row.id)}>
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

			<Panel tone="glass" eyebrow="Labels" title="Project labels">
				<div className={styles.formGridThree}>
					<TextField label="Key" value={activeLabelKey} onChange={event => setLabelKey(event.currentTarget.value)} />
					<TextField label="Value" value={activeLabelValue} onChange={event => setLabelValue(event.currentTarget.value)} />
					<Button disabled={!projectRef || !activeLabelKey || !activeLabelValue || saveLabelMutation.isPending} onClick={saveCurrentLabel}>
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

			<Panel tone="deep" eyebrow="Danger zone" title="Dangerous project actions">
				<div className={styles.dangerZoneGrid}>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Delete project</h3>
						<p className={styles.warningCopy}>Delete this project, disable future assignments, and revoke all probe registration tokens.</p>
						<Button variant="danger" disabled={!projectRef || deleteProjectMutation.isPending} onClick={deleteCurrentProject}>
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
