import { i18n } from "@/i18n";
import { formatDateTime as formatLocaleDateTime } from "@/i18n/format";
import type { ApiAlertIncident, ApiAlertRule, ApiNotification, CreateAlertRuleInput, CreateNotificationInput, UpdateAlertRuleInput, UpdateNotificationInput } from "@/shared/api/types";
import type { BadgeTone } from "@netstamp/ui";
import type { MouseEvent } from "react";

const alertT = i18n.getFixedT(null, "alerts") as (key: string, options?: Record<string, unknown>) => string;

export type AlertTab = "incidents" | "rules" | "notifications";
export type IncidentStatusFilter = "open" | "acknowledged" | "resolved" | "all";
export type RuleStatusFilter = "all" | "enabled" | "disabled";
export type CheckType = CreateAlertRuleInput["scope"]["checkType"];
export type RuleCheckTypeFilter = "all" | CheckType;
export type NotificationStatusFilter = "all" | "enabled" | "disabled";
export type NotificationType = CreateNotificationInput["type"];
export type NotificationTypeFilter = "all" | NotificationType;
export type AlertMetric = CreateAlertRuleInput["condition"]["metric"];
export type AlertOperator = CreateAlertRuleInput["condition"]["operator"];
export type AlertSeverity = CreateAlertRuleInput["severity"];

export interface RuleEditorState {
	mode: "create" | "edit";
	rule?: ApiAlertRule;
}

export interface NotificationEditorState {
	mode: "create" | "edit";
	notification?: ApiNotification;
}

export type NotificationEditorStep = "type" | "detail";

export interface RuleFormState {
	name: string;
	description: string;
	enabled: string;
	severity: AlertSeverity;
	checkType: CheckType;
	probeId: string;
	checkId: string;
	metric: AlertMetric;
	operator: AlertOperator;
	threshold: string;
	windowSeconds: string;
	minSamples: string;
	triggerAfterMinutes: string;
	cooldownSeconds: string;
	selectedNotificationIds: string[];
}

export interface NotificationFormState {
	name: string;
	type: NotificationType;
	url: string;
	botToken: string;
	chatId: string;
	emailTo: string;
	enabled: string;
}

export interface NumericFieldValidation {
	value: number;
	error: string;
}

export interface RuleNumberValidation {
	threshold: NumericFieldValidation;
	windowSeconds: NumericFieldValidation;
	minSamples: NumericFieldValidation;
	triggerAfterMinutes: NumericFieldValidation;
	cooldownSeconds: NumericFieldValidation;
}

export const emptyRules: ApiAlertRule[] = [];
export const emptyIncidents: ApiAlertIncident[] = [];
export const emptyNotifications: ApiNotification[] = [];

export const alertTabs: Array<{ value: AlertTab }> = [{ value: "incidents" }, { value: "rules" }, { value: "notifications" }];

export const incidentStatusOptions: Array<{ value: IncidentStatusFilter }> = [{ value: "open" }, { value: "acknowledged" }, { value: "resolved" }, { value: "all" }];

export const ruleStatusOptions: Array<{ value: RuleStatusFilter }> = [{ value: "all" }, { value: "enabled" }, { value: "disabled" }];

export const ruleCheckTypeOptions: Array<{ value: RuleCheckTypeFilter }> = [{ value: "all" }, { value: "ping" }, { value: "tcp" }, { value: "traceroute" }, { value: "http" }];

export const notificationStatusOptions: Array<{ value: NotificationStatusFilter }> = [{ value: "all" }, { value: "enabled" }, { value: "disabled" }];

export const notificationFilterTypeOptions: Array<{ value: NotificationTypeFilter }> = [
	{ value: "all" },
	{ value: "webhook" },
	{ value: "slack" },
	{ value: "discord" },
	{ value: "telegram" },
	{ value: "email" }
];

export const checkTypeOptions: Array<{ value: CheckType; disabled?: boolean }> = [{ value: "ping" }, { value: "tcp" }, { value: "traceroute", disabled: true }, { value: "http" }];

