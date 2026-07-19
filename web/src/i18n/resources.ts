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
import adminZhTW from "./locales/zh-TW/admin.json";
import alertsZhTW from "./locales/zh-TW/alerts.json";
import authZhTW from "./locales/zh-TW/auth.json";
import checksZhTW from "./locales/zh-TW/checks.json";
import commonZhTW from "./locales/zh-TW/common.json";
import dashboardZhTW from "./locales/zh-TW/dashboard.json";
import errorsZhTW from "./locales/zh-TW/errors.json";
import insightZhTW from "./locales/zh-TW/insight.json";
import labelsZhTW from "./locales/zh-TW/labels.json";
import navigationZhTW from "./locales/zh-TW/navigation.json";
import probesZhTW from "./locales/zh-TW/probes.json";
import projectZhTW from "./locales/zh-TW/project.json";
import settingsZhTW from "./locales/zh-TW/settings.json";
import statusZhTW from "./locales/zh-TW/status.json";

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

const zhTW: typeof en = {
	admin: adminZhTW,
	alerts: alertsZhTW,
	auth: authZhTW,
	checks: checksZhTW,
	common: commonZhTW,
	dashboard: dashboardZhTW,
	errors: errorsZhTW,
	insight: insightZhTW,
	labels: labelsZhTW,
	navigation: navigationZhTW,
	probes: probesZhTW,
	project: projectZhTW,
	settings: settingsZhTW,
	status: statusZhTW
};

export const resources = {
	en,
	"zh-TW": zhTW
} as const;

export type AppNamespace = (typeof namespaces)[number];
export type AppResources = typeof en;
