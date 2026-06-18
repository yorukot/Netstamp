import { CreateProjectModal } from "@/features/project/components/CreateProjectModal";
import { pathForProjectSwitch } from "@/routes/routePaths";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { appFeatures } from "@/shared/config/features";
import { classNames } from "@/shared/utils/classNames";
import { Button, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, Select } from "@netstamp/ui";
import { FolderOpen, FolderPlus } from "@phosphor-icons/react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import styles from "./ProjectSwitcher.module.css";

const CREATE_PROJECT_VALUE = "__create_project__";

interface ProjectSwitcherProps {
	collapsed?: boolean;
	variant?: "sidebar" | "drawer";
}

function createProjectOptionLabel() {
	return (
		<span className={styles.projectCreateOption}>
			<FolderPlus size={17} weight="bold" aria-hidden="true" />
			<span>Create new project</span>
		</span>
	);
}

export function ProjectSwitcher({ collapsed = false, variant = "sidebar" }: ProjectSwitcherProps) {
	const { projectRef, projectsQuery, setSelectedProjectRef } = useCurrentProject();
	const location = useLocation();
	const navigate = useNavigate();
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [compactOpen, setCompactOpen] = useState(false);
	const projects = projectsQuery.data?.projects ?? [];

	function navigateAfterProjectChange(nextProjectRef: string) {
		const nextPath = pathForProjectSwitch(location.pathname, nextProjectRef);
		setCompactOpen(false);

		if (nextPath) {
			navigate(nextPath);
		}
	}

	function selectProject(value: string) {
		if (value === CREATE_PROJECT_VALUE) {
			if (!appFeatures.projectCreation) {
				return;
			}

			setCompactOpen(false);
			setCreateModalOpen(true);
			return;
		}

		setSelectedProjectRef(value);
		setCompactOpen(false);
		navigateAfterProjectChange(value);
	}

	return (
		<>
			<div className={classNames(styles.root, collapsed && styles.collapsed, variant === "drawer" && styles.drawer)}>
				<div className={styles.projectSelect}>
					<label className={styles.projectSelectField}>
						<span>project</span>
						<Select variant="compact" frameClassName={styles.projectFrame} className={styles.projectControl} value={projectRef || ""} onChange={event => selectProject(event.currentTarget.value)}>
							{projects.length ? (
								projects.map(project => (
									<option key={project.id} value={project.slug || project.id}>
										{project.name}
									</option>
								))
							) : (
								<option value="">No project</option>
							)}
							{appFeatures.projectCreation ? <option value={CREATE_PROJECT_VALUE}>{createProjectOptionLabel()}</option> : null}
						</Select>
					</label>
				</div>
				<PopoverRoot open={compactOpen} onOpenChange={setCompactOpen}>
					<div className={styles.projectCompact}>
						<PopoverTrigger asChild>
							<button type="button" className={styles.projectCompactButton} aria-label="Select project">
								<FolderOpen size={18} weight="bold" aria-hidden="true" />
							</button>
						</PopoverTrigger>
					</div>
					<PopoverPortal>
						<PopoverContent className={styles.projectPopover} align="center" side="right" sideOffset={10} collisionPadding={8}>
							<label className={styles.projectSelectField}>
								<span>project</span>
								<Select variant="compact" frameClassName={styles.projectFrame} className={styles.projectControl} value={projectRef || ""} onChange={event => selectProject(event.currentTarget.value)}>
									{projects.length ? (
										projects.map(project => (
											<option key={project.id} value={project.slug || project.id}>
												{project.name}
											</option>
										))
									) : (
										<option value="">No project</option>
									)}
									{appFeatures.projectCreation ? <option value={CREATE_PROJECT_VALUE}>{createProjectOptionLabel()}</option> : null}
								</Select>
							</label>
							{appFeatures.projectCreation ? (
								<Button
									className={styles.projectPopoverCreate}
									type="button"
									variant="ghost"
									size="sm"
									onClick={() => {
										setCompactOpen(false);
										setCreateModalOpen(true);
									}}
								>
									<FolderPlus size={16} weight="bold" aria-hidden="true" />
									Create new project
								</Button>
							) : null}
						</PopoverContent>
					</PopoverPortal>
				</PopoverRoot>
			</div>
			{createModalOpen && appFeatures.projectCreation ? <CreateProjectModal onClose={() => setCreateModalOpen(false)} onCreatedProject={navigateAfterProjectChange} /> : null}
		</>
	);
}
