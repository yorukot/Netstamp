import { CreateProjectModal } from "@/features/project/components/CreateProjectModal";
import styles from "@/layouts/AppShell.module.css";
import { pathForProjectSwitch } from "@/routes/routePaths";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { classNames } from "@/shared/utils/classNames";
import { Button, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, Select } from "@netstamp/ui";
import { FolderOpen } from "@phosphor-icons/react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

const CREATE_PROJECT_VALUE = "__create_project__";

export function ProjectSwitcher() {
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
						<option value={CREATE_PROJECT_VALUE}>Create new project</option>
					</Select>
				</label>
			</div>
			<PopoverRoot open={compactOpen} onOpenChange={setCompactOpen}>
				<div className={styles.projectCompact}>
					<PopoverTrigger asChild>
						<button type="button" className={classNames("ns-cut-frame", styles.projectCompactButton)} aria-label="Select project">
							<FolderOpen size={18} weight="bold" aria-hidden="true" />
						</button>
					</PopoverTrigger>
				</div>
				<PopoverPortal>
					<PopoverContent className={classNames("ns-cut-frame", styles.projectPopover)} align="center" side="right" sideOffset={10} collisionPadding={8}>
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
								<option value={CREATE_PROJECT_VALUE}>Create new project</option>
							</Select>
						</label>
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
							Create new project
						</Button>
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>
			{createModalOpen ? <CreateProjectModal onClose={() => setCreateModalOpen(false)} onCreatedProject={navigateAfterProjectChange} /> : null}
		</>
	);
}
