import { useCreateProjectMutation } from "@/shared/api/mutations";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import type { ApiProject } from "@/shared/api/types";
import { useProjectSelection } from "@/shared/api/useCurrentProject";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { Button, DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle, TextField } from "@netstamp/ui";
import { useQueryClient } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useCallback, useEffect, useId, useState } from "react";
import styles from "./CreateProjectModal.module.css";

interface CreateProjectModalProps {
	onClose: () => void;
	onCreatedProject?: (projectRef: string) => void;
}

interface ProjectListCache {
	projects: ApiProject[];
}

const maxProjectSlugLength = 64;

function slugifyProjectName(name: string) {
	return name
		.toLowerCase()
		.trim()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-|-$/g, "")
		.slice(0, maxProjectSlugLength);
}

export function CreateProjectModal({ onClose, onCreatedProject }: CreateProjectModalProps) {
	const titleId = useId();
	const descriptionId = useId();
	const queryClient = useQueryClient();
	const createProjectMutation = useCreateProjectMutation();
	const { setSelectedProjectRef } = useProjectSelection();
	const [projectName, setProjectName] = useState("");
	const [projectSlug, setProjectSlug] = useState("");
	const [slugEdited, setSlugEdited] = useState(false);
	const projectNameTrimmed = projectName.trim();
	const projectSlugTrimmed = projectSlug.trim();
	const canCreate = Boolean(projectNameTrimmed && projectSlugTrimmed && !createProjectMutation.isPending);

	const closeModal = useCallback(() => {
		if (createProjectMutation.isPending) {
			return;
		}

		onClose();
	}, [createProjectMutation.isPending, onClose]);

	useEffect(() => {
		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		return () => {
			document.body.style.overflow = previousOverflow;
		};
	}, []);

	function updateProjectName(value: string) {
		setProjectName(value);

		if (!slugEdited) {
			setProjectSlug(slugifyProjectName(value));
		}
	}

	function updateProjectSlug(value: string) {
		setSlugEdited(true);
		setProjectSlug(slugifyProjectName(value));
	}

	async function submitProject(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		if (!canCreate) {
			return;
		}

		try {
			const data = await createProjectMutation.mutateAsync({
				name: projectNameTrimmed,
				slug: projectSlugTrimmed
			});
			const createdProjectRef = data.project.slug || data.project.id;

			queryClient.setQueryData<ProjectListCache>(apiQueryKeys.projects.list(), current => ({
				projects: [data.project, ...(current?.projects ?? []).filter(project => project.id !== data.project.id)]
			}));
			setSelectedProjectRef(createdProjectRef);
			onCreatedProject?.(createdProjectRef);
			pushToast({
				title: "Project created",
				message: `${data.project.name} is now selected.`,
				tone: "success"
			});
			onClose();
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Project could not be created.");
		}
	}

	if (typeof document === "undefined") {
		return null;
	}

	return (
		<DialogRoot
			open
			onOpenChange={open => {
				if (!open) {
					closeModal();
				}
			}}
		>
			<DialogPortal>
				<DialogOverlay className={styles.overlay} onMouseDown={closeModal}>
					<div className={styles.lineTop} aria-hidden="true" />
					<div className={styles.lineBottom} aria-hidden="true" />
					<DialogContent asChild aria-describedby={descriptionId}>
						<section className={styles.dialog} onMouseDown={event => event.stopPropagation()}>
							<div className={styles.header}>
								<span>Project registry</span>
								<DialogTitle asChild>
									<strong id={titleId}>Create project</strong>
								</DialogTitle>
								<DialogDescription asChild>
									<p id={descriptionId}>Add a workspace for probes, checks, and members.</p>
								</DialogDescription>
							</div>

							<form className={styles.form} onSubmit={submitProject}>
								<TextField label="Project name" value={projectName} onChange={event => updateProjectName(event.currentTarget.value)} autoComplete="off" autoFocus required />
								<TextField
									label="Slug"
									value={projectSlug}
									onChange={event => updateProjectSlug(event.currentTarget.value)}
									autoComplete="off"
									maxLength={maxProjectSlugLength}
									helper="Lowercase letters, numbers, and hyphens."
									required
								/>

								<div className={styles.actions}>
									<Button type="button" variant="ghost" disabled={createProjectMutation.isPending} onClick={closeModal}>
										Cancel
									</Button>
									<Button type="submit" disabled={!canCreate}>
										{createProjectMutation.isPending ? "Creating" : "Create project"}
									</Button>
								</div>
							</form>
						</section>
					</DialogContent>
				</DialogOverlay>
			</DialogPortal>
		</DialogRoot>
	);
}
