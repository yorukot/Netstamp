import { pathForAlertIncidentDetail, pathForRoute } from "@/routes/routePaths";
import {
	useCreateProjectAlertRuleMutation,
	useCreateProjectNotificationMutation,
	useDeleteProjectAlertRuleMutation,
	useDeleteProjectNotificationMutation,
	useTestProjectNotificationMutation,
	useUpdateProjectAlertRuleMutation,
	useUpdateProjectNotificationMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiAlertIncident, ApiAlertRule, ApiNotification, CreateAlertRuleInput, CreateNotificationInput, UpdateAlertRuleInput, UpdateNotificationInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { LoadingState } from "@/shared/components/LoadingState";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DataTable, Panel, SelectField, Tabs, TextAreaField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { DiscordLogo, EnvelopeSimple, SlackLogo, TelegramLogo, WebhooksLogo } from "@phosphor-icons/react";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState, type FormEvent, type MouseEvent } from "react";
import { useNavigate, useParams } from "react-router-dom";
import styles from "./AlertsPage.module.css";

type AlertTab = "incidents" | "rules" | "notifications";
type IncidentStatusFilter = "open" | "acknowledged" | "resolved" | "all";
type RuleStatusFilter = "all" | "enabled" | "disabled";
type CheckType = CreateAlertRuleInput["scope"]["checkType"];
type RuleCheckTypeFilter = "all" | CheckType;
type NotificationStatusFilter = "all" | "enabled" | "disabled";
type NotificationType = CreateNotificationInput["type"];
type NotificationTypeFilter = "all" | NotificationType;
type AlertMetric = CreateAlertRuleInput["condition"]["metric"];
type AlertOperator = CreateAlertRuleInput["condition"]["operator"];
type AlertSeverity = CreateAlertRuleInput["severity"];

interface RuleEditorState {
	mode: "create" | "edit";
	rule?: ApiAlertRule;
}

interface NotificationEditorState {
	mode: "create" | "edit";
	notification?: ApiNotification;
}

type NotificationEditorStep = "type" | "detail";

interface RuleFormState {
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
	cooldownSeconds: string;
	selectedNotificationIds: string[];
}

interface NotificationFormState {
	name: string;
	type: NotificationType;
	url: string;
	botToken: string;
	chatId: string;
	emailTo: string;
	enabled: string;
}

interface NotificationTypeOption {
	value: NotificationType;
	label: string;
	detail: string;
}

interface NumericFieldValidation {
	value: number;
	error: string;
}

interface RuleNumberValidation {
	threshold: NumericFieldValidation;
	windowSeconds: NumericFieldValidation;
	minSamples: NumericFieldValidation;
	cooldownSeconds: NumericFieldValidation;
}

const emptyRules: ApiAlertRule[] = [];
const emptyIncidents: ApiAlertIncident[] = [];
const emptyNotifications: ApiNotification[] = [];

const alertTabs: Array<{ value: AlertTab; label: string }> = [
	{ value: "incidents", label: "Incidents" },
	{ value: "rules", label: "Rules" },
	{ value: "notifications", label: "Notifications" }
];

const incidentStatusOptions: Array<{ value: IncidentStatusFilter; label: string }> = [
	{ value: "open", label: "Open" },
	{ value: "acknowledged", label: "Acknowledged" },
	{ value: "resolved", label: "Resolved" },
	{ value: "all", label: "All" }
];

const ruleStatusOptions: Array<{ value: RuleStatusFilter; label: string }> = [
	{ value: "all", label: "Any status" },
	{ value: "enabled", label: "Enabled" },
	{ value: "disabled", label: "Disabled" }
];

const ruleCheckTypeOptions: Array<{ value: RuleCheckTypeFilter; label: string }> = [
	{ value: "all", label: "Any type" },
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute" }
];

const notificationStatusOptions: Array<{ value: NotificationStatusFilter; label: string }> = [
	{ value: "all", label: "Any status" },
	{ value: "enabled", label: "Enabled" },
	{ value: "disabled", label: "Disabled" }
];

const notificationFilterTypeOptions: Array<{ value: NotificationTypeFilter; label: string }> = [
	{ value: "all", label: "Any type" },
	{ value: "webhook", label: "Webhook" },
	{ value: "slack", label: "Slack" },
	{ value: "discord", label: "Discord" },
	{ value: "telegram", label: "Telegram" },
	{ value: "email", label: "Email" }
];

const checkTypeOptions: Array<{ value: CheckType; label: string; disabled?: boolean }> = [
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" },
	{ value: "traceroute", label: "Traceroute (alerts not available)", disabled: true }
];

const metricOptions: Record<CheckType, Array<{ value: AlertMetric; label: string; unit?: string }>> = {
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
	traceroute: []
};

const severityOptions: Array<{ value: AlertSeverity; label: string }> = [
	{ value: "info", label: "Info" },
	{ value: "warning", label: "Warning" },
	{ value: "critical", label: "Critical" }
];

const operatorOptions: Array<{ value: AlertOperator; label: string }> = [
	{ value: "gt", label: "Greater than" },
	{ value: "gte", label: "Greater or equal" },
	{ value: "lt", label: "Less than" },
	{ value: "lte", label: "Less or equal" },
	{ value: "eq", label: "Equal" }
];

const enabledOptions = [
	{ value: "true", label: "Enabled" },
	{ value: "false", label: "Disabled" }
];

const webhookNotificationTypeOption: NotificationTypeOption = { value: "webhook", label: "Webhook", detail: "Send raw alert JSON to any HTTPS endpoint." };

const notificationTypeOptions: NotificationTypeOption[] = [
	webhookNotificationTypeOption,
	{ value: "slack", label: "Slack", detail: "Post alert summaries to a Slack incoming webhook." },
	{ value: "discord", label: "Discord", detail: "Post alert summaries to a Discord notification webhook." },
	{ value: "telegram", label: "Telegram", detail: "Send alert summaries through a Telegram bot." },
	{ value: "email", label: "Email", detail: "Send alert summaries to one or more email recipients." }
];

