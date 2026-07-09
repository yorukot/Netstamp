import { useSession } from "@/features/auth/session/SessionContext";
import { AppShell } from "@/layouts/AppShell";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject, useProjectSelection } from "@/shared/api/useCurrentProject";
import { appFeatures } from "@/shared/config/features";
import { Spinner } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { Outlet, Navigate as RouterNavigate, useLocation, useParams } from "react-router-dom";
import { AuthPage, DashboardPage, ForgotPasswordPage, OnboardingPage, ResetPasswordPage, VerifyEmailPage } from "./lazyRoutes";
import { useRouteNavigate } from "./navigation";
import { DefaultProjectRedirect } from "./redirects";
import { pathForRoute } from "./routePaths";
import { lazyRoute } from "./routeSuspense";

interface AuthRouteProps {
	mode: "login" | "register";
}

function routeSpinner(label: string) {
	return <Spinner label={label} layout="page" size="lg" />;
}

export function AuthRoute({ mode }: AuthRouteProps) {
	const { loading, session } = useSession();
	const navigate = useRouteNavigate();

	if (loading) {
		return routeSpinner("Loading session");
	}

	if (session) {
		return <RouterNavigate to={pathForRoute("dashboard")} replace />;
	}

	if (mode === "register" && !appFeatures.registration) {
		return <RouterNavigate to={pathForRoute("login")} replace />;
	}

	return lazyRoute(<AuthPage mode={mode} navigate={navigate} />);
}

export function PasswordResetRoute({ mode }: { mode: "forgot" | "reset" }) {
	const { loading, session } = useSession();
	const navigate = useRouteNavigate();

	if (loading) {
		return routeSpinner("Loading session");
	}

	if (session) {
		return <RouterNavigate to={pathForRoute("dashboard")} replace />;
	}

	return mode === "forgot" ? lazyRoute(<ForgotPasswordPage />) : lazyRoute(<ResetPasswordPage navigate={navigate} />);
}

export function EmailVerificationRoute() {
	const { loading, session } = useSession();
	const navigate = useRouteNavigate();

	if (loading) {
		return routeSpinner("Loading session");
	}

	if (session) {
		return <RouterNavigate to={pathForRoute("dashboard")} replace />;
	}

	return lazyRoute(<VerifyEmailPage navigate={navigate} />);
}

export function OnboardingRoute() {
	const { loading, session } = useSession();
	const location = useLocation();
	const navigate = useRouteNavigate();

	if (loading) {
		return routeSpinner("Loading session");
	}

	if (!session) {
		return <RouterNavigate to={pathForRoute("login")} replace state={{ from: location }} />;
	}

	return lazyRoute(<OnboardingPage navigate={navigate} />);
}

export function DashboardRoute() {
	return lazyRoute(<DashboardPage />);
}

export function ProtectedAppShell() {
	const { loading, session } = useSession();
	const location = useLocation();

	if (loading) {
		return routeSpinner("Loading session");
	}

	if (!session) {
		return <RouterNavigate to={pathForRoute("login")} replace state={{ from: location }} />;
	}

	return <ProjectAppShell />;
}

function ProjectAppShell() {
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return routeSpinner("Loading projects");
	}

	if (projectsQuery.isSuccess && !projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <AppShell />;
}

export function ProjectRouteBoundary() {
	const { projectRef = "" } = useParams();
	const { selectedProjectRef, setSelectedProjectRef } = useProjectSelection();
	const projectsQuery = useQuery(projectQueries.list());
	const projects = projectsQuery.data?.projects ?? [];
	const matchedProject = projects.find(project => project.slug === projectRef || project.id === projectRef);

	useEffect(() => {
		if (projectRef && selectedProjectRef !== projectRef) {
			setSelectedProjectRef(projectRef);
		}
	}, [projectRef, selectedProjectRef, setSelectedProjectRef]);

	if (!projectRef) {
		return <DefaultProjectRedirect />;
	}

	if (projectsQuery.isPending) {
		return routeSpinner("Loading projects");
	}

	if (projectsQuery.isSuccess && !matchedProject) {
		const fallbackProject = projects[0];
		const fallbackProjectRef = fallbackProject?.slug || fallbackProject?.id || "";

		if (!fallbackProjectRef) {
			return <RouterNavigate to={pathForRoute("onboarding")} replace />;
		}

		return <RouterNavigate to={pathForRoute("dashboard", { projectRef: fallbackProjectRef })} replace />;
	}

	if (selectedProjectRef !== projectRef) {
		return routeSpinner("Loading project");
	}

	return <Outlet />;
}
