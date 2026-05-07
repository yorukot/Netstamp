import { AlertsPage } from "@/features/alerts/components/AlertsPage";
import { AuthPage } from "@/features/auth/components/AuthPage";
import { OnboardingPage } from "@/features/auth/components/OnboardingPage";
import { useSession } from "@/features/auth/session/SessionContext";
import { SessionProvider } from "@/features/auth/session/SessionProvider";
import { ChecksPage } from "@/features/checks/components/ChecksPage";
import { DashboardPage } from "@/features/dashboard/components/DashboardPage";
import { InsightPage } from "@/features/insight/components/InsightPage";
import { NewProbeDrawer } from "@/features/probes/components/NewProbeDrawer";
import { ProbesPage } from "@/features/probes/components/ProbesPage";
import { SettingsPage } from "@/features/settings/components/SettingsPage";
import { TeamPage } from "@/features/team/components/TeamPage";
import { AppShell } from "@/layouts/AppShell";
import { createBrowserRouter, Navigate as RouterNavigate, RouterProvider, useLocation, useNavigate } from "react-router-dom";
import { pathForRoute } from "./routePaths";
import type { AppRoute, Navigate } from "./routeTypes";

function appRoutePath(route: AppRoute) {
	return pathForRoute(route).slice(1);
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

	return <AuthPage mode={mode} navigate={navigate} />;
}

function OnboardingRoute() {
	const navigate = useRouteNavigate();

	return <OnboardingPage navigate={navigate} />;
}

function DashboardRoute() {
	const navigate = useRouteNavigate();

	return <DashboardPage navigate={navigate} />;
}

function ProtectedAppShell() {
	const { session } = useSession();
	const location = useLocation();

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
				element: <ProbesPage />,
				children: [{ path: "new", element: <NewProbeDrawer /> }]
			},
			{ path: appRoutePath("insight"), element: <InsightPage /> },
			{ path: appRoutePath("checks"), element: <ChecksPage /> },
			{ path: appRoutePath("alerts"), element: <AlertsPage /> },
			{ path: appRoutePath("team"), element: <TeamPage /> },
			{ path: appRoutePath("settings"), element: <SettingsPage /> }
		]
	},
	{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
]);

export function AppRouter() {
	return (
		<SessionProvider>
			<RouterProvider router={router} />
		</SessionProvider>
	);
}