export const metricOptions: Record<CheckType, Array<{ value: AlertMetric; unit?: string }>> = {
	ping: [
		{ value: "ping.loss_percent", unit: "%" },
		{ value: "ping.average_rtt_ms", unit: "ms" },
		{ value: "ping.max_rtt_ms", unit: "ms" },
		{ value: "ping.success_rate", unit: "%" }
	],
	tcp: [
		{ value: "tcp.failure_percent", unit: "%" },
		{ value: "tcp.average_connect_ms", unit: "ms" },
		{ value: "tcp.max_connect_ms", unit: "ms" },
		{ value: "tcp.success_rate", unit: "%" }
	],
	http: [
		{ value: "http.failure_percent", unit: "%" },
		{ value: "http.average_total_ms", unit: "ms" },
		{ value: "http.max_total_ms", unit: "ms" },
		{ value: "http.average_ttfb_ms", unit: "ms" },
		{ value: "http.max_ttfb_ms", unit: "ms" },
		{ value: "http.success_rate", unit: "%" },
		{ value: "http.certificate_days_remaining", unit: "days" }
	],
	traceroute: []
};

export const severityOptions: Array<{ value: AlertSeverity }> = [{ value: "info" }, { value: "warning" }, { value: "critical" }];

export const operatorOptions: Array<{ value: AlertOperator }> = [{ value: "gt" }, { value: "gte" }, { value: "lt" }, { value: "lte" }, { value: "eq" }];

export const enabledOptions = [{ value: "true" }, { value: "false" }];

export const notificationTypeOptions: Array<{ value: NotificationType }> = [{ value: "webhook" }, { value: "slack" }, { value: "discord" }, { value: "telegram" }, { value: "email" }];

export const operatorSymbols: Record<AlertOperator, string> = {
	gt: ">",
	gte: ">=",
	lt: "<",
	lte: "<=",
	eq: "="
};

export function formatDateTime(value?: string | null) {
	if (!value) {
		return "-";
	}
	const date = new Date(value);

	if (Number.isNaN(date.getTime())) {
		return "-";
	}

	return formatLocaleDateTime(date);
}

export function shortID(value: string) {
	return value.slice(0, 8);
}

export function formatDuration(seconds: number) {
	if (!Number.isFinite(seconds) || seconds <= 0) {
		return "-";
	}

	if (seconds % 86400 === 0) {
		const days = seconds / 86400;
		return alertT("duration.days", { count: days });
	}

	if (seconds % 3600 === 0) {
		const hours = seconds / 3600;
		return alertT("duration.hours", { count: hours });
	}

	if (seconds % 60 === 0) {
		const minutes = seconds / 60;
		return alertT("duration.minutes", { count: minutes });
	}

	return alertT("duration.seconds", { count: seconds });
}

export function severityTone(severity: string): BadgeTone {
	switch (severity) {
		case "critical":
			return "critical";
		case "warning":
			return "warning";
		case "info":
			return "accent";
		default:
			return "neutral";
	}
}

export function incidentTone(status: string): BadgeTone {
	switch (status) {
		case "open":
			return "critical";
		case "acknowledged":
			return "warning";
		case "resolved":
			return "success";
		default:
			return "neutral";
	}
}

export function metricOption(metric: AlertMetric) {
	return Object.values(metricOptions)
		.flat()
		.find(option => option.value === metric);
}

export function metricLabel(metric: AlertMetric) {
	const keys: Partial<Record<AlertMetric, string>> = {
		"ping.loss_percent": "pingLoss",
		"ping.average_rtt_ms": "pingAverageRtt",
		"ping.max_rtt_ms": "pingMaxRtt",
		"ping.success_rate": "pingSuccess",
		"tcp.failure_percent": "tcpFailure",
		"tcp.average_connect_ms": "tcpAverageConnect",
		"tcp.max_connect_ms": "tcpMaxConnect",
		"tcp.success_rate": "tcpSuccess",
		"http.failure_percent": "httpFailure",
		"http.average_total_ms": "httpAverageTotal",
		"http.max_total_ms": "httpMaxTotal",
		"http.average_ttfb_ms": "httpAverageTtfb",
		"http.max_ttfb_ms": "httpMaxTtfb",
		"http.success_rate": "httpSuccess",
		"http.certificate_days_remaining": "certificateDays"
	};
	const key = keys[metric];
	return key ? alertT(`rules.metrics.${key}`) : metric.replace(/_/g, " ");
}

export function metricUnit(metric: AlertMetric) {
	return metricOption(metric)?.unit ?? "";
}

export function formatThreshold(metric: AlertMetric, threshold: number | string) {
	const value = typeof threshold === "number" ? String(threshold) : threshold;
	const unit = metricUnit(metric);
	return unit ? `${value}${unit}` : value;
}

