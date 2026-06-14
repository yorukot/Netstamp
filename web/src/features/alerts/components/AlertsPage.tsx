import {
	useCreateProjectAlertRuleMutation,
	useCreateProjectNotificationChannelMutation,
	useDeleteProjectAlertRuleMutation,
	useDeleteProjectNotificationChannelMutation,
	useTestProjectNotificationChannelMutation,
	useUpdateProjectAlertRuleMutation,
	useUpdateProjectNotificationChannelMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type {
	ApiAlertIncident,
	ApiAlertRule,
	ApiNotificationChannel,
	CreateAlertRuleInput,
	CreateNotificationChannelInput,
	UpdateAlertRuleInput,
	UpdateNotificationChannelInput
} from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { LoadingState } from "@/shared/components/LoadingState";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DataTable, Panel, SelectField, TextAreaField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { DiscordLogo, TelegramLogo, WebhooksLogo } from "@phosphor-icons/react";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState, type FormEvent, type MouseEvent } from "react";
import styles from "./AlertsPage.module.css";

type AlertTab = "incidents" | "rules" | "channels";
type IncidentStatusFilter = "open" | "acknowledged" | "resolved" | "all";
type RuleStatusFilter = "all" | "enabled" | "disabled";
type CheckType = "ping" | "tcp";
type RuleCheckTypeFilter = "all" | CheckType;
type ChannelStatusFilter = "all" | "enabled" | "disabled";
type NotificationChannelType = CreateNotificationChannelInput["type"];
type ChannelTypeFilter = "all" | NotificationChannelType;
type AlertMetric = CreateAlertRuleInput["condition"]["metric"];
type AlertOperator = CreateAlertRuleInput["condition"]["operator"];
type AlertSeverity = CreateAlertRuleInput["severity"];

interface RuleEditorState {
	mode: "create" | "edit";
	rule?: ApiAlertRule;
}

interface ChannelEditorState {
	mode: "create" | "edit";
	channel?: ApiNotificationChannel;
}

type ChannelEditorStep = "type" | "detail";

interface RuleFormState {
	name: string;
	description: string;
	enabled: string;
	severity: AlertSeverity;
	checkType: CheckType;
	metric: AlertMetric;
	operator: AlertOperator;
	threshold: string;
	windowSeconds: string;
	minSamples: string;
	cooldownSeconds: string;
	selectedChannelId: string;
}

interface ChannelFormState {
	name: string;
	type: NotificationChannelType;
	url: string;
	botToken: string;
	chatId: string;
	enabled: string;
}

interface ChannelTypeOption {
	value: NotificationChannelType;
	label: string;
	detail: string;
}

const emptyRules: ApiAlertRule[] = [];
const emptyIncidents: ApiAlertIncident[] = [];
const emptyChannels: ApiNotificationChannel[] = [];

const alertTabs: Array<{ value: AlertTab; label: string }> = [
	{ value: "incidents", label: "Incidents" },
	{ value: "rules", label: "Rules" },
	{ value: "channels", label: "Channels" }
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
	{ value: "tcp", label: "TCP" }
];

const channelStatusOptions: Array<{ value: ChannelStatusFilter; label: string }> = [
	{ value: "all", label: "Any status" },
	{ value: "enabled", label: "Enabled" },
	{ value: "disabled", label: "Disabled" }
];

const channelFilterTypeOptions: Array<{ value: ChannelTypeFilter; label: string }> = [
	{ value: "all", label: "Any type" },
	{ value: "webhook", label: "Webhook" },
	{ value: "discord", label: "Discord" },
	{ value: "telegram", label: "Telegram" }
];

const checkTypeOptions: Array<{ value: CheckType; label: string }> = [
	{ value: "ping", label: "Ping" },
	{ value: "tcp", label: "TCP" }
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
	]
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

const webhookChannelTypeOption: ChannelTypeOption = { value: "webhook", label: "Webhook", detail: "Send raw alert JSON to any HTTPS endpoint." };

