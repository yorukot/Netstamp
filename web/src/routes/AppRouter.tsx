import { useSession } from "@/features/auth/session/SessionContext";
import { SessionProvider } from "@/features/auth/session/SessionProvider";
import { AppShell } from "@/layouts/AppShell";
import { projectQueries } from "@/shared/api/queries";
import { queryClient } from "@/shared/api/queryClient";
import { CurrentProjectProvider, useCurrentProject, useProjectSelection } from "@/shared/api/useCurrentProject";
import { ConfirmProvider } from "@/shared/components/ConfirmProvider";
import { ToastProvider } from "@/shared/components/ToastProvider";
import { appFeatures } from "@/shared/config/features";
import { docsUrl } from "@/shared/config/publicLinks";
import { TrackingConsentBanner } from "@/shared/tracking/TrackingConsentBanner";
import { QueryClientProvider, useQuery } from "@tanstack/react-query";
import { lazy, Suspense, useEffect, type ReactNode } from "react";
import { createBrowserRouter, Outlet, Navigate as RouterNavigate, RouterProvider, useLocation, useNavigate, useParams } from "react-router-dom";
import { pathForCheckDetail, pathForLabelDetail, pathForProbeDetail, pathForRoute, projectRoutePath } from "./routePaths";
import type { Navigate, ProjectAppRoute } from "./routeTypes";

const AuthPage = lazy(() => import("@/features/auth/components/AuthPage").then(module => ({ default: module.AuthPage })));
const AlertsPage = lazy(() => import("@/features/alerts/components/AlertsPage").then(module => ({ default: module.AlertsPage })));
const HomePage = lazy(() => import("@/features/home/components/HomePage").then(module => ({ default: module.HomePage })));
const OnboardingPage = lazy(() => import("@/features/auth/components/OnboardingPage").then(module => ({ default: module.OnboardingPage })));
const ChecksPage = lazy(() => import("@/features/checks/components/ChecksPage").then(module => ({ default: module.ChecksPage })));
const DashboardPage = lazy(() => import("@/features/dashboard/components/DashboardPage").then(module => ({ default: module.DashboardPage })));
const InsightPage = lazy(() => import("@/features/insight/components/InsightPage").then(module => ({ default: module.InsightPage })));
const LabelsPage = lazy(() => import("@/features/labels/components/LabelsPage").then(module => ({ default: module.LabelsPage })));
const MembersPage = lazy(() => import("@/features/project/components/MembersPage").then(module => ({ default: module.MembersPage })));
const NewProbeDrawer = lazy(() => import("@/features/probes/components/NewProbeDrawer").then(module => ({ default: module.NewProbeDrawer })));
const ProbesPage = lazy(() => import("@/features/probes/components/ProbesPage").then(module => ({ default: module.ProbesPage })));
const SettingsPage = lazy(() => import("@/features/settings/components/SettingsPage").then(module => ({ default: module.SettingsPage })));
const ProjectPage = lazy(() => import("@/features/project/components/ProjectPage").then(module => ({ default: module.ProjectPage })));

function lazyRoute(element: ReactNode) {
	return <Suspense fallback={null}>{element}</Suspense>;
}

function useRouteNavigate(defaultProjectRef?: string | null): Navigate {
	const navigate = useNavigate();

	return (route, options) => navigate(pathForRoute(route, { projectRef: options?.projectRef ?? defaultProjectRef }));
}

interface AuthRouteProps {
	mode: "login" | "register";
}

function AuthRoute({ mode }: AuthRouteProps) {
	const { loading, session } = useSession();
	const navigate = useRouteNavigate();

	if (loading) {
		return null;
	}

	if (session) {
		return <RouterNavigate to={pathForRoute("dashboard")} replace />;
	}

	if (mode === "register" && !appFeatures.registration) {
		return <RouterNavigate to={pathForRoute("login")} replace />;
	}

	return lazyRoute(<AuthPage mode={mode} navigate={navigate} />);
}

function OnboardingRoute() {
	const { loading, session } = useSession();
	const location = useLocation();
	const navigate = useRouteNavigate();

	if (loading) {
		return null;
	}

	if (!session) {
		return <RouterNavigate to={pathForRoute("login")} replace state={{ from: location }} />;
	}

	return lazyRoute(<OnboardingPage navigate={navigate} />);
}

function DashboardRoute() {
	const { projectRef } = useCurrentProject();
	const navigate = useRouteNavigate(projectRef);

	return lazyRoute(<DashboardPage navigate={navigate} />);
}

function ProtectedAppShell() {
	const { loading, session } = useSession();
	const location = useLocation();

	if (loading) {
		return null;
	}

	if (!session) {
		return <RouterNavigate to={pathForRoute("login")} replace state={{ from: location }} />;
	}

	return <ProjectAppShell />;
}

