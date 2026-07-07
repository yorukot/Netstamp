import { createBrowserRouter, Navigate as RouterNavigate } from "react-router-dom";
import { AuthRoute, DashboardRoute, EmailVerificationRoute, OnboardingRoute, PasswordResetRoute, ProjectRouteBoundary, ProtectedAppShell } from "./guards";
import { AdminPage, AlertsPage, ChecksPage, InsightPage, LabelsPage, MembersPage, NewProbeDrawer, ProbesPage, ProjectPage, PublicStatusPage, SettingsPage, StatusPagesPage } from "./lazyRoutes";
import { DefaultProjectRedirect, LegacyCheckDetailRedirect, LegacyLabelDetailRedirect, LegacyProbeDetailRedirect } from "./redirects";
import { RouteFrame } from "./RouteFrame";
import { pathForRoute, projectRoutePath } from "./routePaths";
import { lazyRoute } from "./routeSuspense";

export const router = createBrowserRouter([
	{
		element: <RouteFrame />,
		children: [
			{ path: pathForRoute("landing"), element: <RouterNavigate to={pathForRoute("dashboard")} replace /> },
			{ path: pathForRoute("login"), element: <AuthRoute mode="login" /> },
			{ path: pathForRoute("register"), element: <AuthRoute mode="register" /> },
			{ path: pathForRoute("forgotPassword"), element: <PasswordResetRoute mode="forgot" /> },
			{ path: pathForRoute("resetPassword"), element: <PasswordResetRoute mode="reset" /> },
			{ path: pathForRoute("verifyEmail"), element: <EmailVerificationRoute /> },
			{ path: "status/:slug", element: lazyRoute(<PublicStatusPage />) },
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
					{ path: "status-pages", element: <DefaultProjectRedirect route="statusPages" /> },
					{ path: "members", element: <DefaultProjectRedirect route="members" /> },
					{ path: "project", element: <DefaultProjectRedirect route="projectSettings" /> },
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
							{ path: projectRoutePath("statusPages"), element: lazyRoute(<StatusPagesPage />) },
							{ path: projectRoutePath("members"), element: lazyRoute(<MembersPage />) },
							{ path: projectRoutePath("projectSettings"), element: lazyRoute(<ProjectPage />) }
						]
					},
					{ path: "settings", element: lazyRoute(<SettingsPage />) },
					{ path: "admin", element: lazyRoute(<AdminPage />) }
				]
			},
			{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
		]
	}
]);
