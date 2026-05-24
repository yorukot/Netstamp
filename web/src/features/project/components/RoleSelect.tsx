import type { ProjectMemberRole } from "@/shared/api/types";
import { classNames } from "@/shared/utils/classNames";
import { Select } from "@netstamp/ui";
import styles from "./RoleSelect.module.css";

interface RoleSelectProps {
	role: string;
	name: string;
	disabled?: boolean;
	onRoleChange?: (role: ProjectMemberRole) => void;
}

const roleOptions = [
	{ value: "owner", label: "Owner", disabled: true },
	{ value: "admin", label: "Admin" },
	{ value: "editor", label: "Editor" },
	{ value: "viewer", label: "Viewer" }
];

export function RoleSelect({ role, name, disabled, onRoleChange }: RoleSelectProps) {
	const selectedRole = role.toLowerCase();
	const roleClass = styles[selectedRole as keyof typeof styles] || styles.member;

	return (
		<Select
			variant="compact"
			frameClassName={classNames(styles.frame, roleClass)}
			className={styles.select}
			value={selectedRole}
			disabled={disabled}
			aria-label={`Change role for ${name}`}
			onChange={event => onRoleChange?.(event.currentTarget.value as ProjectMemberRole)}
		>
			{roleOptions.map(option => (
				<option key={option.value} value={option.value} disabled={option.disabled}>
					{option.label}
				</option>
			))}
		</Select>
	);
}