const operatorSymbols: Record<AlertOperator, string> = {
	gt: ">",
	gte: ">=",
	lt: "<",
	lte: "<=",
	eq: "="
};

const operatorPhrases: Record<AlertOperator, string> = {
	gt: "is greater than",
	gte: "is at least",
	lt: "is less than",
	lte: "is at most",
	eq: "equals"
};

function formatDateTime(value?: string | null) {
	if (!value) {
		return "-";
	}
	const date = new Date(value);

	if (Number.isNaN(date.getTime())) {
		return "-";
	}

	return date.toLocaleString();
}

function shortID(value: string) {
	return value.slice(0, 8);
}

function formatDuration(seconds: number) {
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

function severityTone(severity: string): BadgeTone {
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

function incidentTone(status: string): BadgeTone {
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

function metricOption(metric: AlertMetric) {
	return Object.values(metricOptions)
		.flat()
		.find(option => option.value === metric);
}

function metricLabel(metric: AlertMetric) {
	return metricOption(metric)?.label ?? metric.replace(/_/g, " ");
}

function metricUnit(metric: AlertMetric) {
	return metricOption(metric)?.unit ?? "";
}

function formatThreshold(metric: AlertMetric, threshold: number | string) {
	const value = typeof threshold === "number" ? String(threshold) : threshold;
	const unit = metricUnit(metric);
	return unit ? `${value}${unit}` : value;
}

function formatAlertCondition(condition: ApiAlertRule["condition"]) {
	return `${metricLabel(condition.metric)} ${operatorSymbols[condition.operator]} ${formatThreshold(condition.metric, condition.threshold)} for ${formatDuration(condition.windowSeconds)}`;
}

function formatIncidentReason(incident: ApiAlertIncident) {
	const summary = incident.lastSummary;
	if (summary.operator && typeof summary.threshold === "number") {
		return `${metricLabel(summary.metric)} ${operatorSymbols[summary.operator]} ${formatThreshold(summary.metric, summary.threshold)}`;
	}

	return metricLabel(summary.metric);
}

function formatRuleScope(rule: ApiAlertRule) {
	const parts = [rule.scope.checkType.toUpperCase()];
	if (rule.scope.probeId) {
		parts.push(`probe ${shortID(rule.scope.probeId)}`);
	}
	if (rule.scope.checkId) {
		parts.push(`check ${shortID(rule.scope.checkId)}`);
	}
	return parts.join(" / ");
}

function incidentProbeName(incident: ApiAlertIncident) {
	return incident.probe?.name || shortID(incident.probeId);
}

function incidentCheckName(incident: ApiAlertIncident) {
	return incident.check?.name || shortID(incident.checkId);
}

function incidentCheckTarget(incident: ApiAlertIncident) {
	return incident.check?.target || shortID(incident.checkId);
}

function incidentTargetTitle(incident: ApiAlertIncident) {
	return `Probe ${incident.probeId} / Check ${incident.checkId}`;
}

function formatIncidentProbe(incident: ApiAlertIncident) {
	return `${incidentProbeName(incident)} (${shortID(incident.probeId)})`;
}

function formatIncidentCheck(incident: ApiAlertIncident) {
	return `${incidentCheckName(incident)} (${incident.checkType.toUpperCase()} / ${shortID(incident.checkId)})`;
}

function notificationConfigString(config: ApiNotification["config"], key: string) {
	if (config && typeof config === "object" && key in config) {
		const value = (config as Record<string, unknown>)[key];
		return typeof value === "string" ? value : "";
	}

	return "";
}

function notificationConfigStringArray(config: ApiNotification["config"], key: string) {
	if (config && typeof config === "object" && key in config) {
		const value = (config as Record<string, unknown>)[key];
		return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : [];
	}

	return [];
}

function notificationURL(notification: ApiNotification) {
	return notificationConfigString(notification.config, "url");
}

function notificationChatID(notification: ApiNotification) {
	return notificationConfigString(notification.config, "chatId");
}

function notificationEmails(notification: ApiNotification) {
	return notificationConfigStringArray(notification.config, "to");
}

function notificationTypeLabel(notificationType: string) {
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

function notificationTypeOption(notificationType: NotificationType) {
	return notificationTypeOptions.find(option => option.value === notificationType) ?? webhookNotificationTypeOption;
}

function notificationWebhookURLLabel(notificationType: NotificationType) {
	switch (notificationType) {
		case "slack":
			return "Slack webhook URL";
		case "discord":
			return "Discord webhook URL";
		default:
			return "Webhook URL";
	}
}

function notificationWebhookURLPlaceholder(notificationType: NotificationType) {
	switch (notificationType) {
		case "slack":
			return "https://hooks.slack.com/services/...";
		case "discord":
			return "https://discord.com/api/webhooks/...";
		default:
			return "https://hooks.example.com/netstamp";
	}
}

function NotificationTypeIcon({ type }: { type: NotificationType }) {
	switch (type) {
		case "slack":
			return <SlackLogo className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "discord":
			return <DiscordLogo className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "telegram":
			return <TelegramLogo className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "email":
			return <EnvelopeSimple className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "webhook":
		default:
			return <WebhooksLogo className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" />;
	}
}

function notificationDestination(notification: ApiNotification) {
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

function notificationNameByID(notifications: ApiNotification[], notificationID: string) {
	return notifications.find(notification => notification.id === notificationID)?.name ?? shortID(notificationID);
}

function notificationLabel(rule: ApiAlertRule, notifications: ApiNotification[]) {
	if (!rule.notificationIds.length) {
		return "No notification";
	}

	return rule.notificationIds.map(notificationID => notificationNameByID(notifications, notificationID)).join(", ");
}

function rulesUsingNotification(rules: ApiAlertRule[], notificationID: string) {
	return rules.filter(rule => rule.notificationIds.includes(notificationID));
}

function metricForCheckType(nextCheckType: CheckType) {
	return metricOptions[nextCheckType][0]?.value;
}

function supportsAlertMetrics(checkType: CheckType) {
	return metricOptions[checkType].length > 0;
}

function metricOptionsForForm(form: RuleFormState) {
	const options = metricOptions[form.checkType];

	if (options.length) {
		return options;
	}

	return [{ value: form.metric, label: metricLabel(form.metric), unit: metricUnit(form.metric) }];
}

function defaultRuleForm(): RuleFormState {
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
		cooldownSeconds: "900",
		selectedNotificationIds: []
	};
}

function ruleFormFromRule(rule: ApiAlertRule): RuleFormState {
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
		cooldownSeconds: String(rule.cooldownSeconds),
		selectedNotificationIds: rule.notificationIds
	};
}

function defaultNotificationForm(): NotificationFormState {
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

function notificationFormFromNotification(notification: ApiNotification): NotificationFormState {
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

function validateNumericField(label: string, value: string, options: { integer?: boolean; min?: number; max?: number } = {}): NumericFieldValidation {
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

function validateRuleNumbers(form: RuleFormState): RuleNumberValidation {
	return {
		threshold: validateNumericField("Threshold", form.threshold, { min: 0 }),
		windowSeconds: validateNumericField("Window seconds", form.windowSeconds, { integer: true, min: 60, max: 86400 }),
		minSamples: validateNumericField("Min samples", form.minSamples, { integer: true, min: 1, max: 10000 }),
		cooldownSeconds: validateNumericField("Cooldown", form.cooldownSeconds, { integer: true, min: 60, max: 86400 })
	};
}

function ruleNumberError(validation: RuleNumberValidation) {
	return validation.threshold.error || validation.windowSeconds.error || validation.minSamples.error || validation.cooldownSeconds.error;
}

function rulePayload(form: RuleFormState): CreateAlertRuleInput | UpdateAlertRuleInput {
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
		cooldownSeconds: numbers.cooldownSeconds.value,
		notificationIds: form.selectedNotificationIds
	};
}

function notificationPayload(form: NotificationFormState): CreateNotificationInput | UpdateNotificationInput {
	const config =
		form.type === "telegram" ? { botToken: form.botToken.trim(), chatId: form.chatId.trim() } : form.type === "email" ? { to: notificationEmailRecipients(form.emailTo) } : { url: form.url.trim() };

	return {
		name: form.name.trim(),
		type: form.type,
		enabled: form.enabled === "true",
		config
	};
}

function notificationFormReady(form: NotificationFormState) {
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

function notificationEmailRecipients(value: string) {
	return value
		.split(/[\n,;]+/)
		.map(item => item.trim())
		.filter(Boolean);
}

function rulePreview(form: RuleFormState, notifications: ApiNotification[], numbers: RuleNumberValidation) {
	const metric = metricLabel(form.metric).toLowerCase();
	const threshold = formatThreshold(form.metric, form.threshold || "0");
	const notification = form.selectedNotificationIds.length ? form.selectedNotificationIds.map(notificationID => notificationNameByID(notifications, notificationID)).join(", ") : "no notification";

	return `Create a ${form.severity} incident when ${metric} ${operatorPhrases[form.operator]} ${threshold} for ${formatDuration(numbers.windowSeconds.value)}. Notify ${notification}, then wait ${formatDuration(numbers.cooldownSeconds.value)} before repeating.`;
}

function stopTableAction(event: MouseEvent<HTMLButtonElement>) {
	event.stopPropagation();
}

function tableState(label: string, detail: string) {
	return <LoadingState label={label} detail={detail} size="compact" />;
}

export function AlertsPage() {
	const confirm = useConfirm();
	const navigate = useNavigate();
	const { incidentId = "" } = useParams();
	const { projectRef } = useCurrentProject();
	const createRuleMutation = useCreateProjectAlertRuleMutation(projectRef);
	const updateRuleMutation = useUpdateProjectAlertRuleMutation(projectRef);
	const deleteRuleMutation = useDeleteProjectAlertRuleMutation(projectRef);
	const createNotificationMutation = useCreateProjectNotificationMutation(projectRef);
	const updateNotificationMutation = useUpdateProjectNotificationMutation(projectRef);
	const deleteNotificationMutation = useDeleteProjectNotificationMutation(projectRef);
	const testNotificationMutation = useTestProjectNotificationMutation(projectRef);
	const [activeTab, setActiveTab] = useState<AlertTab>("incidents");
	const [incidentStatus, setIncidentStatus] = useState<IncidentStatusFilter>("open");
	const [ruleSearch, setRuleSearch] = useState("");
	const [ruleStatus, setRuleStatus] = useState<RuleStatusFilter>("all");
	const [ruleCheckType, setRuleCheckType] = useState<RuleCheckTypeFilter>("all");
	const [notificationStatus, setNotificationStatus] = useState<NotificationStatusFilter>("all");
	const [notificationType, setNotificationType] = useState<NotificationTypeFilter>("all");
	const [ruleEditor, setRuleEditor] = useState<RuleEditorState | null>(null);
	const [notificationEditor, setNotificationEditor] = useState<NotificationEditorState | null>(null);
	const incidentFilters = useMemo(() => ({ limit: 100, ...(incidentStatus === "all" ? {} : { status: incidentStatus }) }), [incidentStatus]);
	const rulesQuery = useQuery({
		...projectQueries.alertRules(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const openIncidentsQuery = useQuery({
		...projectQueries.alertIncidents(projectRef || "", { status: "open", limit: 100 }),
		enabled: Boolean(projectRef),
		refetchInterval: 30 * 1000
	});
	const incidentsQuery = useQuery({
		...projectQueries.alertIncidents(projectRef || "", incidentFilters),
		enabled: Boolean(projectRef),
		refetchInterval: 30 * 1000
	});
	const incidentDetailQuery = useQuery({
		...projectQueries.alertIncidentDetail(projectRef || "", incidentId),
		enabled: Boolean(projectRef && incidentId),
		refetchInterval: 30 * 1000
	});
	const notificationsQuery = useQuery({
		...projectQueries.notifications(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const rules = rulesQuery.data?.rules ?? emptyRules;
	const incidents = incidentsQuery.data?.incidents ?? emptyIncidents;
	const openIncidents = openIncidentsQuery.data?.incidents ?? emptyIncidents;
	const notifications = notificationsQuery.data?.notifications ?? emptyNotifications;
	const selectedIncident = incidentDetailQuery.data?.incident ?? incidents.find(incident => incident.id === incidentId) ?? null;
	const visibleTab = incidentId ? "incidents" : activeTab;
	const filteredRules = useMemo(() => {
		const search = ruleSearch.trim().toLowerCase();

		return rules.filter(rule => {
			if (ruleStatus === "enabled" && !rule.enabled) {
				return false;
			}
			if (ruleStatus === "disabled" && rule.enabled) {
				return false;
			}
			if (ruleCheckType !== "all" && rule.scope.checkType !== ruleCheckType) {
				return false;
			}
			if (!search) {
				return true;
			}
			return [rule.name, rule.description, rule.scope.checkType, formatAlertCondition(rule.condition), notificationLabel(rule, notifications)].filter(Boolean).join(" ").toLowerCase().includes(search);
		});
	}, [notifications, ruleCheckType, ruleSearch, ruleStatus, rules]);
	const filteredNotifications = useMemo(
		() =>
			notifications.filter(notification => {
				if (notificationStatus === "enabled" && !notification.enabled) {
					return false;
				}
				if (notificationStatus === "disabled" && notification.enabled) {
					return false;
				}
				if (notificationType !== "all" && notification.type !== notificationType) {
					return false;
				}
				return true;
			}),
		[notificationStatus, notificationType, notifications]
	);
	const enabledRules = rules.filter(rule => rule.enabled).length;
	const ruleColumns: DataColumn<ApiAlertRule>[] = [
		{
			key: "name",
			label: "Rule",
			render: rule => (
				<div className={styles.primaryCell}>
					<strong>{rule.name}</strong>
					<span>{rule.description || "No description"}</span>
				</div>
			)
		},
		{
			key: "status",
			label: "Status",
			render: rule => <Badge tone={rule.enabled ? "success" : "muted"}>{rule.enabled ? "enabled" : "disabled"}</Badge>
		},
		{
			key: "scope",
			label: "Scope",
			render: rule => <span className={styles.monoCell}>{formatRuleScope(rule)}</span>
		},
		{
			key: "condition",
			label: "Condition",
			render: rule => <span className={styles.monoCell}>{formatAlertCondition(rule.condition)}</span>
		},
		{
			key: "notify",
			label: "Notify",
			render: rule => <span className={styles.urlCell}>{notificationLabel(rule, notifications)}</span>
		},
		{
			key: "cooldown",
			label: "Cooldown",
			render: rule => formatDuration(rule.cooldownSeconds)
		},
		{
			key: "updatedAt",
			label: "Updated",
			render: rule => formatDateTime(rule.updatedAt)
		},
		{
			key: "delete",
			label: "",
			render: rule => (
				<Button
					variant="danger"
					size="sm"
					disabled={deleteRuleMutation.isPending}
					onClick={event => {
						stopTableAction(event);
						void deleteRule(rule);
					}}
				>
					Delete
				</Button>
			)
		}
	];
	const incidentColumns: DataColumn<ApiAlertIncident>[] = [
		{
			key: "severity",
			label: "Severity",
			render: incident => <Badge tone={severityTone(incident.severity)}>{incident.severity}</Badge>
		},
		{
			key: "target",
			label: "Target",
			render: incident => (
				<div className={styles.primaryCell}>
					<strong title={incidentTargetTitle(incident)}>{incidentCheckName(incident)}</strong>
					<span title={incidentTargetTitle(incident)}>
						{incidentProbeName(incident)} / {incident.checkType.toUpperCase()} / {incidentCheckTarget(incident)}
					</span>
				</div>
			)
		},
		{
			key: "why",
			label: "Why",
			render: incident => (
				<div className={styles.primaryCell}>
					<strong>{formatIncidentReason(incident)}</strong>
					<span>{incident.lastEvaluationState}</span>
				</div>
			)
		},
		{
			key: "value",
			label: "Value",
			render: incident => (typeof incident.lastValue === "number" ? formatThreshold(incident.lastSummary.metric, Number(incident.lastValue.toFixed(2))) : "-")
		},
		{
			key: "openedAt",
			label: "Opened",
			render: incident => formatDateTime(incident.openedAt)
		},
		{
			key: "lastEvaluatedAt",
			label: "Last checked",
			render: incident => formatDateTime(incident.lastEvaluatedAt)
		},
		{
			key: "notifications",
			label: "Notifications",
			render: incident => (incident.suppressedNotificationCount > 0 ? `${incident.suppressedNotificationCount} suppressed` : formatDateTime(incident.lastNotificationSentAt))
		}
	];
	const notificationColumns: DataColumn<ApiNotification>[] = [
		{
			key: "name",
			label: "Notification",
			render: notification => (
				<div className={styles.primaryCell}>
					<strong>{notification.name}</strong>
					<span>{notificationTypeLabel(notification.type)}</span>
				</div>
			)
		},
		{
			key: "status",
			label: "Status",
			render: notification => <Badge tone={notification.enabled ? "success" : "muted"}>{notification.enabled ? "enabled" : "disabled"}</Badge>
		},
		{
			key: "url",
			label: "Destination",
			render: notification => (
				<span className={styles.urlCell} title={notificationDestination(notification)}>
					{notificationDestination(notification)}
				</span>
			)
		},
		{
			key: "usedBy",
			label: "Used by rules",
			render: notification => rulesUsingNotification(rules, notification.id).length
		},
		{
			key: "actions",
			label: "",
			render: notification => (
				<div className={styles.rowActions}>
					<Button
						variant="secondary"
						size="sm"
						disabled={testNotificationMutation.isPending}
						onClick={event => {
							stopTableAction(event);
							void testNotification(notification);
						}}
					>
						Test
					</Button>
					<Button
						variant="danger"
						size="sm"
						disabled={deleteNotificationMutation.isPending}
						onClick={event => {
							stopTableAction(event);
							void deleteNotification(notification);
						}}
					>
						Delete
					</Button>
				</div>
			)
		}
	];

	function selectIncident(incident: ApiAlertIncident) {
		navigate(pathForAlertIncidentDetail(projectRef, incident.id));
	}

	function closeIncidentDetail() {
		navigate(pathForRoute("alerts", { projectRef }));
	}

	async function deleteRule(rule: ApiAlertRule) {
		const accepted = await confirm({
			title: "Delete alert rule?",
			message: `This removes "${rule.name}". Existing incidents stay in history, but this rule will stop evaluating.`,
			confirmLabel: "Delete rule",
			tone: "danger"
		});

		if (!accepted) {
			return;
		}

		try {
			await deleteRuleMutation.mutateAsync(rule.id);
			pushToast({ title: "Rule deleted", message: rule.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function deleteNotification(notification: ApiNotification) {
		const usedBy = rulesUsingNotification(rules, notification.id).length;
		const accepted = await confirm({
			title: "Delete notification?",
			message: usedBy
				? `"${notification.name}" is used by ${usedBy} rule${usedBy === 1 ? "" : "s"}. Remove it only after moving those rules to another notification.`
				: `This removes "${notification.name}".`,
			confirmLabel: "Delete notification",
			tone: "danger"
		});

		if (!accepted) {
			return;
		}

		try {
			await deleteNotificationMutation.mutateAsync(notification.id);
			pushToast({ title: "Notification deleted", message: notification.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function testNotification(notification: ApiNotification) {
		try {
			const response = await testNotificationMutation.mutateAsync(notification.id);
			if (response.result.delivered) {
				pushToast({ title: "Test delivered", message: notification.name, tone: "success" });
				return;
			}
			pushToast({ title: "Test failed", message: response.result.message || response.result.code || notification.name, tone: "critical" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	return (
		<PageStack>
			<ScreenHeader title="Alerts" />
			<section className={styles.summaryGrid} aria-label="Alert summary">
				<SummaryCard
					label="Open incidents"
					value={openIncidents.length}
					tone={openIncidents.length ? "critical" : "success"}
					detail={openIncidentsQuery.isLoading ? "Loading" : openIncidents.length ? "Needs attention" : "No current incidents"}
				/>
				<SummaryCard label="Enabled rules" value={enabledRules} tone={enabledRules ? "success" : "muted"} detail={`${rules.length} total rules`} />
				<SummaryCard
					label="Notifications"
					value={notifications.length}
					tone={notifications.length ? "accent" : "muted"}
					detail={notifications.length ? "Ready to notify" : "No notification configured"}
				/>
			</section>
			<Tabs tabs={alertTabs} value={visibleTab} ariaLabel="Alert sections" onValueChange={value => setActiveTab(value as AlertTab)} />
			{visibleTab === "incidents" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.singlePanelAction}>
							<SelectField label="Status" value={incidentStatus} options={incidentStatusOptions} onChange={event => setIncidentStatus(event.currentTarget.value as IncidentStatusFilter)} />
						</div>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={incidentColumns}
						rows={incidents}
						density="compact"
						minWidth="58rem"
						getRowKey={incident => incident.id}
						getRowAriaLabel={incident => `Open incident ${shortID(incident.id)}`}
						onRowClick={selectIncident}
						selectedKey={incidentId || undefined}
						emptyLabel={
							incidentsQuery.isLoading
								? tableState("Loading incidents", "Fetching current alert incidents for this project.")
								: incidentStatus === "open"
									? "No open incidents"
									: "No incidents match this view"
						}
					/>
				</Panel>
			) : null}
			{visibleTab === "rules" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.panelActions}>
							<TextField label="Search" value={ruleSearch} onChange={event => setRuleSearch(event.currentTarget.value)} placeholder="loss, RTT, notification" />
							<SelectField label="Status" value={ruleStatus} options={ruleStatusOptions} onChange={event => setRuleStatus(event.currentTarget.value as RuleStatusFilter)} />
							<SelectField label="Type" value={ruleCheckType} options={ruleCheckTypeOptions} onChange={event => setRuleCheckType(event.currentTarget.value as RuleCheckTypeFilter)} />
						</div>
						<Button type="button" onClick={() => setRuleEditor({ mode: "create" })}>
							Create rule
						</Button>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={ruleColumns}
						rows={filteredRules}
						density="compact"
						minWidth="72rem"
						getRowKey={rule => rule.id}
						getRowAriaLabel={rule => `Edit alert rule ${rule.name}`}
						onRowClick={rule => setRuleEditor({ mode: "edit", rule })}
						selectedKey={ruleEditor?.rule?.id}
						emptyLabel={
							rulesQuery.isLoading ? (
								tableState("Loading alert rules", "Fetching rule conditions and notification targets.")
							) : rules.length ? (
								"No alert rules match this view"
							) : (
								<EmptyAction label="No alert rules yet" action="Create rule" onClick={() => setRuleEditor({ mode: "create" })} />
							)
						}
					/>
				</Panel>
			) : null}
			{visibleTab === "notifications" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.notificationPanelActions}>
							<SelectField
								label="Status"
								value={notificationStatus}
								options={notificationStatusOptions}
								onChange={event => setNotificationStatus(event.currentTarget.value as NotificationStatusFilter)}
							/>
							<SelectField label="Type" value={notificationType} options={notificationFilterTypeOptions} onChange={event => setNotificationType(event.currentTarget.value as NotificationTypeFilter)} />
						</div>
						<Button type="button" onClick={() => setNotificationEditor({ mode: "create" })}>
							Add notification
						</Button>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={notificationColumns}
						rows={filteredNotifications}
						density="compact"
						minWidth="52rem"
						getRowKey={notification => notification.id}
						getRowAriaLabel={notification => `Edit notification ${notification.name}`}
						onRowClick={notification => setNotificationEditor({ mode: "edit", notification })}
						selectedKey={notificationEditor?.notification?.id}
						emptyLabel={
							notificationsQuery.isLoading ? (
								tableState("Loading notifications", "Fetching notifications.")
							) : notifications.length ? (
								"No notifications match this view"
							) : (
								<EmptyAction label="No notifications yet" action="Add notification" onClick={() => setNotificationEditor({ mode: "create" })} />
							)
						}
					/>
				</Panel>
			) : null}
			{incidentId ? (
				<IncidentDetailDrawer
					incident={selectedIncident}
					isLoading={!selectedIncident && incidentDetailQuery.isPending}
					error={selectedIncident ? null : incidentDetailQuery.error}
					onClose={closeIncidentDetail}
				/>
			) : null}
			{ruleEditor ? (
				<RuleEditorDrawer
					editor={ruleEditor}
					notifications={notifications}
					isPending={createRuleMutation.isPending || updateRuleMutation.isPending}
					onClose={() => setRuleEditor(null)}
					onSubmit={async form => {
						try {
							const body = rulePayload(form);
							if (ruleEditor.mode === "edit" && ruleEditor.rule) {
								await updateRuleMutation.mutateAsync({ ruleId: ruleEditor.rule.id, body });
								pushToast({ title: "Rule updated", message: body.name, tone: "success" });
							} else {
								await createRuleMutation.mutateAsync(body);
								pushToast({ title: "Rule created", message: body.name, tone: "success" });
							}
							setRuleEditor(null);
						} catch (error) {
							pushErrorToast(requestErrorMessage(error));
						}
					}}
				/>
			) : null}
			{notificationEditor ? (
				<NotificationEditorDrawer
					editor={notificationEditor}
					isPending={createNotificationMutation.isPending || updateNotificationMutation.isPending}
					onClose={() => setNotificationEditor(null)}
					onSubmit={async form => {
						try {
							if (notificationEditor.mode === "edit" && notificationEditor.notification) {
								const body = notificationPayload({ ...form, type: notificationEditor.notification.type });
								await updateNotificationMutation.mutateAsync({ notificationId: notificationEditor.notification.id, body });
								pushToast({ title: "Notification updated", message: body.name, tone: "success" });
							} else {
								const body = notificationPayload(form);
								await createNotificationMutation.mutateAsync(body);
								pushToast({ title: "Notification created", message: body.name, tone: "success" });
							}
							setNotificationEditor(null);
						} catch (error) {
							pushErrorToast(requestErrorMessage(error));
						}
					}}
				/>
			) : null}
		</PageStack>
	);
}

function SummaryCard({ label, value, tone, detail }: { label: string; value: number; tone: BadgeTone; detail: string }) {
	return (
		<Panel tone="matte" title={label}>
			<div className={styles.summaryContent}>
				<strong className={styles.summaryValue}>{value}</strong>
				<Badge tone={tone}>{detail}</Badge>
			</div>
		</Panel>
	);
}

function EmptyAction({ label, action, onClick }: { label: string; action: string; onClick: () => void }) {
	return (
		<div className={styles.emptyAction}>
			<span>{label}</span>
			<Button type="button" size="sm" variant="secondary" onClick={onClick}>
				{action}
			</Button>
		</div>
	);
}

function IncidentDetailDrawer({ incident, isLoading, error, onClose }: { incident: ApiAlertIncident | null; isLoading: boolean; error: unknown; onClose: () => void }) {
	return (
		<EditorDrawer open title="Incident detail" ariaLabel="Incident detail" backLabel="back to incidents" onClose={onClose}>
			{incident ? (
				<div className={styles.detailStack}>
					<div className={styles.detailHeader}>
						<Badge tone={incidentTone(incident.status)}>{incident.status}</Badge>
						<Badge tone={severityTone(incident.severity)}>{incident.severity}</Badge>
					</div>
					<Panel tone="matte" title="What happened">
						<p className={styles.detailLead}>{formatIncidentReason(incident)}</p>
						<div className={styles.keyValueGrid}>
							<KeyValue label="Probe" value={formatIncidentProbe(incident)} />
							<KeyValue label="Check" value={formatIncidentCheck(incident)} />
							<KeyValue label="Target" value={incidentCheckTarget(incident)} />
							<KeyValue label="State" value={incident.lastEvaluationState} />
							<KeyValue label="Value" value={typeof incident.lastValue === "number" ? formatThreshold(incident.lastSummary.metric, Number(incident.lastValue.toFixed(2))) : "-"} />
							<KeyValue label="Rule" value={shortID(incident.ruleId)} />
						</div>
					</Panel>
					<Panel tone="matte" title="Timeline">
						<div className={styles.keyValueGrid}>
							<KeyValue label="Opened" value={formatDateTime(incident.openedAt)} />
							<KeyValue label="Resolved" value={formatDateTime(incident.resolvedAt)} />
							<KeyValue label="Last checked" value={formatDateTime(incident.lastEvaluatedAt)} />
							<KeyValue label="Last triggered" value={formatDateTime(incident.lastTriggeredAt)} />
						</div>
					</Panel>
					<Panel tone="matte" title="Notifications">
						<div className={styles.keyValueGrid}>
							<KeyValue label="Last sent" value={formatDateTime(incident.lastNotificationSentAt)} />
							<KeyValue label="Next eligible" value={formatDateTime(incident.nextNotificationEligibleAt)} />
							<KeyValue label="Suppressed" value={String(incident.suppressedNotificationCount)} />
						</div>
					</Panel>
				</div>
			) : (
				<LoadingState label={isLoading ? "Loading incident" : "Incident unavailable"} detail={error ? requestErrorMessage(error) : "Fetching incident detail for this project."} size="compact" />
			)}
		</EditorDrawer>
	);
}

function RuleEditorDrawer({
	editor,
	notifications,
	isPending,
	onClose,
	onSubmit
}: {
	editor: RuleEditorState;
	notifications: ApiNotification[];
	isPending: boolean;
	onClose: () => void;
	onSubmit: (form: RuleFormState) => Promise<void>;
}) {
	const [form, setForm] = useState<RuleFormState>(() => (editor.mode === "edit" && editor.rule ? ruleFormFromRule(editor.rule) : defaultRuleForm()));
	const metricSelectOptions = useMemo(() => metricOptionsForForm(form), [form]);
	const title = editor.mode === "edit" ? "Edit rule" : "Create rule";
	const checkTypeSupported = supportsAlertMetrics(form.checkType);
	const numberValidation = validateRuleNumbers(form);
	const numberError = ruleNumberError(numberValidation);

	function updateForm(patch: Partial<RuleFormState>) {
		setForm(current => ({ ...current, ...patch }));
	}

	function handleCheckTypeChange(nextCheckType: CheckType) {
		const nextMetric = metricOptions[nextCheckType].some(option => option.value === form.metric) ? form.metric : (metricForCheckType(nextCheckType) ?? form.metric);
		updateForm({ checkType: nextCheckType, metric: nextMetric });
	}

	function toggleNotification(notificationID: string, checked: boolean) {
		updateForm({
			selectedNotificationIds: checked ? [...form.selectedNotificationIds, notificationID] : form.selectedNotificationIds.filter(selectedNotificationID => selectedNotificationID !== notificationID)
		});
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (!checkTypeSupported || numberError) {
			return;
		}
		await onSubmit(form);
	}

	return (
		<EditorDrawer open title={title} ariaLabel={title} backLabel="back to rules" onClose={onClose}>
			<form className={styles.drawerForm} onSubmit={handleSubmit}>
				<Panel tone="matte" title="Target">
					<div className={styles.formGrid}>
						<TextField label="Name" value={form.name} onChange={event => updateForm({ name: event.currentTarget.value })} maxLength={128} required />
						<TextAreaField label="Description" value={form.description} onChange={event => updateForm({ description: event.currentTarget.value })} rows={3} />
						<div className={styles.twoColumns}>
							<SelectField label="Status" value={form.enabled} options={enabledOptions} onChange={event => updateForm({ enabled: event.currentTarget.value })} />
							<SelectField label="Check type" value={form.checkType} options={checkTypeOptions} onChange={event => handleCheckTypeChange(event.currentTarget.value as CheckType)} />
						</div>
						<div className={styles.twoColumns}>
							<TextField label="Probe ID" value={form.probeId} onChange={event => updateForm({ probeId: event.currentTarget.value })} placeholder="Optional probe UUID" />
							<TextField label="Check ID" value={form.checkId} onChange={event => updateForm({ checkId: event.currentTarget.value })} placeholder="Optional check UUID" />
						</div>
					</div>
				</Panel>
				<Panel tone="matte" title="Condition">
					{checkTypeSupported ? (
						<div className={styles.formGrid}>
							<SelectField label="Metric" value={form.metric} options={metricSelectOptions} onChange={event => updateForm({ metric: event.currentTarget.value as AlertMetric })} />
							<div className={styles.twoColumns}>
								<SelectField label="Operator" value={form.operator} options={operatorOptions} onChange={event => updateForm({ operator: event.currentTarget.value as AlertOperator })} />
								<TextField
									label="Threshold"
									value={form.threshold}
									onChange={event => updateForm({ threshold: event.currentTarget.value })}
									inputMode="decimal"
									error={numberValidation.threshold.error || undefined}
									required
								/>
							</div>
						</div>
					) : (
						<p className={styles.unsupportedNotice}>Traceroute alert rules are not available yet because the controller API only exposes alert metrics for ping and TCP checks.</p>
					)}
				</Panel>
				<Panel tone="matte" title="Notify">
					<div className={styles.formGrid}>
						<div className={styles.twoColumns}>
							<SelectField label="Severity" value={form.severity} options={severityOptions} onChange={event => updateForm({ severity: event.currentTarget.value as AlertSeverity })} />
							<div className={styles.notificationPicker} aria-label="Notification targets">
								<span className={styles.notificationPickerLabel}>Notifications</span>
								{notifications.length ? (
									notifications.map(notification => (
										<label className={styles.notificationPickerOption} key={notification.id}>
											<input type="checkbox" checked={form.selectedNotificationIds.includes(notification.id)} onChange={event => toggleNotification(notification.id, event.currentTarget.checked)} />
											<span>{notification.name}</span>
										</label>
									))
								) : (
									<span className={styles.notificationPickerEmpty}>No notifications configured</span>
								)}
							</div>
						</div>
						<details className={styles.advancedTiming}>
							<summary>Advanced timing</summary>
							<div className={styles.threeColumns}>
								<TextField
									label="Window seconds"
									value={form.windowSeconds}
									onChange={event => updateForm({ windowSeconds: event.currentTarget.value })}
									inputMode="numeric"
									min={60}
									max={86400}
									error={numberValidation.windowSeconds.error || undefined}
									required
								/>
								<TextField
									label="Min samples"
									value={form.minSamples}
									onChange={event => updateForm({ minSamples: event.currentTarget.value })}
									inputMode="numeric"
									min={1}
									max={10000}
									error={numberValidation.minSamples.error || undefined}
									required
								/>
								<TextField
									label="Cooldown"
									value={form.cooldownSeconds}
									onChange={event => updateForm({ cooldownSeconds: event.currentTarget.value })}
									inputMode="numeric"
									min={60}
									max={86400}
									error={numberValidation.cooldownSeconds.error || undefined}
									required
								/>
							</div>
						</details>
					</div>
				</Panel>
				{checkTypeSupported && !numberError ? <div className={styles.previewSentence}>{rulePreview(form, notifications, numberValidation)}</div> : null}
				<div className={styles.drawerActions}>
					<Button type="button" variant="ghost" disabled={isPending} onClick={onClose}>
						Cancel
					</Button>
					<Button type="submit" disabled={isPending || !form.name.trim() || !checkTypeSupported || Boolean(numberError)}>
						{editor.mode === "edit" ? "Save rule" : "Create rule"}
					</Button>
				</div>
			</form>
		</EditorDrawer>
	);
}

function NotificationEditorDrawer({
	editor,
	isPending,
	onClose,
	onSubmit
}: {
	editor: NotificationEditorState;
	isPending: boolean;
	onClose: () => void;
	onSubmit: (form: NotificationFormState) => Promise<void>;
}) {
	const isEditing = editor.mode === "edit";
	const [form, setForm] = useState<NotificationFormState>(() => (editor.mode === "edit" && editor.notification ? notificationFormFromNotification(editor.notification) : defaultNotificationForm()));
	const [step, setStep] = useState<NotificationEditorStep>(isEditing ? "detail" : "type");
	const selectedType = notificationTypeOption(form.type);
	const title = isEditing ? "Edit notification" : "Add notification";

	function updateForm(patch: Partial<NotificationFormState>) {
		setForm(current => ({ ...current, ...patch }));
	}

	function chooseType(type: NotificationType) {
		updateForm({ type });
		setStep("detail");
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		await onSubmit(form);
	}

	if (!isEditing && step === "type") {
		return (
			<EditorDrawer open title={title} ariaLabel={title} backLabel="back to notifications" onClose={onClose}>
				<div className={styles.notificationTypeGrid}>
					{notificationTypeOptions.map(option => (
						<button type="button" className={styles.notificationTypeOption} key={option.value} onClick={() => chooseType(option.value)}>
							<NotificationTypeIcon type={option.value} />
							<span className={styles.notificationTypeText}>
								<strong>{option.label}</strong>
								<span>{option.detail}</span>
							</span>
						</button>
					))}
				</div>
			</EditorDrawer>
		);
	}

	return (
		<EditorDrawer open title={title} ariaLabel={title} backLabel="back to notifications" onClose={onClose}>
			<form className={styles.drawerForm} onSubmit={handleSubmit}>
				<Panel tone="matte" title="Notification type">
					<div className={styles.notificationTypeSummary}>
						<NotificationTypeIcon type={selectedType.value} />
						<span className={styles.notificationTypeText}>
							<strong>{selectedType.label}</strong>
							<span>{selectedType.detail}</span>
						</span>
					</div>
				</Panel>
				<Panel tone="matte" title={`${notificationTypeLabel(form.type)} settings`}>
					<div className={styles.formGrid}>
						<TextField label="Name" value={form.name} onChange={event => updateForm({ name: event.currentTarget.value })} maxLength={128} required />
						{form.type === "telegram" ? (
							<>
								<TextField
									label="Bot token"
									value={form.botToken}
									onChange={event => updateForm({ botToken: event.currentTarget.value })}
									placeholder="123456:telegram-bot-token"
									autoComplete="off"
									required
								/>
								<TextField label="Chat ID" value={form.chatId} onChange={event => updateForm({ chatId: event.currentTarget.value })} placeholder="-1001234567890" required />
							</>
						) : form.type === "email" ? (
							<TextAreaField
								label="Recipients"
								value={form.emailTo}
								onChange={event => updateForm({ emailTo: event.currentTarget.value })}
								placeholder="ops@example.com, sre@example.com"
								rows={4}
								required
							/>
						) : (
							<TextField
								label={notificationWebhookURLLabel(form.type)}
								value={form.url}
								onChange={event => updateForm({ url: event.currentTarget.value })}
								inputMode="url"
								placeholder={notificationWebhookURLPlaceholder(form.type)}
								required
							/>
						)}
						<SelectField label="Status" value={form.enabled} options={enabledOptions} onChange={event => updateForm({ enabled: event.currentTarget.value })} />
					</div>
				</Panel>
				<div className={styles.drawerActions}>
					{!isEditing ? (
						<Button type="button" variant="ghost" disabled={isPending} onClick={() => setStep("type")}>
							Change type
						</Button>
					) : null}
					<Button type="button" variant="ghost" disabled={isPending} onClick={onClose}>
						Cancel
					</Button>
					<Button type="submit" disabled={isPending || !notificationFormReady(form)}>
						{editor.mode === "edit" ? "Save notification" : "Add notification"}
					</Button>
				</div>
			</form>
		</EditorDrawer>
	);
}

function KeyValue({ label, value }: { label: string; value: string }) {
	return (
		<div className={styles.keyValue}>
			<span>{label}</span>
			<strong>{value}</strong>
		</div>
	);
}