export function formatAlertCondition(condition: ApiAlertRule["condition"]) {
	return alertT("rules.conditionFor", {
		condition: `${metricLabel(condition.metric)} ${operatorSymbols[condition.operator]} ${formatThreshold(condition.metric, condition.threshold)}`,
		duration: formatDuration(condition.windowSeconds)
	});
}

export function formatIncidentReason(incident: ApiAlertIncident) {
	const summary = incident.lastSummary;
	if (summary.operator && typeof summary.threshold === "number") {
		return `${metricLabel(summary.metric)} ${operatorSymbols[summary.operator]} ${formatThreshold(summary.metric, summary.threshold)}`;
	}

	return metricLabel(summary.metric);
}

export function formatRuleScope(rule: ApiAlertRule) {
	const parts = [rule.scope.checkType.toUpperCase()];
	if (rule.scope.probeId) {
		parts.push(`${alertT("incidents.probe")} ${shortID(rule.scope.probeId)}`);
	}
	if (rule.scope.checkId) {
		parts.push(`${alertT("incidents.check")} ${shortID(rule.scope.checkId)}`);
	}
	return parts.join(" / ");
}

export function incidentProbeName(incident: ApiAlertIncident) {
	return incident.probe?.name || shortID(incident.probeId);
}

export function incidentCheckName(incident: ApiAlertIncident) {
	return incident.check?.name || shortID(incident.checkId);
}

export function incidentCheckTarget(incident: ApiAlertIncident) {
	return incident.check?.target || shortID(incident.checkId);
}

export function incidentTargetTitle(incident: ApiAlertIncident) {
	return alertT("incidents.targetTitle", { probeId: incident.probeId, checkId: incident.checkId });
}

export function formatIncidentProbe(incident: ApiAlertIncident) {
	return `${incidentProbeName(incident)} (${shortID(incident.probeId)})`;
}

export function formatIncidentCheck(incident: ApiAlertIncident) {
	return `${incidentCheckName(incident)} (${incident.checkType.toUpperCase()} / ${shortID(incident.checkId)})`;
}

export function notificationConfigString(config: ApiNotification["config"], key: string) {
	if (config && typeof config === "object" && key in config) {
		const value = (config as Record<string, unknown>)[key];
		return typeof value === "string" ? value : "";
	}

	return "";
}

export function notificationConfigStringArray(config: ApiNotification["config"], key: string) {
	if (config && typeof config === "object" && key in config) {
		const value = (config as Record<string, unknown>)[key];
		return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
	}

	return [];
}

export function notificationURL(notification: ApiNotification) {
	return notificationConfigString(notification.config, "url");
}

export function notificationChatID(notification: ApiNotification) {
	return notificationConfigString(notification.config, "chatId");
}

export function notificationEmails(notification: ApiNotification) {
	return notificationConfigStringArray(notification.config, "to");
}

export function notificationTypeLabel(notificationType: string) {
	switch (notificationType) {
		case "webhook":
			return alertT("notifications.types.webhook");
		case "slack":
			return alertT("notifications.types.slack");
		case "discord":
			return alertT("notifications.types.discord");
		case "telegram":
			return alertT("notifications.types.telegram");
		case "email":
			return alertT("notifications.types.email");
		default:
			return notificationType;
	}
}

export function notificationTypeOption(notificationType: NotificationType) {
	return {
		value: notificationType,
		label: notificationTypeLabel(notificationType),
		detail: alertT(`notifications.details.${notificationType}`)
	};
}

export function notificationWebhookURLLabel(notificationType: NotificationType) {
	switch (notificationType) {
		case "slack":
			return alertT("notifications.slackUrl");
		case "discord":
			return alertT("notifications.discordUrl");
		default:
			return alertT("notifications.webhookUrl");
	}
}

export function notificationWebhookURLPlaceholder(notificationType: NotificationType) {
	switch (notificationType) {
		case "slack":
			return "https://hooks.slack.com/services/...";
		case "discord":
			return "https://discord.com/api/webhooks/...";
		default:
			return "https://hooks.example.com/netstamp";
	}
}

