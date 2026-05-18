import { useSession } from "@/features/auth/session/SessionContext";
import { SessionProvider } from "@/features/auth/session/SessionProvider";
import { AppShell } from "@/layouts/AppShell";
import { queryClient } from "@/shared/api/queryClient";
import { CurrentProjectProvider } from "@/shared/api/useCurrentProject";
import { ToastProvider } from "@/shared/components/ToastProvider";
import { QueryClientProvider } from "@tanstack/react-query";
import { lazy, Suspense, type ReactNode } from "react";
import { createBrowserRouter, Navigate as RouterNavigate, RouterProvider, useLocation, useNavigate } from "react-router-dom";
import { pathForRoute } from "./routePaths";
import type { AppRoute, Navigate } from "./routeTypes";

const AlertsPage = lazy(() => import("@/features/alerts/components/AlertsPage").then(module => ({ default: module.AlertsPage })));
const AuthPage = lazy(() => import("@/features/auth/components/AuthPage").then(module => ({ default: module.AuthPage })));
const OnboardingPage = lazy(() => import("@/features/auth/components/OnboardingPage").then(module => ({ default: module.OnboardingPage })));
const ChecksPage = lazy(() => import("@/features/checks/components/ChecksPage").then(module => ({ default: module.ChecksPage })));
const DashboardPage = lazy(() => import("@/features/dashboard/components/DashboardPage").then(module => ({ default: module.DashboardPage })));
const InsightPage = lazy(() => import("@/features/insight/components/InsightPage").then(module => ({ default: module.InsightPage })));
const NewProbeDrawer = lazy(() => import("@/features/probes/components/NewProbeDrawer").then(module => ({ default: module.NewProbeDrawer })));
const ProbesPage = lazy(() => import("@/features/probes/components/ProbesPage").then(module => ({ default: module.ProbesPage })));
const SettingsPage = lazy(() => import("@/features/settings/components/SettingsPage").then(module => ({ default: module.SettingsPage })));
const TeamPage = lazy(() => import("@/features/team/components/TeamPage").then(module => ({ default: module.TeamPage })));

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
			{ path: appRoutePath("insight"), element: lazyRoute(<InsightPage />) },
			{ path: appRoutePath("checks"), element: lazyRoute(<ChecksPage />) },
			{ path: appRoutePath("alerts"), element: lazyRoute(<AlertsPage />) },
			{ path: appRoutePath("team"), element: lazyRoute(<TeamPage />) },
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
					<RouterProvider router={router} />
					<ToastProvider />
				</CurrentProjectProvider>
			</SessionProvider>
		</QueryClientProvider>
	);
}
