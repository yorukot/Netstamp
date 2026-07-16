import type { ApiAlertIncident, ApiAlertRule, ApiNotification, CreateAlertRuleInput, CreateNotificationInput, UpdateAlertRuleInput, UpdateNotificationInput } from "@/shared/api/types";
import type { BadgeTone } from "@netstamp/ui";
import type { MouseEvent } from "react";

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

export interface NotificationTypeOption {
	value: NotificationType;
	label: string;
	detail: string;
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

export const alertTabs: Array<{ value: AlertTab; label: string }> = [
	{ value: "incidents", label: "Incidents" },
	{ value: "rules", label: "Rules" },
	{ value: "notifications", label: "Notifications" }
];

export const incidentStatusOptions: Array<{ value: IncidentStatusFilter; label: string }> = [
	{ value: "open", label: "Open" },
	{ value: "acknowledged", label: "Acknowledged" },
	{ value: "resolved", label: "Resolved" },
	{ value: "all", label: "All" }
];

export const ruleStatusOptions: Array<{ value: RuleStatusFilter; label: string }> = [
	{ value: "all", label: "Any status" },
	{ value: "enabled", label: "Enabled" },
	{ value: "disabled", label: "Disabled" }
];

export const ruleCheckTypeOptions: Array<{ value: RuleCheckTypeFilter; label: string }> = [
	{ value: "all", label: "Any type" },
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute" },
	{ value: "http", label: "HTTP" }
];

export const notificationStatusOptions: Array<{ value: NotificationStatusFilter; label: string }> = [
	{ value: "all", label: "Any status" },
	{ value: "enabled", label: "Enabled" },
	{ value: "disabled", label: "Disabled" }
];

export const notificationFilterTypeOptions: Array<{ value: NotificationTypeFilter; label: string }> = [
	{ value: "all", label: "Any type" },
	{ value: "webhook", label: "Webhook" },
	{ value: "slack", label: "Slack" },
	{ value: "discord", label: "Discord" },
	{ value: "telegram", label: "Telegram" },
	{ value: "email", label: "Email" }
];

export const checkTypeOptions: Array<{ value: CheckType; label: string; disabled?: boolean }> = [
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute (alerts not available)", disabled: true },
	{ value: "http", label: "HTTP / HTTPS" }
];

export const metricOptions: Record<CheckType, Array<{ value: AlertMetric; label: string; unit?: string }>> = {
	ping: [
		{ value: "ping.loss_percent", label: "Ping loss percent", unit: "%" },
		{ value: "ping.average_rtt_ms", label: "Ping average RTT", unit: "ms" },
		{ value: "ping.max_rtt_ms", label: "Ping max RTT", unit: "ms" },
		{ value: "ping.success_rate", label: "Ping success rate", unit: "%" }
	],
	tcp: [
		{ value: "tcp.failure_percent", label: "TCP failure percent", unit: "%" },
		{ value: "tcp.average_connect_ms", label: "TCP average connect", unit: "ms" },
		{ value: "tcp.max_connect_ms", label: "TCP max connect", unit: "ms" },
		{ value: "tcp.success_rate", label: "TCP success rate", unit: "%" }
	],
	http: [
		{ value: "http.failure_percent", label: "HTTP failure percent", unit: "%" },
		{ value: "http.average_total_ms", label: "HTTP average total", unit: "ms" },
		{ value: "http.max_total_ms", label: "HTTP max total", unit: "ms" },
		{ value: "http.average_ttfb_ms", label: "HTTP average TTFB", unit: "ms" },
		{ value: "http.max_ttfb_ms", label: "HTTP max TTFB", unit: "ms" },
		{ value: "http.success_rate", label: "HTTP success rate", unit: "%" },
		{ value: "http.certificate_days_remaining", label: "Certificate days remaining", unit: "days" }
	],
	traceroute: []
};

export const severityOptions: Array<{ value: AlertSeverity; label: string }> = [
	{ value: "info", label: "Info" },
	{ value: "warning", label: "Warning" },
	{ value: "critical", label: "Critical" }
];

export const operatorOptions: Array<{ value: AlertOperator; label: string }> = [
	{ value: "gt", label: "Greater than" },
	{ value: "gte", label: "Greater or equal" },
	{ value: "lt", label: "Less than" },
	{ value: "lte", label: "Less or equal" },
	{ value: "eq", label: "Equal" }
];

export const enabledOptions = [
	{ value: "true", label: "Enabled" },
	{ value: "false", label: "Disabled" }
];

export const webhookNotificationTypeOption: NotificationTypeOption = { value: "webhook", label: "Webhook", detail: "Send raw alert JSON to any HTTPS endpoint." };

export const notificationTypeOptions: NotificationTypeOption[] = [
	webhookNotificationTypeOption,
	{ value: "slack", label: "Slack", detail: "Post alert summaries to a Slack incoming webhook." },
	{ value: "discord", label: "Discord", detail: "Post alert summaries to a Discord notification webhook." },
	{ value: "telegram", label: "Telegram", detail: "Send alert summaries through a Telegram bot." },
	{ value: "email", label: "Email", detail: "Send alert summaries to one or more email recipients." }
];

export const operatorSymbols: Record<AlertOperator, string> = {
	gt: ">",
	gte: ">=",
	lt: "<",
	lte: "<=",
	eq: "="
};

export const operatorPhrases: Record<AlertOperator, string> = {
	gt: "is greater than",
	gte: "is at least",
	lt: "is less than",
	lte: "is at most",
	eq: "equals"
};

export function formatDateTime(value?: string | null) {
	if (!value) {
		return "-";
	}
	const date = new Date(value);

	if (Number.isNaN(date.getTime())) {
		return "-";
	}

	return date.toLocaleString();
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
		return `${days} ${days === 1 ? "day" : "days"}`;
	}

