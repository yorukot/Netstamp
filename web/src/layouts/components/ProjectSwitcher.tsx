import { CreateProjectModal } from "@/features/project/components/CreateProjectModal";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { Button, Select } from "@netstamp/ui";
import { useState } from "react";
import styles from "../AppShell.module.css";

export function ProjectSwitcher() {
	const { projectRef, projectsQuery, setSelectedProjectRef } = useCurrentProject();
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const projects = projectsQuery.data?.projects ?? [];

	return (
		<>
			<div className={styles.projectSelect}>
				<label className={styles.projectSelectField}>
					<span>project</span>
					<Select
						variant="compact"
						frameClassName={styles.projectFrame}
						className={styles.projectControl}
						value={projectRef || ""}
						onChange={event => setSelectedProjectRef(event.currentTarget.value)}
					>
						{projects.length ? (
							projects.map(project => (
								<option key={project.id} value={project.slug || project.id}>
									{project.name}
								</option>
							))
						) : (
							<option value="">No project</option>
						)}
					</Select>
				</label>
				<Button className={styles.projectCreateButton} type="button" variant="outline" size="sm" onClick={() => setCreateModalOpen(true)}>
					Create new project
				</Button>
			</div>
			{createModalOpen ? <CreateProjectModal onClose={() => setCreateModalOpen(false)} /> : null}
		</>
	);
}
