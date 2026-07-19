import { CreateProjectModal } from "@/features/project/components/CreateProjectModal";
import { pathForProjectSwitch } from "@/routes/routePaths";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { appFeatures } from "@/shared/config/features";
import { classNames } from "@/shared/utils/classNames";
import { Select } from "@netstamp/ui";
import { FolderOpenIcon } from "@phosphor-icons/react/dist/csr/FolderOpen";
import { FolderPlusIcon } from "@phosphor-icons/react/dist/csr/FolderPlus";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate } from "react-router-dom";
import styles from "./ProjectSwitcher.module.css";

const CREATE_PROJECT_VALUE = "__create_project__";

interface ProjectSwitcherProps {
	collapsed?: boolean;
	variant?: "sidebar" | "drawer";
}

const createProjectOptionLabel = (label: string) => {
	return (
		<span className={styles.projectCreateOption}>
			<FolderPlusIcon size={17} weight="bold" aria-hidden="true" focusable="false" />
			<span>{label}</span>
		</span>
	);
};

const projectOptionLabel = (name: string) => {
	return (
		<span className={styles.projectOptionLabel}>
			<FolderOpenIcon className={styles.projectOptionIcon} size={18} weight="bold" aria-hidden="true" focusable="false" />
			<span className={styles.projectOptionName}>{name}</span>
		</span>
	);
};

export function ProjectSwitcher({ collapsed = false, variant = "sidebar" }: ProjectSwitcherProps) {
	const { t } = useTranslation("project");
	const { projectRef, projectsQuery, setSelectedProjectRef } = useCurrentProject();
	const location = useLocation();
	const navigate = useNavigate();
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const projects = projectsQuery.data?.projects ?? [];

	function navigateAfterProjectChange(nextProjectRef: string) {
		const nextPath = pathForProjectSwitch(location.pathname, nextProjectRef);

		if (nextPath) {
			navigate(nextPath);
		}
	}

	function selectProject(value: string) {
		if (value === CREATE_PROJECT_VALUE) {
			if (!appFeatures.projectCreation) {
				return;
			}

			setCreateModalOpen(true);
			return;
		}

		setSelectedProjectRef(value);
		navigateAfterProjectChange(value);
	}

	return (
		<>
			<div className={classNames(styles.root, collapsed && styles.collapsed, variant === "drawer" && styles.drawer)}>
				<div className={styles.projectSelect}>
					<div className={styles.projectSelectField}>
						<Select
							aria-label={t("switcher.select")}
							variant="compact"
							frameClassName={styles.projectFrame}
							menuClassName={classNames("ns-theme-dark", styles.projectMenu)}
							className={styles.projectControl}
							value={projectRef || ""}
							onChange={event => selectProject(event.currentTarget.value)}
						>
							{projects.length ? (
								projects.map(project => (
									<option key={project.id} value={project.slug || project.id}>
										{projectOptionLabel(project.name)}
									</option>
								))
							) : (
								<option value="">{projectOptionLabel(t("switcher.noProject"))}</option>
							)}
							{appFeatures.projectCreation ? <option value={CREATE_PROJECT_VALUE}>{createProjectOptionLabel(t("switcher.createNew"))}</option> : null}
						</Select>
					</div>
				</div>
			</div>
			{createModalOpen && appFeatures.projectCreation ? <CreateProjectModal onClose={() => setCreateModalOpen(false)} onCreatedProject={navigateAfterProjectChange} /> : null}
		</>
	);
}
