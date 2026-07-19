import type { ProjectMemberRole } from "@/shared/api/types";
import { classNames } from "@/shared/utils/classNames";
import { Select } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
import styles from "./RoleSelect.module.css";

interface RoleSelectProps {
	role: string;
	name: string;
	disabled?: boolean;
	onRoleChange?: (role: ProjectMemberRole) => void;
}

export function RoleSelect({ role, name, disabled, onRoleChange }: RoleSelectProps) {
	const { t } = useTranslation("project");
	const selectedRole = role.toLowerCase();
	const roleClass = styles[selectedRole as keyof typeof styles] || styles.member;
	const roleOptions = [
		{ value: "owner", label: t("roles.owner"), disabled: true },
		{ value: "admin", label: t("roles.admin") },
		{ value: "editor", label: t("roles.editor") },
		{ value: "viewer", label: t("roles.viewer") }
	];

	return (
		<Select
			variant="compact"
			frameClassName={classNames(styles.frame, roleClass)}
			className={styles.select}
			value={selectedRole}
			disabled={disabled}
			aria-label={t("members.changeRole", { name })}
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