export function notificationDestination(notification: ApiNotification) {
	switch (notification.type) {
		case "slack":
		case "discord":
		case "webhook":
			return notificationURL(notification);
		case "telegram":
			return notificationChatID(notification) ? alertT("notifications.chatDestination", { id: notificationChatID(notification) }) : "-";
		case "email":
			return notificationEmails(notification).join(", ") || "-";
		default:
			return "-";
	}
}

export function notificationNameByID(notifications: ApiNotification[], notificationID: string) {
	return notifications.find(notification => notification.id === notificationID)?.name ?? shortID(notificationID);
}

export function notificationLabel(rule: ApiAlertRule, notifications: ApiNotification[]) {
	if (!rule.notificationIds.length) {
		return alertT("notifications.none");
	}

	return rule.notificationIds.map(notificationID => notificationNameByID(notifications, notificationID)).join(", ");
}

export function rulesUsingNotification(rules: ApiAlertRule[], notificationID: string) {
	return rules.filter(rule => rule.notificationIds.includes(notificationID));
}

export function metricForCheckType(nextCheckType: CheckType) {
	return metricOptions[nextCheckType][0]?.value;
}

export function supportsAlertMetrics(checkType: CheckType) {
	return metricOptions[checkType].length > 0;
}

export function metricOptionsForForm(form: RuleFormState) {
	const options = metricOptions[form.checkType];

	if (options.length) {
		return options;
	}

	return [{ value: form.metric, unit: metricUnit(form.metric) }];
}

export function defaultRuleForm(): RuleFormState {
	return {
		name: "",
		description: "",
		enabled: "true",
		severity: "critical",
		checkType: "ping",
		probeId: "",
		checkId: "",
		metric: "ping.loss_percent",
		operator: "gte",
		threshold: "10",
		windowSeconds: "300",
		minSamples: "3",
		triggerAfterMinutes: "1",
		cooldownSeconds: "900",
		selectedNotificationIds: []
	};
}

export function ruleFormFromRule(rule: ApiAlertRule): RuleFormState {
	const checkType = rule.scope.checkType;
	const hasCompatibleMetric = metricOptions[checkType].some(option => option.value === rule.condition.metric);
	return {
		name: rule.name,
		description: rule.description ?? "",
		enabled: String(rule.enabled),
		severity: rule.severity,
		checkType,
		probeId: rule.scope.probeId ?? "",
		checkId: rule.scope.checkId ?? "",
		metric: hasCompatibleMetric ? rule.condition.metric : (metricForCheckType(checkType) ?? rule.condition.metric),
		operator: rule.condition.operator,
		threshold: String(rule.condition.threshold),
		windowSeconds: String(rule.condition.windowSeconds),
		minSamples: String(rule.condition.minSamples),
		triggerAfterMinutes: String(rule.triggerAfterSeconds / 60),
		cooldownSeconds: String(rule.cooldownSeconds),
		selectedNotificationIds: rule.notificationIds
	};
}

export function defaultNotificationForm(): NotificationFormState {
	return {
		name: "",
		type: "webhook",
		url: "",
		botToken: "",
		chatId: "",
		emailTo: "",
		enabled: "true"
	};
}

export function notificationFormFromNotification(notification: ApiNotification): NotificationFormState {
	return {
		name: notification.name,
		type: notification.type,
		url: notificationURL(notification),
		botToken: "",
		chatId: notificationChatID(notification),
		emailTo: notificationEmails(notification).join(", "),
		enabled: String(notification.enabled)
	};
}

export function validateNumericField(label: string, value: string, options: { integer?: boolean; min?: number; max?: number } = {}): NumericFieldValidation {
	const trimmed = value.trim();

	if (!trimmed) {
		return { value: Number.NaN, error: alertT("rules.validation.required", { label }) };
	}

	const parsed = Number(trimmed);

	if (!Number.isFinite(parsed)) {
		return { value: Number.NaN, error: alertT("rules.validation.number", { label }) };
	}

	if (options.integer && !Number.isInteger(parsed)) {
		return { value: parsed, error: alertT("rules.validation.whole", { label }) };
	}

	if (typeof options.min === "number" && parsed < options.min) {
		return { value: parsed, error: alertT("rules.validation.minimum", { label, min: options.min }) };
	}

	if (typeof options.max === "number" && parsed > options.max) {
		return { value: parsed, error: alertT("rules.validation.maximum", { label, max: options.max }) };
	}

	return { value: parsed, error: "" };
}

