import { useQuery } from "@tanstack/react-query";
import { createContext, createElement, type ReactNode, useContext, useMemo, useState } from "react";
import { projectQueries } from "./queries";

interface CurrentProjectContextValue {
	selectedProjectRef: string;
	setSelectedProjectRef: (ref: string) => void;
}

const CurrentProjectContext = createContext<CurrentProjectContextValue | null>(null);

export function CurrentProjectProvider({ children }: { children: ReactNode }) {
	const [selectedProjectRef, setSelectedProjectRef] = useState("");
	const value = useMemo(() => ({ selectedProjectRef, setSelectedProjectRef }), [selectedProjectRef]);

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
