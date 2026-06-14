import { useCreateProjectAlertRuleMutation, useCreateProjectNotificationChannelMutation, useDeleteProjectAlertRuleMutation, useDeleteProjectNotificationChannelMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiAlertIncident, ApiAlertRule, ApiNotificationChannel, CreateAlertRuleInput, CreateNotificationChannelInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DataTable, Panel, SelectField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState, type FormEvent } from "react";
import styles from "./AlertsPage.module.css";

type CheckType = "ping" | "tcp";
type AlertMetric = CreateAlertRuleInput["condition"]["metric"];
type AlertOperator = CreateAlertRuleInput["condition"]["operator"];
type AlertSeverity = CreateAlertRuleInput["severity"];

const emptyRules: ApiAlertRule[] = [];
const emptyIncidents: ApiAlertIncident[] = [];
const emptyChannels: ApiNotificationChannel[] = [];

const metricOptions: Record<CheckType, Array<{ value: AlertMetric; label: string }>> = {
	ping: [
		{ value: "ping.loss_percent", label: "Ping loss percent" },
		{ value: "ping.average_rtt_ms", label: "Ping average RTT" },
		{ value: "ping.max_rtt_ms", label: "Ping max RTT" },
		{ value: "ping.success_rate", label: "Ping success rate" }
	],
	tcp: [
		{ value: "tcp.failure_percent", label: "TCP failure percent" },
		{ value: "tcp.average_connect_ms", label: "TCP average connect" },
		{ value: "tcp.max_connect_ms", label: "TCP max connect" },
		{ value: "tcp.success_rate", label: "TCP success rate" }
	]
};

const severityOptions = [
	{ value: "info", label: "Info" },
	{ value: "warning", label: "Warning" },
	{ value: "critical", label: "Critical" }
];

