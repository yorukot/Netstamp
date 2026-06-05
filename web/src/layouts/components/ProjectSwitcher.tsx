import { CreateProjectModal } from "@/features/project/components/CreateProjectModal";
import styles from "@/layouts/AppShell.module.css";
import { pathForProjectSwitch } from "@/routes/routePaths";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { Select } from "@netstamp/ui";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

const CREATE_PROJECT_VALUE = "__create_project__";

export function ProjectSwitcher() {
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
			setCreateModalOpen(true);
			return;
		}

		setSelectedProjectRef(value);
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
			{createModalOpen ? <CreateProjectModal onClose={() => setCreateModalOpen(false)} onCreatedProject={navigateAfterProjectChange} /> : null}
		</>
	);
}
