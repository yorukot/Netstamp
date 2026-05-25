import { useSession } from "@/features/auth/session/SessionContext";
import { SessionProvider } from "@/features/auth/session/SessionProvider";
import { AppShell } from "@/layouts/AppShell";
import { queryClient } from "@/shared/api/queryClient";
import { CurrentProjectProvider } from "@/shared/api/useCurrentProject";
import { ConfirmProvider } from "@/shared/components/ConfirmProvider";
import { ToastProvider } from "@/shared/components/ToastProvider";
import { QueryClientProvider } from "@tanstack/react-query";
import { lazy, Suspense, type ReactNode } from "react";
import { createBrowserRouter, Navigate as RouterNavigate, RouterProvider, useLocation, useNavigate } from "react-router-dom";
import { pathForRoute } from "./routePaths";
import type { AppRoute, Navigate } from "./routeTypes";

const AuthPage = lazy(() => import("@/features/auth/components/AuthPage").then(module => ({ default: module.AuthPage })));
const OnboardingPage = lazy(() => import("@/features/auth/components/OnboardingPage").then(module => ({ default: module.OnboardingPage })));
const ChecksPage = lazy(() => import("@/features/checks/components/ChecksPage").then(module => ({ default: module.ChecksPage })));
const DashboardPage = lazy(() => import("@/features/dashboard/components/DashboardPage").then(module => ({ default: module.DashboardPage })));
const InsightPage = lazy(() => import("@/features/insight/components/InsightPage").then(module => ({ default: module.InsightPage })));
const LabelsPage = lazy(() => import("@/features/labels/components/LabelsPage").then(module => ({ default: module.LabelsPage })));
const NewProbeDrawer = lazy(() => import("@/features/probes/components/NewProbeDrawer").then(module => ({ default: module.NewProbeDrawer })));
const ProbesPage = lazy(() => import("@/features/probes/components/ProbesPage").then(module => ({ default: module.ProbesPage })));
const SettingsPage = lazy(() => import("@/features/settings/components/SettingsPage").then(module => ({ default: module.SettingsPage })));
const ProjectPage = lazy(() => import("@/features/project/components/ProjectPage").then(module => ({ default: module.ProjectPage })));

function appRoutePath(route: AppRoute) {
	return pathForRoute(route).slice(1);
}

function lazyRoute(element: ReactNode) {
	return <Suspense fallback={null}>{element}</Suspense>;
}

function useRouteNavigate(): Navigate {
	const navigate = useNavigate();

	return route => navigate(pathForRoute(route));
}

interface AuthRouteProps {
	mode: "login" | "register";
}

function AuthRoute({ mode }: AuthRouteProps) {
	const navigate = useRouteNavigate();

	return lazyRoute(<AuthPage mode={mode} navigate={navigate} />);
}

function OnboardingRoute() {
	const navigate = useRouteNavigate();

	return lazyRoute(<OnboardingPage navigate={navigate} />);
}

function DashboardRoute() {
	const navigate = useRouteNavigate();

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

	return <AppShell />;
}

const router = createBrowserRouter([
	{ path: pathForRoute("landing"), element: <RouterNavigate to={pathForRoute("login")} replace /> },
	{ path: pathForRoute("login"), element: <AuthRoute mode="login" /> },
	{ path: pathForRoute("register"), element: <AuthRoute mode="register" /> },
	{ path: pathForRoute("onboarding"), element: <OnboardingRoute /> },
	{
		element: <ProtectedAppShell />,
		children: [
			{ path: appRoutePath("dashboard"), element: <DashboardRoute /> },
			{
				path: appRoutePath("probes"),
				element: lazyRoute(<ProbesPage />),
				children: [{ path: "new", element: lazyRoute(<NewProbeDrawer />) }]
			},
			{ path: appRoutePath("labels"), element: lazyRoute(<LabelsPage />) },
			{ path: appRoutePath("checks"), element: lazyRoute(<ChecksPage />) },
			{ path: appRoutePath("insight"), element: lazyRoute(<InsightPage />) },
			{ path: appRoutePath("project"), element: lazyRoute(<ProjectPage />) },
			{ path: appRoutePath("settings"), element: lazyRoute(<SettingsPage />) }
		]
	},
	{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
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
