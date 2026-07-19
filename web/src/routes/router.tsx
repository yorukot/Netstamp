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
import { DefaultProjectRedirect, LegacyCheckDetailRedirect, LegacyLabelDetailRedirect, LegacyProbeDetailRedirect, LegacyStatusPageEditorRedirect } from "./redirects";
import { RouteFrame } from "./RouteFrame";
import { pathForRoute, projectRoutePath } from "./routePaths";
import { lazyRoute } from "./routeSuspense";

export const router = createBrowserRouter([
	{
		element: <RouteFrame />,
		children: [
			{ path: pathForRoute("landing"), handle: pageTitleHandle("pageTitles.overview"), element: <RouterNavigate to={pathForRoute("dashboard")} replace /> },
			{ path: pathForRoute("login"), handle: pageTitleHandle("pageTitles.login"), element: <AuthRoute mode="login" /> },
			{ path: pathForRoute("register"), handle: pageTitleHandle("pageTitles.signUp"), element: <AuthRoute mode="register" /> },
			{ path: pathForRoute("forgotPassword"), handle: pageTitleHandle("pageTitles.forgotPassword"), element: <PasswordResetRoute mode="forgot" /> },
			{ path: pathForRoute("resetPassword"), handle: pageTitleHandle("pageTitles.setNewPassword"), element: <PasswordResetRoute mode="reset" /> },
			{ path: pathForRoute("verifyEmail"), handle: pageTitleHandle("pageTitles.verifyEmail"), element: <EmailVerificationRoute /> },
			{ path: "status/:slug", handle: pageTitleHandle("pageTitles.status"), element: lazyRoute(<PublicStatusPage />) },
			{ path: pathForRoute("onboarding"), handle: pageTitleHandle("pageTitles.onboarding"), element: <OnboardingRoute /> },
			{
				element: <ProtectedAppShell />,
				children: [
					{ path: "dashboard", handle: pageTitleHandle("pageTitles.overview"), element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "probes",
						handle: pageTitleHandle("pageTitles.probes"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="probes" /> },
							{ path: "new", handle: pageTitleHandle("pageTitles.newProbe"), element: <DefaultProjectRedirect route="newProbe" /> },
							{ path: ":probeId", handle: pageTitleHandle("pageTitles.probeDetail"), element: <LegacyProbeDetailRedirect /> }
						]
					},
					{
						path: "labels",
						handle: pageTitleHandle("pageTitles.labels"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="labels" /> },
							{ path: ":labelId", handle: pageTitleHandle("pageTitles.labelDetail"), element: <LegacyLabelDetailRedirect /> }
						]
					},
					{
						path: "checks",
						handle: pageTitleHandle("pageTitles.checks"),
						children: [
							{ index: true, element: <DefaultProjectRedirect route="checks" /> },
							{ path: ":checkId", handle: pageTitleHandle("pageTitles.checkDetail"), element: <LegacyCheckDetailRedirect /> }
						]
					},
					{ path: "insight", handle: pageTitleHandle("pageTitles.insight"), element: <DefaultProjectRedirect route="insight" /> },
					{ path: "alerts", handle: pageTitleHandle("pageTitles.alerts"), element: <DefaultProjectRedirect route="alerts" /> },
					{ path: "status-pages", handle: pageTitleHandle("pageTitles.statusPages"), element: <DefaultProjectRedirect route="statusPages" /> },
					{ path: "status-pages/:pageId/edit", handle: pageTitleHandle("pageTitles.statusPageEditor"), element: <LegacyStatusPageEditorRedirect /> },
					{ path: "members", handle: pageTitleHandle("pageTitles.members"), element: <DefaultProjectRedirect route="members" /> },
					{ path: "project", handle: pageTitleHandle("pageTitles.projectSettings"), element: <DefaultProjectRedirect route="projectSettings" /> },
					{ path: "projects", handle: pageTitleHandle("pageTitles.overview"), element: <DefaultProjectRedirect route="dashboard" /> },
					{
						path: "projects/:projectRef",
						element: <ProjectRouteBoundary />,
						children: [
							{ index: true, handle: pageTitleHandle("pageTitles.overview"), element: <RouterNavigate to={projectRoutePath("dashboard")} replace /> },
							{ path: projectRoutePath("dashboard"), handle: pageTitleHandle("pageTitles.overview"), element: <DashboardRoute /> },
							{
								path: projectRoutePath("probes"),
								handle: pageTitleHandle("pageTitles.probes"),
								element: lazyRoute(<ProbesPage />),
								children: [
									{ path: "new", handle: pageTitleHandle("pageTitles.newProbe"), element: lazyRoute(<NewProbeDrawer />) },
									{ path: ":probeId", handle: pageTitleHandle("pageTitles.probeDetail"), element: null }
								]
							},
							{ path: projectRoutePath("labels"), handle: pageTitleHandle("pageTitles.labels"), element: lazyRoute(<LabelsPage />) },
							{ path: `${projectRoutePath("labels")}/:labelId`, handle: pageTitleHandle("pageTitles.labelDetail"), element: lazyRoute(<LabelsPage />) },
							{ path: projectRoutePath("checks"), handle: pageTitleHandle("pageTitles.checks"), element: lazyRoute(<ChecksPage />) },
							{ path: `${projectRoutePath("checks")}/:checkId`, handle: pageTitleHandle("pageTitles.checkDetail"), element: lazyRoute(<ChecksPage />) },
							{ path: projectRoutePath("insight"), handle: pageTitleHandle("pageTitles.insight"), element: lazyRoute(<InsightPage />) },
							{ path: projectRoutePath("alerts"), handle: pageTitleHandle("pageTitles.alerts"), element: lazyRoute(<AlertsPage />) },
							{ path: `${projectRoutePath("alerts")}/incident/:incidentId`, handle: pageTitleHandle("pageTitles.incidentDetail"), element: lazyRoute(<AlertsPage />) },
							{ path: projectRoutePath("statusPages"), handle: pageTitleHandle("pageTitles.statusPages"), element: lazyRoute(<StatusPagesPage />) },
							{ path: `${projectRoutePath("statusPages")}/:pageId/edit`, handle: pageTitleHandle("pageTitles.statusPageEditor"), element: lazyRoute(<StatusPageBuilderPage />) },
							{ path: projectRoutePath("members"), handle: pageTitleHandle("pageTitles.members"), element: lazyRoute(<MembersPage />) },
							{ path: projectRoutePath("projectSettings"), handle: pageTitleHandle("pageTitles.projectSettings"), element: lazyRoute(<ProjectPage />) }
						]
					},
					{ path: "settings", handle: pageTitleHandle("pageTitles.account"), element: lazyRoute(<SettingsPage />) },
					{ path: "admin", handle: pageTitleHandle("pageTitles.systemSettings"), element: lazyRoute(<AdminPage />) }
				]
			},
			{ path: "*", element: <RouterNavigate to={pathForRoute("login")} replace /> }
		]
	}
]);