	if (seconds % 3600 === 0) {
		const hours = seconds / 3600;
		return `${hours} ${hours === 1 ? "hour" : "hours"}`;
	}

	if (seconds % 60 === 0) {
		const minutes = seconds / 60;
		return `${minutes} ${minutes === 1 ? "min" : "min"}`;
	}

	return `${seconds} sec`;
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
	return metricOption(metric)?.label ?? metric.replace(/_/g, " ");
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
	return `${metricLabel(condition.metric)} ${operatorSymbols[condition.operator]} ${formatThreshold(condition.metric, condition.threshold)} for ${formatDuration(condition.windowSeconds)}`;
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
		parts.push(`probe ${shortID(rule.scope.probeId)}`);
	}
	if (rule.scope.checkId) {
		parts.push(`check ${shortID(rule.scope.checkId)}`);
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
	return `Probe ${incident.probeId} / Check ${incident.checkId}`;
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
			return "Webhook";
		case "slack":
			return "Slack";
		case "discord":
			return "Discord";
		case "telegram":
			return "Telegram";
		case "email":
			return "Email";
		default:
			return notificationType;
	}
}

export function notificationTypeOption(notificationType: NotificationType) {
	return notificationTypeOptions.find(option => option.value === notificationType) ?? webhookNotificationTypeOption;
}

export function notificationWebhookURLLabel(notificationType: NotificationType) {
	switch (notificationType) {
		case "slack":
			return "Slack webhook URL";
		case "discord":
			return "Discord webhook URL";
		default:
			return "Webhook URL";
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
			return notificationChatID(notification) ? `chat ${notificationChatID(notification)}` : "-";
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
		return "No notification";
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

	return [{ value: form.metric, label: metricLabel(form.metric), unit: metricUnit(form.metric) }];
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
		return { value: Number.NaN, error: `${label} is required.` };
	}

	const parsed = Number(trimmed);

	if (!Number.isFinite(parsed)) {
		return { value: Number.NaN, error: `${label} must be a number.` };
	}

	if (options.integer && !Number.isInteger(parsed)) {
		return { value: parsed, error: `${label} must be a whole number.` };
	}

	if (typeof options.min === "number" && parsed < options.min) {
		return { value: parsed, error: `${label} must be at least ${options.min}.` };
	}

	if (typeof options.max === "number" && parsed > options.max) {
		return { value: parsed, error: `${label} must be at most ${options.max}.` };
	}

	return { value: parsed, error: "" };
}

export function validateRuleNumbers(form: RuleFormState): RuleNumberValidation {
	return {
		threshold: validateNumericField("Threshold", form.threshold, { min: 0 }),
		windowSeconds: validateNumericField("Window seconds", form.windowSeconds, { integer: true, min: 60, max: 86400 }),
		minSamples: validateNumericField("Min samples", form.minSamples, { integer: true, min: 1, max: 10000 }),
		triggerAfterMinutes: validateNumericField("Trigger after", form.triggerAfterMinutes, { integer: true, min: 1, max: 1440 }),
		cooldownSeconds: validateNumericField("Cooldown", form.cooldownSeconds, { integer: true, min: 60, max: 86400 })
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
	const metric = metricLabel(form.metric).toLowerCase();
	const threshold = formatThreshold(form.metric, form.threshold || "0");
	const notification = form.selectedNotificationIds.length ? form.selectedNotificationIds.map(notificationID => notificationNameByID(notifications, notificationID)).join(", ") : "no notification";

	return `Create a ${form.severity} incident when ${metric} ${operatorPhrases[form.operator]} ${threshold}, evaluated over ${formatDuration(numbers.windowSeconds.value)}, remains firing for ${formatDuration(numbers.triggerAfterMinutes.value * 60)}. Notify ${notification}, then wait ${formatDuration(numbers.cooldownSeconds.value)} before repeating.`;
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
