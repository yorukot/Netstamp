import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { Select } from "@netstamp/ui";
import styles from "../AppShell.module.css";

export function ProjectSwitcher() {
	const { projectRef, projectsQuery, setSelectedProjectRef } = useCurrentProject();
	const projects = projectsQuery.data?.projects ?? [];

	return (
		<label className={styles.projectSelect}>
			<span>project</span>
			<Select variant="compact" frameClassName={styles.projectFrame} className={styles.projectControl} value={projectRef || ""} onChange={event => setSelectedProjectRef(event.currentTarget.value)}>
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
	);
}
