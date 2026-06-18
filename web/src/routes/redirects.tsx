import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { Navigate as RouterNavigate, useParams } from "react-router-dom";
import { pathForCheckDetail, pathForLabelDetail, pathForProbeDetail, pathForRoute } from "./routePaths";
import type { ProjectAppRoute } from "./routeTypes";

export function DefaultProjectRedirect({ route = "dashboard" }: { route?: ProjectAppRoute }) {
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (!projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <RouterNavigate to={pathForRoute(route, { projectRef })} replace />;
}

export function LegacyProbeDetailRedirect() {
	const { probeId = "" } = useParams();
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (!projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <RouterNavigate to={probeId ? pathForProbeDetail(projectRef, probeId) : pathForRoute("probes", { projectRef })} replace />;
}

export function LegacyLabelDetailRedirect() {
	const { labelId = "" } = useParams();
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (!projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <RouterNavigate to={labelId ? pathForLabelDetail(projectRef, labelId) : pathForRoute("labels", { projectRef })} replace />;
}

export function LegacyCheckDetailRedirect() {
	const { checkId = "" } = useParams();
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (!projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <RouterNavigate to={checkId ? pathForCheckDetail(projectRef, checkId) : pathForRoute("checks", { projectRef })} replace />;
}
