import { lazy } from "react";

export const AuthPage = lazy(() => import("@/features/auth/components/AuthPage").then(module => ({ default: module.AuthPage })));
export const AlertsPage = lazy(() => import("@/features/alerts/components/AlertsPage").then(module => ({ default: module.AlertsPage })));
export const OnboardingPage = lazy(() => import("@/features/auth/components/OnboardingPage").then(module => ({ default: module.OnboardingPage })));
export const ChecksPage = lazy(() => import("@/features/checks/components/ChecksPage").then(module => ({ default: module.ChecksPage })));
export const DashboardPage = lazy(() => import("@/features/dashboard/components/DashboardPage").then(module => ({ default: module.DashboardPage })));
export const InsightPage = lazy(() => import("@/features/insight/components/InsightPage").then(module => ({ default: module.InsightPage })));
export const LabelsPage = lazy(() => import("@/features/labels/components/LabelsPage").then(module => ({ default: module.LabelsPage })));
export const MembersPage = lazy(() => import("@/features/project/components/MembersPage").then(module => ({ default: module.MembersPage })));
export const NewProbeDrawer = lazy(() => import("@/features/probes/components/NewProbeDrawer").then(module => ({ default: module.NewProbeDrawer })));
export const ProbesPage = lazy(() => import("@/features/probes/components/ProbesPage").then(module => ({ default: module.ProbesPage })));
export const SettingsPage = lazy(() => import("@/features/settings/components/SettingsPage").then(module => ({ default: module.SettingsPage })));
export const ProjectPage = lazy(() => import("@/features/project/components/ProjectPage").then(module => ({ default: module.ProjectPage })));
export const PublicStatusPage = lazy(() => import("@/features/status-pages/components/PublicStatusPage").then(module => ({ default: module.PublicStatusPage })));
export const StatusPagesPage = lazy(() => import("@/features/status-pages/components/StatusPagesPage").then(module => ({ default: module.StatusPagesPage })));