function ProjectAppShell() {
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (projectsQuery.isSuccess && !projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <AppShell />;
}

function DefaultProjectRedirect({ route = "dashboard" }: { route?: ProjectAppRoute }) {
	const { projectRef, projectsQuery } = useCurrentProject();

	if (projectsQuery.isPending) {
		return null;
	}

	if (!projectRef) {
		return <RouterNavigate to={pathForRoute("onboarding")} replace />;
	}

	return <RouterNavigate to={pathForRoute(route, { projectRef })} replace />;
}

function LegacyProbeDetailRedirect() {
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

function LegacyLabelDetailRedirect() {
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

function LegacyCheckDetailRedirect() {
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

function ProjectRouteBoundary() {
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
		return null;
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
		return null;
	}

	return <Outlet />;
}

function RouteFrame() {
	return (
		<>
			<Outlet />
			<TrackingConsentBanner />
		</>
	);
}

function PublicDocsRedirect({ basePath }: { basePath: "/docs/" | "/storybook/" }) {
	const location = useLocation();
	const params = useParams();
	const wildcard = params["*"] ?? "";
	const targetPath = wildcard ? `${basePath}${wildcard}` : basePath;
	const target = `${docsUrl(targetPath)}${location.search}${location.hash}`;

	useEffect(() => {
		window.location.replace(target);
	}, [target]);

	return (
		<div style={{ minHeight: "100svh", display: "grid", placeItems: "center", padding: "1rem" }}>
			<a href={target}>Open Netstamp docs</a>
		</div>
	);
}

const router = createBrowserRouter([
	{
		element: <RouteFrame />,
		children: [
			{ path: pathForRoute("landing"), element: lazyRoute(<HomePage />) },
			{ path: "docs", element: <PublicDocsRedirect basePath="/docs/" /> },
			{ path: "docs/*", element: <PublicDocsRedirect basePath="/docs/" /> },
			{ path: "storybook", element: <PublicDocsRedirect basePath="/storybook/" /> },
			{ path: "storybook/*", element: <PublicDocsRedirect basePath="/storybook/" /> },
			{ path: pathForRoute("login"), element: <AuthRoute mode="login" /> },
			{ path: pathForRoute("register"), element: <AuthRoute mode="register" /> },
			{ path: pathForRoute("onboarding"), element: <OnboardingRoute /> },
			{
				element: <ProtectedAppShell />,
				children: [
					{ path: "dashboard", element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "probes",
						children: [
							{ index: true, element: <DefaultProjectRedirect route="probes" /> },
							{ path: "new", element: <DefaultProjectRedirect route="newProbe" /> },
							{ path: ":probeId", element: <LegacyProbeDetailRedirect /> }
						]
					},
					{
						path: "labels",
						children: [
							{ index: true, element: <DefaultProjectRedirect route="labels" /> },
							{ path: ":labelId", element: <LegacyLabelDetailRedirect /> }
						]
					},
					{
						path: "checks",
						children: [
							{ index: true, element: <DefaultProjectRedirect route="checks" /> },
							{ path: ":checkId", element: <LegacyCheckDetailRedirect /> }
						]
					},
					{ path: "insight", element: <DefaultProjectRedirect route="insight" /> },
					{ path: "alerts", element: <DefaultProjectRedirect route="alerts" /> },
					{ path: "members", element: <DefaultProjectRedirect route="members" /> },
					{ path: "project", element: <DefaultProjectRedirect route="project" /> },
					{ path: "projects", element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "projects/:projectRef",
						element: <ProjectRouteBoundary />,
						children: [
							{ index: true, element: <RouterNavigate to={projectRoutePath("dashboard")} replace /> },
							{ path: projectRoutePath("dashboard"), element: <DashboardRoute /> },
							{
								path: projectRoutePath("probes"),
								element: lazyRoute(<ProbesPage />),
								children: [
									{ path: "new", element: lazyRoute(<NewProbeDrawer />) },
									{ path: ":probeId", element: null }
								]
							},
							{ path: projectRoutePath("labels"), element: lazyRoute(<LabelsPage />) },
							{ path: `${projectRoutePath("labels")}/:labelId`, element: lazyRoute(<LabelsPage />) },
							{ path: projectRoutePath("checks"), element: lazyRoute(<ChecksPage />) },
							{ path: `${projectRoutePath("checks")}/:checkId`, element: lazyRoute(<ChecksPage />) },
							{ path: projectRoutePath("insight"), element: lazyRoute(<InsightPage />) },
							{ path: projectRoutePath("alerts"), element: lazyRoute(<AlertsPage />) },
							{ path: `${projectRoutePath("alerts")}/incident/:incidentId`, element: lazyRoute(<AlertsPage />) },
							{ path: projectRoutePath("members"), element: lazyRoute(<MembersPage />) },
							{ path: projectRoutePath("project"), element: lazyRoute(<ProjectPage />) }
						]
					},
					{ path: "settings", element: lazyRoute(<SettingsPage />) }
				]
			},
			{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
		]
	}
]);

export function AppRouter() {
	return (
		<QueryClientProvider client={queryClient}>
			<SessionProvider>
				<CurrentProjectProvider>
					<ConfirmProvider>
						<RouterProvider router={router} />
						<ToastProvider />
					</ConfirmProvider>
				</CurrentProjectProvider>
			</SessionProvider>
		</QueryClientProvider>
	);
}
