import { createBrowserRouter, Navigate as RouterNavigate } from "react-router-dom";
import { AuthRoute, DashboardRoute, EmailVerificationRoute, OnboardingRoute, PasswordResetRoute, ProjectRouteBoundary, ProtectedAppShell } from "./guards";
import {
	AdminPage,
	AlertsPage,
	ChecksPage,
	InsightPage,
	LabelsPage,
	MembersPage,
	NewProbeDrawer,
	ProbesPage,
	ProjectPage,
	PublicStatusPage,
	SettingsPage,
	StatusPageBuilderPage,
	StatusPagesPage
} from "./lazyRoutes";
import { pageTitleHandle } from "./pageTitles";
import { DefaultProjectRedirect, LegacyCheckDetailRedirect, LegacyLabelDetailRedirect, LegacyProbeDetailRedirect } from "./redirects";
import { RouteFrame } from "./RouteFrame";
import { pathForRoute, projectRoutePath } from "./routePaths";
import { lazyRoute } from "./routeSuspense";

export const router = createBrowserRouter([
	{
		element: <RouteFrame />,
		children: [
			{ path: pathForRoute("landing"), handle: pageTitleHandle("Overview"), element: <RouterNavigate to={pathForRoute("dashboard")} replace /> },
			{ path: pathForRoute("login"), handle: pageTitleHandle("Login"), element: <AuthRoute mode="login" /> },
			{ path: pathForRoute("register"), handle: pageTitleHandle("Sign Up"), element: <AuthRoute mode="register" /> },
			{ path: pathForRoute("forgotPassword"), handle: pageTitleHandle("Forgot Password"), element: <PasswordResetRoute mode="forgot" /> },
			{ path: pathForRoute("resetPassword"), handle: pageTitleHandle("Set New Password"), element: <PasswordResetRoute mode="reset" /> },
			{ path: pathForRoute("verifyEmail"), handle: pageTitleHandle("Verify Email"), element: <EmailVerificationRoute /> },
			{ path: "status/:slug", handle: pageTitleHandle("Status"), element: lazyRoute(<PublicStatusPage />) },
			{ path: pathForRoute("onboarding"), handle: pageTitleHandle("Onboarding"), element: <OnboardingRoute /> },
			{
				element: <ProtectedAppShell />,
				children: [
					{ path: "dashboard", handle: pageTitleHandle("Overview"), element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "probes",
						handle: pageTitleHandle("Probes"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="probes" /> },
							{ path: "new", handle: pageTitleHandle("New Probe"), element: <DefaultProjectRedirect route="newProbe" /> },
							{ path: ":probeId", handle: pageTitleHandle("Probe Detail"), element: <LegacyProbeDetailRedirect /> }
						]
					},
					{
						path: "labels",
						handle: pageTitleHandle("Labels"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="labels" /> },
							{ path: ":labelId", handle: pageTitleHandle("Label Detail"), element: <LegacyLabelDetailRedirect /> }
						]
					},
					{
						path: "checks",
						handle: pageTitleHandle("Checks"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="checks" /> },
							{ path: ":checkId", handle: pageTitleHandle("Check Detail"), element: <LegacyCheckDetailRedirect /> }
						]
					},
					{ path: "insight", handle: pageTitleHandle("Insight"), element: <DefaultProjectRedirect route="insight" /> },
					{ path: "alerts", handle: pageTitleHandle("Alerts"), element: <DefaultProjectRedirect route="alerts" /> },
					{ path: "status-pages", handle: pageTitleHandle("Status Pages"), element: <DefaultProjectRedirect route="statusPages" /> },
					{ path: "status-pages/:pageId/edit", handle: pageTitleHandle("Status Page Editor"), element: <DefaultProjectRedirect route="statusPages" /> },
					{ path: "members", handle: pageTitleHandle("Members"), element: <DefaultProjectRedirect route="members" /> },
					{ path: "project", handle: pageTitleHandle("Project Settings"), element: <DefaultProjectRedirect route="projectSettings" /> },
					{ path: "projects", handle: pageTitleHandle("Overview"), element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "projects/:projectRef",
						element: <ProjectRouteBoundary />,
						children: [
							{ index: true, handle: pageTitleHandle("Overview"), element: <RouterNavigate to={projectRoutePath("dashboard")} replace /> },
							{ path: projectRoutePath("dashboard"), handle: pageTitleHandle("Overview"), element: <DashboardRoute /> },
							{
								path: projectRoutePath("probes"),
								handle: pageTitleHandle("Probes"),
								element: lazyRoute(<ProbesPage />),
								children: [
									{ path: "new", handle: pageTitleHandle("New Probe"), element: lazyRoute(<NewProbeDrawer />) },
									{ path: ":probeId", handle: pageTitleHandle("Probe Detail"), element: null }
								]
							},
							{ path: projectRoutePath("labels"), handle: pageTitleHandle("Labels"), element: lazyRoute(<LabelsPage />) },
							{ path: `${projectRoutePath("labels")}/:labelId`, handle: pageTitleHandle("Label Detail"), element: lazyRoute(<LabelsPage />) },
							{ path: projectRoutePath("checks"), handle: pageTitleHandle("Checks"), element: lazyRoute(<ChecksPage />) },
							{ path: `${projectRoutePath("checks")}/:checkId`, handle: pageTitleHandle("Check Detail"), element: lazyRoute(<ChecksPage />) },
							{ path: projectRoutePath("insight"), handle: pageTitleHandle("Insight"), element: lazyRoute(<InsightPage />) },
							{ path: projectRoutePath("alerts"), handle: pageTitleHandle("Alerts"), element: lazyRoute(<AlertsPage />) },
							{ path: `${projectRoutePath("alerts")}/incident/:incidentId`, handle: pageTitleHandle("Incident Detail"), element: lazyRoute(<AlertsPage />) },
							{ path: projectRoutePath("statusPages"), handle: pageTitleHandle("Status Pages"), element: lazyRoute(<StatusPagesPage />) },
							{ path: `${projectRoutePath("statusPages")}/:pageId/edit`, handle: pageTitleHandle("Status Page Editor"), element: lazyRoute(<StatusPageBuilderPage />) },
							{ path: projectRoutePath("members"), handle: pageTitleHandle("Members"), element: lazyRoute(<MembersPage />) },
							{ path: projectRoutePath("projectSettings"), handle: pageTitleHandle("Project Settings"), element: lazyRoute(<ProjectPage />) }
						]
					},
					{ path: "settings", handle: pageTitleHandle("Account"), element: lazyRoute(<SettingsPage />) },
					{ path: "admin", handle: pageTitleHandle("System Settings"), element: lazyRoute(<AdminPage />) }
				]
			},
			{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
		]
	}
]);
