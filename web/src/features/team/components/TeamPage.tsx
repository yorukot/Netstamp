import { members } from "@/features/team/data/members";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, DataTable, Panel, SelectField, Surface, TextField, type DataColumn } from "@netstamp/ui";
import { RoleSelect } from "./RoleSelect";
import styles from "./TeamPage.module.css";

interface MemberRow {
	name: string;
	email: string;
	role: string;
	lastActive: string;
}

const memberRows: MemberRow[] = members.map(([name, email, role, lastActive]) => ({ name, email, role, lastActive }));
const memberColumns: DataColumn<MemberRow>[] = [
	{ key: "name", label: "Name" },
	{ key: "email", label: "Email" },
	{ key: "role", label: "Role", render: row => <RoleSelect role={row.role} name={row.name} /> },
	{ key: "lastActive", label: "Last active" },
	{
		key: "delete",
		label: "Delete",
		render: () => (
			<Button variant="danger" size="sm">
				Delete
			</Button>
		)
	}
];

export function TeamPage() {
	return (
		<PageStack>
			<ScreenHeader eyebrow="Team settings" title="Team" copy="Organization profile, member management, and destructive organization actions." />

			<Panel tone="glass" eyebrow="Organization" title="Org info">
				<div className={styles.orgInfoGrid}>
					<TextField label="Organization name" defaultValue="Vector IX" />
					<TextField label="Slug" defaultValue="vector-ix" />
				</div>
				<Button>Save changes</Button>
			</Panel>

			<Panel tone="glass" eyebrow="Members" title="Member management">
				<div className={styles.formGridThree}>
					<TextField label="Email" defaultValue="sre@vector.example" />
					<SelectField
						label="Role"
						defaultValue="admin"
						options={[
							{ value: "owner", label: "Owner" },
							{ value: "admin", label: "Admin" },
							{ value: "member", label: "Member" }
						]}
					/>
					<Button>Add member</Button>
				</div>
				<DataTable columns={memberColumns} rows={memberRows} />
			</Panel>

			<Panel tone="deep" eyebrow="Danger zone" title="Dangerous organization actions">
				<div className={styles.dangerZoneGrid}>
					<Surface as="article" tone="danger" cut="md" padding="md">
						<h3>Delete organization</h3>
						<p className={styles.warningCopy}>Delete this organization, disable future assignments, and revoke all probe registration tokens.</p>
						<Button variant="danger">Delete organization</Button>
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
