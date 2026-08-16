import adminEn from "./locales/en/admin.json";
import alertsEn from "./locales/en/alerts.json";
import authEn from "./locales/en/auth.json";
import checksEn from "./locales/en/checks.json";
import commonEn from "./locales/en/common.json";
import dashboardEn from "./locales/en/dashboard.json";
import errorsEn from "./locales/en/errors.json";
import insightEn from "./locales/en/insight.json";
import labelsEn from "./locales/en/labels.json";
import navigationEn from "./locales/en/navigation.json";
import probesEn from "./locales/en/probes.json";
import projectEn from "./locales/en/project.json";
import settingsEn from "./locales/en/settings.json";
import statusEn from "./locales/en/status.json";

export const defaultNamespace = "common";
export const namespaces = ["common", "navigation", "auth", "dashboard", "project", "probes", "checks", "labels", "insight", "alerts", "status", "settings", "admin", "errors"] as const;

const en = {
	admin: adminEn,
	alerts: alertsEn,
	auth: authEn,
	checks: checksEn,
	common: commonEn,
	dashboard: dashboardEn,
	errors: errorsEn,
	insight: insightEn,
	labels: labelsEn,
	navigation: navigationEn,
	probes: probesEn,
	project: projectEn,
	settings: settingsEn,
	status: statusEn
};

export const resources = {
	en
} as const;

export type AppNamespace = (typeof namespaces)[number];
export type AppResources = typeof en;