const channelTypeOptions: ChannelTypeOption[] = [
	webhookChannelTypeOption,
	{ value: "discord", label: "Discord", detail: "Post alert summaries to a Discord channel webhook." },
	{ value: "telegram", label: "Telegram", detail: "Send alert summaries through a Telegram bot." }
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

function channelConfigString(config: ApiNotificationChannel["config"], key: string) {
	if (config && typeof config === "object" && key in config) {
		const value = (config as Record<string, unknown>)[key];
		return typeof value === "string" ? value : "";
	}

	return "";
}

function channelURL(channel: ApiNotificationChannel) {
	return channelConfigString(channel.config, "url");
}

function channelChatID(channel: ApiNotificationChannel) {
	return channelConfigString(channel.config, "chatId");
}

function channelTypeLabel(channelType: string) {
	switch (channelType) {
		case "webhook":
			return "Webhook";
		case "discord":
			return "Discord";
		case "telegram":
			return "Telegram";
		default:
			return channelType;
	}
}

function channelTypeOption(channelType: NotificationChannelType) {
	return channelTypeOptions.find(option => option.value === channelType) ?? webhookChannelTypeOption;
}

function ChannelTypeIcon({ type }: { type: NotificationChannelType }) {
	switch (type) {
		case "discord":
			return <DiscordLogo className={styles.channelTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "telegram":
			return <TelegramLogo className={styles.channelTypeIcon} size={28} weight="bold" aria-hidden="true" />;
		case "webhook":
		default:
			return <WebhooksLogo className={styles.channelTypeIcon} size={28} weight="bold" aria-hidden="true" />;
	}
}

function channelDestination(channel: ApiNotificationChannel) {
	switch (channel.type) {
		case "discord":
		case "webhook":
			return channelURL(channel);
		case "telegram":
			return channelChatID(channel) ? `chat ${channelChatID(channel)}` : "-";
		default:
			return "-";
	}
}

function channelNameByID(channels: ApiNotificationChannel[], channelID: string) {
	return channels.find(channel => channel.id === channelID)?.name ?? shortID(channelID);
}

function notificationLabel(rule: ApiAlertRule, channels: ApiNotificationChannel[]) {
	if (!rule.notificationChannelIds.length) {
		return "No notification";
	}

	return rule.notificationChannelIds.map(channelID => channelNameByID(channels, channelID)).join(", ");
}

function rulesUsingChannel(rules: ApiAlertRule[], channelID: string) {
	return rules.filter(rule => rule.notificationChannelIds.includes(channelID));
}

function parseInteger(value: string, fallback: number) {
	const parsed = Number.parseInt(value, 10);

	return Number.isFinite(parsed) ? parsed : fallback;
}

function parseFloatValue(value: string, fallback: number) {
	const parsed = Number.parseFloat(value);

	return Number.isFinite(parsed) ? parsed : fallback;
}

function metricForCheckType(nextCheckType: CheckType) {
	return metricOptions[nextCheckType][0]?.value ?? "ping.loss_percent";
}

function editableCheckType(checkType: string): CheckType {
	return checkType === "tcp" ? "tcp" : "ping";
}

function defaultRuleForm(): RuleFormState {
	return {
		name: "",
		description: "",
		enabled: "true",
		severity: "critical",
		checkType: "ping",
		metric: "ping.loss_percent",
		operator: "gte",
		threshold: "10",
		windowSeconds: "300",
		minSamples: "3",
		cooldownSeconds: "900",
		selectedChannelId: ""
	};
}

function ruleFormFromRule(rule: ApiAlertRule): RuleFormState {
	const checkType = editableCheckType(rule.scope.checkType);
	return {
		name: rule.name,
		description: rule.description ?? "",
		enabled: String(rule.enabled),
		severity: rule.severity,
		checkType,
		metric: metricOptions[checkType].some(option => option.value === rule.condition.metric) ? rule.condition.metric : metricForCheckType(checkType),
		operator: rule.condition.operator,
		threshold: String(rule.condition.threshold),
		windowSeconds: String(rule.condition.windowSeconds),
		minSamples: String(rule.condition.minSamples),
		cooldownSeconds: String(rule.cooldownSeconds),
		selectedChannelId: rule.notificationChannelIds[0] ?? ""
	};
}

function defaultChannelForm(): ChannelFormState {
	return {
		name: "",
		type: "webhook",
		url: "",
		botToken: "",
		chatId: "",
		enabled: "true"
	};
}

function channelFormFromChannel(channel: ApiNotificationChannel): ChannelFormState {
	return {
		name: channel.name,
		type: channel.type,
		url: channelURL(channel),
		botToken: "",
		chatId: channelChatID(channel),
		enabled: String(channel.enabled)
	};
}

function rulePayload(form: RuleFormState): CreateAlertRuleInput | UpdateAlertRuleInput {
	const description = form.description.trim();

	return {
		name: form.name.trim(),
		description: description || undefined,
		enabled: form.enabled === "true",
		severity: form.severity,
		scope: { checkType: form.checkType },
		condition: {
			type: "metric_threshold",
			metric: form.metric,
			operator: form.operator,
			threshold: parseFloatValue(form.threshold, 0),
			windowSeconds: parseInteger(form.windowSeconds, 300),
			minSamples: parseInteger(form.minSamples, 3)
		},
		cooldownSeconds: parseInteger(form.cooldownSeconds, 900),
		notificationChannelIds: form.selectedChannelId ? [form.selectedChannelId] : []
	};
}

function channelPayload(form: ChannelFormState): CreateNotificationChannelInput | UpdateNotificationChannelInput {
	const config = form.type === "telegram" ? { botToken: form.botToken.trim(), chatId: form.chatId.trim() } : { url: form.url.trim() };

	return {
		name: form.name.trim(),
		type: form.type,
		enabled: form.enabled === "true",
		config
	};
}

function channelFormReady(form: ChannelFormState) {
	if (!form.name.trim()) {
		return false;
	}
	if (form.type === "telegram") {
		return Boolean(form.botToken.trim() && form.chatId.trim());
	}
	return Boolean(form.url.trim());
}

function rulePreview(form: RuleFormState, channels: ApiNotificationChannel[]) {
	const metric = metricLabel(form.metric).toLowerCase();
	const threshold = formatThreshold(form.metric, form.threshold || "0");
	const channel = form.selectedChannelId ? channelNameByID(channels, form.selectedChannelId) : "no channel";

	return `Create a ${form.severity} incident when ${metric} ${operatorPhrases[form.operator]} ${threshold} for ${formatDuration(parseInteger(form.windowSeconds, 300))}. Notify ${channel}, then wait ${formatDuration(parseInteger(form.cooldownSeconds, 900))} before repeating.`;
}

function stopTableAction(event: MouseEvent<HTMLButtonElement>) {
	event.stopPropagation();
}

function tableState(label: string, detail: string) {
	return <LoadingState label={label} detail={detail} size="compact" />;
}

export function AlertsPage() {
	const confirm = useConfirm();
	const { projectRef } = useCurrentProject();
	const createRuleMutation = useCreateProjectAlertRuleMutation(projectRef);
	const updateRuleMutation = useUpdateProjectAlertRuleMutation(projectRef);
	const deleteRuleMutation = useDeleteProjectAlertRuleMutation(projectRef);
	const createChannelMutation = useCreateProjectNotificationChannelMutation(projectRef);
	const updateChannelMutation = useUpdateProjectNotificationChannelMutation(projectRef);
	const deleteChannelMutation = useDeleteProjectNotificationChannelMutation(projectRef);
	const testChannelMutation = useTestProjectNotificationChannelMutation(projectRef);
	const [activeTab, setActiveTab] = useState<AlertTab>("incidents");
	const [incidentStatus, setIncidentStatus] = useState<IncidentStatusFilter>("open");
	const [ruleSearch, setRuleSearch] = useState("");
	const [ruleStatus, setRuleStatus] = useState<RuleStatusFilter>("all");
	const [ruleCheckType, setRuleCheckType] = useState<RuleCheckTypeFilter>("all");
	const [channelStatus, setChannelStatus] = useState<ChannelStatusFilter>("all");
	const [channelType, setChannelType] = useState<ChannelTypeFilter>("all");
	const [selectedIncident, setSelectedIncident] = useState<ApiAlertIncident | null>(null);
	const [ruleEditor, setRuleEditor] = useState<RuleEditorState | null>(null);
	const [channelEditor, setChannelEditor] = useState<ChannelEditorState | null>(null);
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
	const channelsQuery = useQuery({
		...projectQueries.notificationChannels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const rules = rulesQuery.data?.rules ?? emptyRules;
	const incidents = incidentsQuery.data?.incidents ?? emptyIncidents;
	const openIncidents = openIncidentsQuery.data?.incidents ?? emptyIncidents;
	const channels = channelsQuery.data?.channels ?? emptyChannels;
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
			return [rule.name, rule.description, rule.scope.checkType, formatAlertCondition(rule.condition), notificationLabel(rule, channels)].filter(Boolean).join(" ").toLowerCase().includes(search);
		});
	}, [channels, ruleCheckType, ruleSearch, ruleStatus, rules]);
	const filteredChannels = useMemo(
		() =>
			channels.filter(channel => {
				if (channelStatus === "enabled" && !channel.enabled) {
					return false;
				}
				if (channelStatus === "disabled" && channel.enabled) {
					return false;
				}
				if (channelType !== "all" && channel.type !== channelType) {
					return false;
				}
				return true;
			}),
		[channelStatus, channelType, channels]
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
			render: rule => <span className={styles.urlCell}>{notificationLabel(rule, channels)}</span>
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
	const channelColumns: DataColumn<ApiNotificationChannel>[] = [
		{
			key: "name",
			label: "Channel",
			render: channel => (
				<div className={styles.primaryCell}>
					<strong>{channel.name}</strong>
					<span>{channelTypeLabel(channel.type)}</span>
				</div>
			)
		},
		{
			key: "status",
			label: "Status",
			render: channel => <Badge tone={channel.enabled ? "success" : "muted"}>{channel.enabled ? "enabled" : "disabled"}</Badge>
		},
		{
			key: "url",
			label: "Destination",
			render: channel => (
				<span className={styles.urlCell} title={channelDestination(channel)}>
					{channelDestination(channel)}
				</span>
			)
		},
		{
			key: "usedBy",
			label: "Used by rules",
			render: channel => rulesUsingChannel(rules, channel.id).length
		},
		{
			key: "actions",
			label: "",
			render: channel => (
				<div className={styles.rowActions}>
					<Button
						variant="secondary"
						size="sm"
						disabled={testChannelMutation.isPending}
						onClick={event => {
							stopTableAction(event);
							void testChannel(channel);
						}}
					>
						Test
					</Button>
					<Button
						variant="danger"
						size="sm"
						disabled={deleteChannelMutation.isPending}
						onClick={event => {
							stopTableAction(event);
							void deleteChannel(channel);
						}}
					>
						Delete
					</Button>
				</div>
			)
		}
	];

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

	async function deleteChannel(channel: ApiNotificationChannel) {
		const usedBy = rulesUsingChannel(rules, channel.id).length;
		const accepted = await confirm({
			title: "Delete notification channel?",
			message: usedBy ? `"${channel.name}" is used by ${usedBy} rule${usedBy === 1 ? "" : "s"}. Remove it only after moving those rules to another channel.` : `This removes "${channel.name}".`,
			confirmLabel: "Delete channel",
			tone: "danger"
		});

		if (!accepted) {
			return;
		}

		try {
			await deleteChannelMutation.mutateAsync(channel.id);
			pushToast({ title: "Channel deleted", message: channel.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function testChannel(channel: ApiNotificationChannel) {
		try {
			const response = await testChannelMutation.mutateAsync(channel.id);
			if (response.result.delivered) {
				pushToast({ title: "Test delivered", message: channel.name, tone: "success" });
				return;
			}
			pushToast({ title: "Test failed", message: response.result.message || response.result.code || channel.name, tone: "critical" });
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
				<SummaryCard label="Channels" value={channels.length} tone={channels.length ? "accent" : "muted"} detail={channels.length ? "Ready to notify" : "No channel configured"} />
			</section>
			<nav className={styles.tabs} aria-label="Alert sections">
				{alertTabs.map(tab => (
					<button
						type="button"
						className={classNames(styles.tab, activeTab === tab.value && styles.tabActive)}
						aria-pressed={activeTab === tab.value}
						key={tab.value}
						onClick={() => setActiveTab(tab.value)}
					>
						{tab.label}
					</button>
				))}
			</nav>
			{activeTab === "incidents" ? (
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
						onRowClick={setSelectedIncident}
						selectedKey={selectedIncident?.id}
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
			{activeTab === "rules" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.panelActions}>
							<TextField label="Search" value={ruleSearch} onChange={event => setRuleSearch(event.currentTarget.value)} placeholder="loss, RTT, channel" />
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
			{activeTab === "channels" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.channelPanelActions}>
							<SelectField label="Status" value={channelStatus} options={channelStatusOptions} onChange={event => setChannelStatus(event.currentTarget.value as ChannelStatusFilter)} />
							<SelectField label="Type" value={channelType} options={channelFilterTypeOptions} onChange={event => setChannelType(event.currentTarget.value as ChannelTypeFilter)} />
						</div>
						<Button type="button" onClick={() => setChannelEditor({ mode: "create" })}>
							Add channel
						</Button>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={channelColumns}
						rows={filteredChannels}
						density="compact"
						minWidth="52rem"
						getRowKey={channel => channel.id}
						getRowAriaLabel={channel => `Edit notification channel ${channel.name}`}
						onRowClick={channel => setChannelEditor({ mode: "edit", channel })}
						selectedKey={channelEditor?.channel?.id}
						emptyLabel={
							channelsQuery.isLoading ? (
								tableState("Loading channels", "Fetching notification channels.")
							) : channels.length ? (
								"No notification channels match this view"
							) : (
								<EmptyAction label="No notification channels yet" action="Add channel" onClick={() => setChannelEditor({ mode: "create" })} />
							)
						}
					/>
				</Panel>
			) : null}
			{selectedIncident ? <IncidentDetailDrawer incident={selectedIncident} onClose={() => setSelectedIncident(null)} /> : null}
			{ruleEditor ? (
				<RuleEditorDrawer
					editor={ruleEditor}
					channels={channels}
					isPending={createRuleMutation.isPending || updateRuleMutation.isPending}
					onClose={() => setRuleEditor(null)}
					onSubmit={async form => {
						const body = rulePayload(form);
						try {
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
			{channelEditor ? (
				<ChannelEditorDrawer
					editor={channelEditor}
					isPending={createChannelMutation.isPending || updateChannelMutation.isPending}
					onClose={() => setChannelEditor(null)}
					onSubmit={async form => {
						try {
							if (channelEditor.mode === "edit" && channelEditor.channel) {
								const body = channelPayload({ ...form, type: channelEditor.channel.type });
								await updateChannelMutation.mutateAsync({ channelId: channelEditor.channel.id, body });
								pushToast({ title: "Channel updated", message: body.name, tone: "success" });
							} else {
								const body = channelPayload(form);
								await createChannelMutation.mutateAsync(body);
								pushToast({ title: "Channel created", message: body.name, tone: "success" });
							}
							setChannelEditor(null);
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

function IncidentDetailDrawer({ incident, onClose }: { incident: ApiAlertIncident; onClose: () => void }) {
	return (
		<EditorDrawer open title="Incident detail" ariaLabel="Incident detail" backLabel="back to incidents" onClose={onClose}>
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
		</EditorDrawer>
	);
}

function RuleEditorDrawer({
	editor,
	channels,
	isPending,
	onClose,
	onSubmit
}: {
	editor: RuleEditorState;
	channels: ApiNotificationChannel[];
	isPending: boolean;
	onClose: () => void;
	onSubmit: (form: RuleFormState) => Promise<void>;
}) {
	const [form, setForm] = useState<RuleFormState>(() => (editor.mode === "edit" && editor.rule ? ruleFormFromRule(editor.rule) : defaultRuleForm()));
	const channelOptions = useMemo(() => [{ value: "", label: "No notification" }, ...channels.map(channel => ({ value: channel.id, label: channel.name }))], [channels]);
	const title = editor.mode === "edit" ? "Edit rule" : "Create rule";

	function updateForm(patch: Partial<RuleFormState>) {
		setForm(current => ({ ...current, ...patch }));
	}

	function handleCheckTypeChange(nextCheckType: CheckType) {
		updateForm({ checkType: nextCheckType, metric: metricForCheckType(nextCheckType) });
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
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
					</div>
				</Panel>
				<Panel tone="matte" title="Condition">
					<div className={styles.formGrid}>
						<SelectField label="Metric" value={form.metric} options={metricOptions[form.checkType]} onChange={event => updateForm({ metric: event.currentTarget.value as AlertMetric })} />
						<div className={styles.twoColumns}>
							<SelectField label="Operator" value={form.operator} options={operatorOptions} onChange={event => updateForm({ operator: event.currentTarget.value as AlertOperator })} />
							<TextField label="Threshold" value={form.threshold} onChange={event => updateForm({ threshold: event.currentTarget.value })} inputMode="decimal" required />
						</div>
					</div>
				</Panel>
				<Panel tone="matte" title="Notify">
					<div className={styles.formGrid}>
						<div className={styles.twoColumns}>
							<SelectField label="Severity" value={form.severity} options={severityOptions} onChange={event => updateForm({ severity: event.currentTarget.value as AlertSeverity })} />
							<SelectField label="Channel" value={form.selectedChannelId} options={channelOptions} onChange={event => updateForm({ selectedChannelId: event.currentTarget.value })} />
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
									required
								/>
								<TextField label="Min samples" value={form.minSamples} onChange={event => updateForm({ minSamples: event.currentTarget.value })} inputMode="numeric" min={1} max={10000} required />
								<TextField
									label="Cooldown"
									value={form.cooldownSeconds}
									onChange={event => updateForm({ cooldownSeconds: event.currentTarget.value })}
									inputMode="numeric"
									min={60}
									max={86400}
									required
								/>
							</div>
						</details>
					</div>
				</Panel>
				<div className={classNames("ns-cut-frame", styles.previewSentence)}>{rulePreview(form, channels)}</div>
				<div className={styles.drawerActions}>
					<Button type="button" variant="ghost" disabled={isPending} onClick={onClose}>
						Cancel
					</Button>
					<Button type="submit" disabled={isPending || !form.name.trim()}>
						{editor.mode === "edit" ? "Save rule" : "Create rule"}
					</Button>
				</div>
			</form>
		</EditorDrawer>
	);
}

function ChannelEditorDrawer({ editor, isPending, onClose, onSubmit }: { editor: ChannelEditorState; isPending: boolean; onClose: () => void; onSubmit: (form: ChannelFormState) => Promise<void> }) {
	const isEditing = editor.mode === "edit";
	const [form, setForm] = useState<ChannelFormState>(() => (editor.mode === "edit" && editor.channel ? channelFormFromChannel(editor.channel) : defaultChannelForm()));
	const [step, setStep] = useState<ChannelEditorStep>(isEditing ? "detail" : "type");
	const selectedType = channelTypeOption(form.type);
	const title = isEditing ? "Edit channel" : "Add channel";

	function updateForm(patch: Partial<ChannelFormState>) {
		setForm(current => ({ ...current, ...patch }));
	}

	function chooseType(type: NotificationChannelType) {
		updateForm({ type });
		setStep("detail");
	}

	async function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		await onSubmit(form);
	}

	if (!isEditing && step === "type") {
		return (
			<EditorDrawer open title={title} ariaLabel={title} backLabel="back to channels" onClose={onClose}>
				<div className={styles.channelTypeGrid}>
					{channelTypeOptions.map(option => (
						<button type="button" className={classNames("ns-cut-frame", styles.channelTypeOption)} key={option.value} onClick={() => chooseType(option.value)}>
							<ChannelTypeIcon type={option.value} />
							<span className={styles.channelTypeText}>
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
		<EditorDrawer open title={title} ariaLabel={title} backLabel="back to channels" onClose={onClose}>
			<form className={styles.drawerForm} onSubmit={handleSubmit}>
				<Panel tone="matte" title="Channel type">
					<div className={classNames("ns-cut-frame", styles.channelTypeSummary)}>
						<ChannelTypeIcon type={selectedType.value} />
						<span className={styles.channelTypeText}>
							<strong>{selectedType.label}</strong>
							<span>{selectedType.detail}</span>
						</span>
					</div>
				</Panel>
				<Panel tone="matte" title={`${channelTypeLabel(form.type)} settings`}>
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
						) : (
							<TextField
								label={form.type === "discord" ? "Discord webhook URL" : "Webhook URL"}
								value={form.url}
								onChange={event => updateForm({ url: event.currentTarget.value })}
								inputMode="url"
								placeholder={form.type === "discord" ? "https://discord.com/api/webhooks/..." : "https://hooks.example.com/netstamp"}
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
					<Button type="submit" disabled={isPending || !channelFormReady(form)}>
						{editor.mode === "edit" ? "Save channel" : "Add channel"}
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
