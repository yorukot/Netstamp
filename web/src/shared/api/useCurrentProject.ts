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

export function useCurrentProject() {
	const context = useContext(CurrentProjectContext);
	const projectsQuery = useQuery(projectQueries.list());
	const projects = projectsQuery.data?.projects ?? [];
	const listProject = projects.find(item => item.slug === context?.selectedProjectRef || item.id === context?.selectedProjectRef) ?? projects[0] ?? null;
	const listProjectRef = listProject?.slug || listProject?.id || null;
	const projectDetailQuery = useQuery({
		...projectQueries.detail(listProjectRef || ""),
		enabled: Boolean(listProjectRef)
	});
	const project = projectDetailQuery.data?.project ?? listProject;
	const projectRef = project?.slug || project?.id || null;

	return {
		project,
		projectRef,
		projectDetailQuery,
		projectsQuery,
		selectedProjectRef: context?.selectedProjectRef ?? "",
		setSelectedProjectRef: context?.setSelectedProjectRef ?? (() => undefined)
	};
}