const operatorOptions = [
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

function ruleCondition(rule: ApiAlertRule) {
	const metric = rule.condition.metric.replace(/_/g, " ");
	return `${metric} ${rule.condition.operator} ${rule.condition.threshold}`;
}

function channelURL(channel: ApiNotificationChannel) {
	return channel.config.url;
}

export function AlertsPage() {
	const { projectRef } = useCurrentProject();
	const createRuleMutation = useCreateProjectAlertRuleMutation(projectRef);
	const deleteRuleMutation = useDeleteProjectAlertRuleMutation(projectRef);
	const createChannelMutation = useCreateProjectNotificationChannelMutation(projectRef);
	const deleteChannelMutation = useDeleteProjectNotificationChannelMutation(projectRef);
	const rulesQuery = useQuery({
		...projectQueries.alertRules(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const incidentsQuery = useQuery({
		...projectQueries.alertIncidents(projectRef || "", { limit: 100 }),
		enabled: Boolean(projectRef),
		refetchInterval: 30 * 1000
	});
	const channelsQuery = useQuery({
		...projectQueries.notificationChannels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const rules = rulesQuery.data?.rules ?? emptyRules;
	const incidents = incidentsQuery.data?.incidents ?? emptyIncidents;
	const channels = channelsQuery.data?.channels ?? emptyChannels;
	const [channelName, setChannelName] = useState("");
	const [webhookURL, setWebhookURL] = useState("");
	const [channelEnabled, setChannelEnabled] = useState("true");
	const [ruleName, setRuleName] = useState("");
	const [ruleEnabled, setRuleEnabled] = useState("true");
	const [ruleSeverity, setRuleSeverity] = useState<AlertSeverity>("critical");
	const [checkType, setCheckType] = useState<CheckType>("ping");
	const [metric, setMetric] = useState<AlertMetric>("ping.loss_percent");
	const [operator, setOperator] = useState<AlertOperator>("gte");
	const [threshold, setThreshold] = useState("10");
	const [windowSeconds, setWindowSeconds] = useState("300");
	const [minSamples, setMinSamples] = useState("3");
	const [cooldownSeconds, setCooldownSeconds] = useState("900");
	const [selectedChannelId, setSelectedChannelId] = useState("");
	const activeIncidents = incidents.filter(incident => incident.status !== "resolved").length;
	const channelOptions = useMemo(() => [{ value: "", label: "No notification" }, ...channels.map(channel => ({ value: channel.id, label: channel.name }))], [channels]);
	const ruleColumns: DataColumn<ApiAlertRule>[] = [
		{
			key: "name",
			label: "Rule",
			render: rule => (
				<div className={styles.primaryCell}>
					<strong>{rule.name}</strong>
					<span>{rule.scope.checkType}</span>
				</div>
			)
		},
		{
			key: "severity",
			label: "Severity",
			render: rule => <Badge tone={severityTone(rule.severity)}>{rule.severity}</Badge>
		},
		{
			key: "condition",
			label: "Condition",
			render: rule => <span className={styles.monoCell}>{ruleCondition(rule)}</span>
		},
		{
			key: "status",
			label: "Status",
			render: rule => <Badge tone={rule.enabled ? "success" : "muted"}>{rule.enabled ? "enabled" : "disabled"}</Badge>
		},
		{
			key: "channels",
			label: "Channels",
			render: rule => rule.notificationChannelIds.length
		},
		{
			key: "updatedAt",
			label: "Updated",
			render: rule => formatDateTime(rule.updatedAt)
		},
		{
			key: "delete",
			label: "Delete",
			render: rule => (
				<Button variant="danger" size="sm" disabled={deleteRuleMutation.isPending} onClick={() => void deleteRule(rule.id)}>
					Delete
				</Button>
			)
		}
	];
	const incidentColumns: DataColumn<ApiAlertIncident>[] = [
		{
			key: "status",
			label: "Status",
			render: incident => <Badge tone={incidentTone(incident.status)}>{incident.status}</Badge>
		},
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
					<strong>{incident.checkType}</strong>
					<span>
						{shortID(incident.probeId)} / {shortID(incident.checkId)}
					</span>
				</div>
			)
		},
		{
			key: "state",
			label: "State",
			render: incident => incident.lastEvaluationState
		},
		{
			key: "value",
			label: "Value",
			render: incident => (typeof incident.lastValue === "number" ? incident.lastValue.toFixed(2) : "-")
		},
		{
			key: "openedAt",
			label: "Opened",
			render: incident => formatDateTime(incident.openedAt)
		}
	];
	const channelColumns: DataColumn<ApiNotificationChannel>[] = [
		{
			key: "name",
			label: "Channel",
			render: channel => (
				<div className={styles.primaryCell}>
					<strong>{channel.name}</strong>
					<span>{channel.type}</span>
				</div>
			)
		},
		{
			key: "url",
			label: "Webhook",
			render: channel => (
				<span className={styles.urlCell} title={channelURL(channel)}>
					{channelURL(channel)}
				</span>
			)
		},
		{
			key: "status",
			label: "Status",
			render: channel => <Badge tone={channel.enabled ? "success" : "muted"}>{channel.enabled ? "enabled" : "disabled"}</Badge>
		},
		{
			key: "delete",
			label: "Delete",
			render: channel => (
				<Button variant="danger" size="sm" disabled={deleteChannelMutation.isPending} onClick={() => void deleteChannel(channel.id)}>
					Delete
				</Button>
			)
		}
	];

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

	function handleCheckTypeChange(nextCheckType: CheckType) {
		setCheckType(nextCheckType);
		setMetric(metricForCheckType(nextCheckType));
	}

	async function createChannel(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		const body: CreateNotificationChannelInput = {
			name: channelName.trim(),
			type: "webhook",
			enabled: channelEnabled === "true",
			config: { url: webhookURL.trim() }
		};

		try {
			await createChannelMutation.mutateAsync(body);
			setChannelName("");
			setWebhookURL("");
			pushToast({ title: "Channel created", message: body.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function deleteChannel(channelId: string) {
		try {
			await deleteChannelMutation.mutateAsync(channelId);
			pushToast({ title: "Channel deleted", message: shortID(channelId), tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function createRule(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();

		const body: CreateAlertRuleInput = {
			name: ruleName.trim(),
			enabled: ruleEnabled === "true",
			severity: ruleSeverity,
			scope: { checkType },
			condition: {
				type: "metric_threshold",
				metric,
				operator,
				threshold: parseFloatValue(threshold, 0),
				windowSeconds: parseInteger(windowSeconds, 300),
				minSamples: parseInteger(minSamples, 3)
			},
			cooldownSeconds: parseInteger(cooldownSeconds, 900),
			notificationChannelIds: selectedChannelId ? [selectedChannelId] : []
		};

		try {
			await createRuleMutation.mutateAsync(body);
			setRuleName("");
			pushToast({ title: "Rule created", message: body.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function deleteRule(ruleId: string) {
		try {
			await deleteRuleMutation.mutateAsync(ruleId);
			pushToast({ title: "Rule deleted", message: shortID(ruleId), tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	return (
		<PageStack>
			<ScreenHeader title="Alerts" />
			<section className={styles.summaryGrid} aria-label="Alert summary">
				<Panel tone="matte" title="Rules">
					<strong className={styles.summaryValue}>{rules.length}</strong>
				</Panel>
				<Panel tone="matte" title="Active incidents">
					<strong className={styles.summaryValue}>{activeIncidents}</strong>
				</Panel>
				<Panel tone="matte" title="Channels">
					<strong className={styles.summaryValue}>{channels.length}</strong>
				</Panel>
			</section>
			<section className={styles.grid}>
				<div className={styles.tableStack}>
					<Panel title="Alert rules" padded={false}>
						<DataTable columns={ruleColumns} rows={rules} density="compact" minWidth="56rem" getRowKey={rule => rule.id} emptyLabel={rulesQuery.isLoading ? "Loading alert rules" : "No alert rules"} />
					</Panel>
					<Panel title="Incidents" padded={false}>
						<DataTable
							columns={incidentColumns}
							rows={incidents}
							density="compact"
							minWidth="52rem"
							getRowKey={incident => incident.id}
							emptyLabel={incidentsQuery.isLoading ? "Loading incidents" : "No incidents"}
						/>
					</Panel>
					<Panel title="Notification channels" padded={false}>
						<DataTable
							columns={channelColumns}
							rows={channels}
							density="compact"
							minWidth="44rem"
							getRowKey={channel => channel.id}
							emptyLabel={channelsQuery.isLoading ? "Loading channels" : "No notification channels"}
						/>
					</Panel>
				</div>
				<div className={styles.formStack}>
					<Panel title="New webhook channel">
						<form className={styles.form} onSubmit={createChannel}>
							<TextField label="Name" value={channelName} onChange={event => setChannelName(event.currentTarget.value)} maxLength={128} required />
							<TextField
								label="Webhook URL"
								value={webhookURL}
								onChange={event => setWebhookURL(event.currentTarget.value)}
								inputMode="url"
								placeholder="https://hooks.example.com/netstamp"
								required
							/>
							<SelectField label="Status" value={channelEnabled} options={enabledOptions} onChange={event => setChannelEnabled(event.currentTarget.value)} />
							<Button type="submit" disabled={!projectRef || createChannelMutation.isPending || !channelName.trim() || !webhookURL.trim()}>
								Create channel
							</Button>
						</form>
					</Panel>
					<Panel title="New alert rule">
						<form className={styles.form} onSubmit={createRule}>
							<TextField label="Name" value={ruleName} onChange={event => setRuleName(event.currentTarget.value)} maxLength={128} required />
							<div className={styles.twoColumns}>
								<SelectField label="Status" value={ruleEnabled} options={enabledOptions} onChange={event => setRuleEnabled(event.currentTarget.value)} />
								<SelectField label="Severity" value={ruleSeverity} options={severityOptions} onChange={event => setRuleSeverity(event.currentTarget.value as AlertSeverity)} />
							</div>
							<div className={styles.twoColumns}>
								<SelectField
									label="Check type"
									value={checkType}
									options={[
										{ value: "ping", label: "Ping" },
										{ value: "tcp", label: "TCP" }
									]}
									onChange={event => handleCheckTypeChange(event.currentTarget.value as CheckType)}
								/>
								<SelectField label="Metric" value={metric} options={metricOptions[checkType]} onChange={event => setMetric(event.currentTarget.value as AlertMetric)} />
							</div>
							<div className={styles.twoColumns}>
								<SelectField label="Operator" value={operator} options={operatorOptions} onChange={event => setOperator(event.currentTarget.value as AlertOperator)} />
								<TextField label="Threshold" value={threshold} onChange={event => setThreshold(event.currentTarget.value)} inputMode="decimal" required />
							</div>
							<div className={styles.threeColumns}>
								<TextField label="Window seconds" value={windowSeconds} onChange={event => setWindowSeconds(event.currentTarget.value)} inputMode="numeric" min={60} max={86400} required />
								<TextField label="Min samples" value={minSamples} onChange={event => setMinSamples(event.currentTarget.value)} inputMode="numeric" min={1} max={10000} required />
								<TextField label="Cooldown" value={cooldownSeconds} onChange={event => setCooldownSeconds(event.currentTarget.value)} inputMode="numeric" min={60} max={86400} required />
							</div>
							<SelectField label="Channel" value={selectedChannelId} options={channelOptions} onChange={event => setSelectedChannelId(event.currentTarget.value)} />
							<Button type="submit" disabled={!projectRef || createRuleMutation.isPending || !ruleName.trim()}>
								Create rule
							</Button>
						</form>
					</Panel>
				</div>
			</section>
		</PageStack>
	);
}
