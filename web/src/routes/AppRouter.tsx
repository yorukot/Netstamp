import { AlertsPage } from "@/features/alerts/components/AlertsPage";
import { AuthPage } from "@/features/auth/components/AuthPage";
import { OnboardingPage } from "@/features/auth/components/OnboardingPage";
import { ChecksPage } from "@/features/checks/components/ChecksPage";
import { DashboardPage } from "@/features/dashboard/components/DashboardPage";
import { InsightPage } from "@/features/insight/components/InsightPage";
import { LandingPage } from "@/features/landing/components/LandingPage";
import { ProbesPage } from "@/features/probes/components/ProbesPage";
import { SettingsPage } from "@/features/settings/components/SettingsPage";
import { TeamPage } from "@/features/team/components/TeamPage";
import { AppShell } from "@/layouts/AppShell";
import type { AppRoute, Navigate } from "@/shared/utils/mockData";
import { createBrowserRouter, Navigate as RouterNavigate, RouterProvider, useNavigate } from "react-router-dom";
import { pathForRoute } from "./routePaths";

function appRoutePath(route: AppRoute) {
	return pathForRoute(route).slice(1);
}

function useRouteNavigate(): Navigate {
	const navigate = useNavigate();

	return (route, hash) => navigate(`${pathForRoute(route)}${hash ?? ""}`);
}

function LandingRoute() {
	const navigate = useRouteNavigate();

	return <LandingPage navigate={navigate} />;
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

const router = createBrowserRouter([
	{ path: pathForRoute("landing"), element: <LandingRoute /> },
	{ path: pathForRoute("login"), element: <AuthRoute mode="login" /> },
	{ path: pathForRoute("register"), element: <AuthRoute mode="register" /> },
	{ path: pathForRoute("onboarding"), element: <OnboardingRoute /> },
	{
		element: <AppShell />,
		children: [
			{ path: appRoutePath("dashboard"), element: <DashboardRoute /> },
			{ path: appRoutePath("probes"), element: <ProbesPage /> },
			{ path: appRoutePath("insight"), element: <InsightPage /> },
			{ path: appRoutePath("checks"), element: <ChecksPage /> },
			{ path: appRoutePath("alerts"), element: <AlertsPage /> },
			{ path: appRoutePath("team"), element: <TeamPage /> },
			{ path: appRoutePath("settings"), element: <SettingsPage /> }
		]
	},
	{ path: "*", element: <RouterNavigate to={pathForRoute("landing")} replace /> }
]);

export function AppRouter() {
	return <RouterProvider router={router} />;
}