export function validateRuleNumbers(form: RuleFormState): RuleNumberValidation {
	return {
		threshold: validateNumericField(alertT("rules.threshold"), form.threshold, { min: 0 }),
		windowSeconds: validateNumericField(alertT("rules.windowSeconds"), form.windowSeconds, { integer: true, min: 60, max: 86400 }),
		minSamples: validateNumericField(alertT("rules.minSamples"), form.minSamples, { integer: true, min: 1, max: 10000 }),
		triggerAfterMinutes: validateNumericField(alertT("rules.triggerMinutes"), form.triggerAfterMinutes, { integer: true, min: 1, max: 1440 }),
		cooldownSeconds: validateNumericField(alertT("rules.cooldownSeconds"), form.cooldownSeconds, { integer: true, min: 60, max: 86400 })
	};
}

export function ruleNumberError(validation: RuleNumberValidation) {
	return validation.threshold.error || validation.windowSeconds.error || validation.minSamples.error || validation.triggerAfterMinutes.error || validation.cooldownSeconds.error;
}

export function rulePayload(form: RuleFormState): CreateAlertRuleInput | UpdateAlertRuleInput {
	const description = form.description.trim();
	const probeId = form.probeId.trim();
	const checkId = form.checkId.trim();
	const numbers = validateRuleNumbers(form);
	const numberError = ruleNumberError(numbers);

	if (numberError) {
		throw new Error(numberError);
	}

	return {
		name: form.name.trim(),
		description: description || undefined,
		enabled: form.enabled === "true",
		severity: form.severity,
		scope: {
			checkType: form.checkType,
			...(probeId ? { probeId } : {}),
			...(checkId ? { checkId } : {})
		},
		condition: {
			type: "metric_threshold",
			metric: form.metric,
			operator: form.operator,
			threshold: numbers.threshold.value,
			windowSeconds: numbers.windowSeconds.value,
			minSamples: numbers.minSamples.value
		},
		triggerAfterSeconds: numbers.triggerAfterMinutes.value * 60,
		cooldownSeconds: numbers.cooldownSeconds.value,
		notificationIds: form.selectedNotificationIds
	};
}

export function notificationPayload(form: NotificationFormState): CreateNotificationInput | UpdateNotificationInput {
	const config =
		form.type === "telegram" ? { botToken: form.botToken.trim(), chatId: form.chatId.trim() } : form.type === "email" ? { to: notificationEmailRecipients(form.emailTo) } : { url: form.url.trim() };

	return {
		name: form.name.trim(),
		type: form.type,
		enabled: form.enabled === "true",
		config
	};
}

export function notificationFormReady(form: NotificationFormState) {
	if (!form.name.trim()) {
		return false;
	}
	if (form.type === "telegram") {
		return Boolean(form.botToken.trim() && form.chatId.trim());
	}
	if (form.type === "email") {
		return notificationEmailRecipients(form.emailTo).length > 0;
	}
	return Boolean(form.url.trim());
}

export function notificationEmailRecipients(value: string) {
	return value
		.split(/[\n,;]+/)
		.map(item => item.trim())
		.filter(Boolean);
}

export function rulePreview(form: RuleFormState, notifications: ApiNotification[], numbers: RuleNumberValidation) {
	const metric = metricLabel(form.metric);
	const threshold = formatThreshold(form.metric, form.threshold || "0");
	const notification = form.selectedNotificationIds.length
		? form.selectedNotificationIds.map(notificationID => notificationNameByID(notifications, notificationID)).join(", ")
		: alertT("rules.noNotification");

	return alertT("rules.preview", {
		severity: alertT(`rules.severityOptions.${form.severity}`),
		metric,
		operator: alertT(`rules.operatorPhrases.${form.operator}`),
		threshold,
		window: formatDuration(numbers.windowSeconds.value),
		trigger: formatDuration(numbers.triggerAfterMinutes.value * 60),
		notification,
		cooldown: formatDuration(numbers.cooldownSeconds.value)
	});
}

export function stopTableAction(event: MouseEvent<HTMLButtonElement>) {
	event.stopPropagation();
}

export function optionValue<TValue extends string>(value: string | null, options: Array<{ value: TValue }>, fallback: TValue): TValue {
	return options.some(option => option.value === value) ? (value as TValue) : fallback;
}

export function pathWithSearch(path: string, search: string) {
	return search ? `${path}?${search}` : path;
}
