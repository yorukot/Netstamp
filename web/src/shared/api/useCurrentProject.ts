import { useQuery } from "@tanstack/react-query";
import { createContext, createElement, type ReactNode, useContext, useEffect, useMemo, useState } from "react";
import { projectQueries } from "./queries";

const selectedProjectStorageKey = "netstamp:selected-project-ref";

interface CurrentProjectContextValue {
	selectedProjectRef: string;
	setSelectedProjectRef: (ref: string) => void;
}

const CurrentProjectContext = createContext<CurrentProjectContextValue | null>(null);

function readStoredProjectRef() {
	try {
		return window.localStorage.getItem(selectedProjectStorageKey) ?? "";
	} catch {
		return "";
	}
}

export function CurrentProjectProvider({ children }: { children: ReactNode }) {
	const [selectedProjectRef, setSelectedProjectRef] = useState(readStoredProjectRef);
	const value = useMemo(() => ({ selectedProjectRef, setSelectedProjectRef }), [selectedProjectRef]);

	useEffect(() => {
		try {
			if (selectedProjectRef) {
				window.localStorage.setItem(selectedProjectStorageKey, selectedProjectRef);
				return;
			}

			window.localStorage.removeItem(selectedProjectStorageKey);
		} catch {
			// Storage access can fail in private browsing or restricted embeds.
		}
	}, [selectedProjectRef]);

	return createElement(CurrentProjectContext.Provider, { value }, children);
}

export function useProjectSelection() {
	const context = useContext(CurrentProjectContext);

	return {
		selectedProjectRef: context?.selectedProjectRef ?? "",
		setSelectedProjectRef: context?.setSelectedProjectRef ?? (() => undefined)
	};
}

export function useCurrentProject() {
	const { selectedProjectRef, setSelectedProjectRef } = useProjectSelection();
	const projectsQuery = useQuery(projectQueries.list());
	const projects = projectsQuery.data?.projects ?? [];
	const selectedProject = projects.find(item => item.slug === selectedProjectRef || item.id === selectedProjectRef) ?? null;
	const listProject = selectedProject ?? projects[0] ?? null;
	const selectedProjectRefFallback = selectedProjectRef && (projectsQuery.isPending || projectsQuery.isError) ? selectedProjectRef : "";
	const listProjectRef = selectedProject?.slug || selectedProject?.id || selectedProjectRefFallback || listProject?.slug || listProject?.id || null;
	const projectDetailQuery = useQuery({
		...projectQueries.detail(listProjectRef || ""),
		enabled: Boolean(listProjectRef)
	});
	const project = projectDetailQuery.data?.project ?? selectedProject ?? listProject;
	const projectRef = project?.slug || project?.id || listProjectRef;

	return {
		project,
		projectRef,
		projectDetailQuery,
		projectsQuery,
		selectedProjectRef,
		setSelectedProjectRef
	};
}
