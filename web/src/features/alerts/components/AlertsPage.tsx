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
import type { ApiAlertIncident, ApiAlertRule, ApiNotification } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, BodyCopy, Button, DataTable, EmptyState, KeyValueRow, Panel, SelectableRow, SelectField, Spinner, Tabs, TextAreaField, TextField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { DiscordLogoIcon } from "@phosphor-icons/react/dist/csr/DiscordLogo";
import { EnvelopeSimpleIcon } from "@phosphor-icons/react/dist/csr/EnvelopeSimple";
import { SlackLogoIcon } from "@phosphor-icons/react/dist/csr/SlackLogo";
import { TelegramLogoIcon } from "@phosphor-icons/react/dist/csr/TelegramLogo";
import { WebhooksLogoIcon } from "@phosphor-icons/react/dist/csr/WebhooksLogo";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState, type FormEvent } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import {
	alertTabs,
	checkTypeOptions,
	defaultNotificationForm,
	defaultRuleForm,
	emptyIncidents,
	emptyNotifications,
	emptyRules,
	enabledOptions,
	formatAlertCondition,
	formatDateTime,
	formatDuration,
	formatIncidentCheck,
	formatIncidentProbe,
	formatIncidentReason,
	formatRuleScope,
	formatThreshold,
	incidentCheckName,
	incidentCheckTarget,
	incidentProbeName,
	incidentStatusOptions,
	incidentTargetTitle,
	incidentTone,
	metricForCheckType,
	metricOptions,
	metricOptionsForForm,
	notificationDestination,
	notificationFilterTypeOptions,
	notificationFormFromNotification,
	notificationFormReady,
	notificationLabel,
	notificationPayload,
	notificationStatusOptions,
	notificationTypeLabel,
	notificationTypeOption,
	notificationTypeOptions,
	notificationWebhookURLLabel,
	notificationWebhookURLPlaceholder,
	operatorOptions,
	optionValue,
	pathWithSearch,
	ruleCheckTypeOptions,
	ruleFormFromRule,
	ruleNumberError,
	rulePayload,
	rulePreview,
	ruleStatusOptions,
	rulesUsingNotification,
	severityOptions,
	severityTone,
	shortID,
	stopTableAction,
	supportsAlertMetrics,
	validateRuleNumbers,
	type AlertMetric,
	type AlertOperator,
	type AlertSeverity,
	type AlertTab,
	type CheckType,
	type NotificationEditorState,
	type NotificationEditorStep,
	type NotificationFormState,
	type NotificationType,
	type RuleEditorState,
	type RuleFormState
} from "./alertPageModel";
import styles from "./AlertsPage.module.css";

function NotificationTypeIcon({ type }: { type: NotificationType }) {
	switch (type) {
		case "slack":
			return <SlackLogoIcon className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" focusable="false" />;
		case "discord":
			return <DiscordLogoIcon className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" focusable="false" />;
		case "telegram":
			return <TelegramLogoIcon className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" focusable="false" />;
		case "email":
			return <EnvelopeSimpleIcon className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" focusable="false" />;
		case "webhook":
		default:
			return <WebhooksLogoIcon className={styles.notificationTypeIcon} size={28} weight="bold" aria-hidden="true" focusable="false" />;
	}
}

function tableSpinner(label: string) {
	return <Spinner label={label} layout="compact" size="lg" />;
}

function sameValue(left: unknown, right: unknown) {
	return JSON.stringify(left) === JSON.stringify(right);
}

export function AlertsPage() {
	const confirm = useConfirm();
	const navigate = useNavigate();
	const { incidentId = "" } = useParams();
	const [searchParams, setSearchParams] = useSearchParams();
	const { projectRef } = useCurrentProject();
	const createRuleMutation = useCreateProjectAlertRuleMutation(projectRef, { suppressGlobalErrorToast: true });
	const updateRuleMutation = useUpdateProjectAlertRuleMutation(projectRef, { suppressGlobalErrorToast: true });
	const deleteRuleMutation = useDeleteProjectAlertRuleMutation(projectRef, { suppressGlobalErrorToast: true });
	const createNotificationMutation = useCreateProjectNotificationMutation(projectRef, { suppressGlobalErrorToast: true });
	const updateNotificationMutation = useUpdateProjectNotificationMutation(projectRef, { suppressGlobalErrorToast: true });
	const deleteNotificationMutation = useDeleteProjectNotificationMutation(projectRef, { suppressGlobalErrorToast: true });
	const testNotificationMutation = useTestProjectNotificationMutation(projectRef, { suppressGlobalErrorToast: true });
	const searchParamString = searchParams.toString();
	const activeTab = optionValue(searchParams.get("tab"), alertTabs, "incidents");
	const incidentStatus = optionValue(searchParams.get("incidentStatus"), incidentStatusOptions, "open");
	const ruleSearch = searchParams.get("ruleSearch") ?? "";
	const ruleStatus = optionValue(searchParams.get("ruleStatus"), ruleStatusOptions, "all");
	const ruleCheckType = optionValue(searchParams.get("ruleType"), ruleCheckTypeOptions, "all");
	const notificationStatus = optionValue(searchParams.get("notificationStatus"), notificationStatusOptions, "all");
	const notificationType = optionValue(searchParams.get("notificationType"), notificationFilterTypeOptions, "all");
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
		navigate(pathWithSearch(pathForAlertIncidentDetail(projectRef, incident.id), searchParamString));
	}

	function closeIncidentDetail() {
		navigate(pathWithSearch(pathForRoute("alerts", { projectRef }), searchParamString));
	}

	function updateAlertSearchParam(key: string, value: string, fallback: string) {
		const next = new URLSearchParams(searchParamString);

		if (value === fallback || !value.trim()) {
			next.delete(key);
		} else {
			next.set(key, value);
		}

		setSearchParams(next, { replace: true });
	}

	function changeAlertTab(value: AlertTab) {
		const next = new URLSearchParams(searchParamString);

		if (value === "incidents") {
			next.delete("tab");
		} else {
			next.set("tab", value);
		}

		if (incidentId) {
			navigate(pathWithSearch(pathForRoute("alerts", { projectRef }), next.toString()));
			return;
		}

		setSearchParams(next);
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
			<Tabs tabs={alertTabs} value={visibleTab} ariaLabel="Alert sections" onValueChange={value => changeAlertTab(value as AlertTab)} />
			{visibleTab === "incidents" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.singlePanelAction}>
							<SelectField label="Status" value={incidentStatus} options={incidentStatusOptions} onChange={event => updateAlertSearchParam("incidentStatus", event.currentTarget.value, "open")} />
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
						emptyLabel={incidentsQuery.isLoading ? tableSpinner("Loading incidents") : incidentStatus === "open" ? "No open incidents" : "No incidents match this view"}
					/>
				</Panel>
			) : null}
			{visibleTab === "rules" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.panelActions}>
							<TextField label="Search" value={ruleSearch} onChange={event => updateAlertSearchParam("ruleSearch", event.currentTarget.value, "")} placeholder="loss, RTT, notification" />
							<SelectField label="Status" value={ruleStatus} options={ruleStatusOptions} onChange={event => updateAlertSearchParam("ruleStatus", event.currentTarget.value, "all")} />
							<SelectField label="Type" value={ruleCheckType} options={ruleCheckTypeOptions} onChange={event => updateAlertSearchParam("ruleType", event.currentTarget.value, "all")} />
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
								tableSpinner("Loading alert rules")
							) : rules.length ? (
								"No alert rules match this view"
							) : (
								<EmptyState
									title="No alert rules yet"
									action={
										<Button type="button" size="sm" variant="secondary" onClick={() => setRuleEditor({ mode: "create" })}>
											Create rule
										</Button>
									}
								/>
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
								onChange={event => updateAlertSearchParam("notificationStatus", event.currentTarget.value, "all")}
							/>
							<SelectField
								label="Type"
								value={notificationType}
								options={notificationFilterTypeOptions}
								onChange={event => updateAlertSearchParam("notificationType", event.currentTarget.value, "all")}
							/>
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
								tableSpinner("Loading notifications")
							) : notifications.length ? (
								"No notifications match this view"
							) : (
								<EmptyState
									title="No notifications yet"
									action={
										<Button type="button" size="sm" variant="secondary" onClick={() => setNotificationEditor({ mode: "create" })}>
											Add notification
										</Button>
									}
								/>
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

function IncidentDetailDrawer({ incident, isLoading, error, onClose }: { incident: ApiAlertIncident | null; isLoading: boolean; error: unknown; onClose: () => void }) {
	return (
		<EditorDrawer open title="Incident detail" ariaLabel="Incident detail" onClose={onClose}>
			{incident ? (
				<div className={styles.detailStack}>
					<div className={styles.detailHeader}>
						<Badge tone={incidentTone(incident.status)}>{incident.status}</Badge>
						<Badge tone={severityTone(incident.severity)}>{incident.severity}</Badge>
					</div>
					<Panel tone="matte" title="What happened">
						<p className={styles.detailLead}>{formatIncidentReason(incident)}</p>
						<div className={styles.keyValueGrid}>
							<KeyValueRow label="Probe" value={formatIncidentProbe(incident)} />
							<KeyValueRow label="Check" value={formatIncidentCheck(incident)} />
							<KeyValueRow label="Target" value={incidentCheckTarget(incident)} />
							<KeyValueRow label="State" value={incident.lastEvaluationState} />
							<KeyValueRow label="Value" value={typeof incident.lastValue === "number" ? formatThreshold(incident.lastSummary.metric, Number(incident.lastValue.toFixed(2))) : "-"} />
							<KeyValueRow label="Rule" value={shortID(incident.ruleId)} />
						</div>
					</Panel>
					<Panel tone="matte" title="Timeline">
						<div className={styles.keyValueGrid}>
							<KeyValueRow label="Opened" value={formatDateTime(incident.openedAt)} />
							<KeyValueRow label="Resolved" value={formatDateTime(incident.resolvedAt)} />
							<KeyValueRow label="Last checked" value={formatDateTime(incident.lastEvaluatedAt)} />
							<KeyValueRow label="Last triggered" value={formatDateTime(incident.lastTriggeredAt)} />
						</div>
					</Panel>
					<Panel tone="matte" title="Notifications">
						<div className={styles.keyValueGrid}>
							<KeyValueRow label="Last sent" value={formatDateTime(incident.lastNotificationSentAt)} />
							<KeyValueRow label="Next eligible" value={formatDateTime(incident.nextNotificationEligibleAt)} />
							<KeyValueRow label="Suppressed" value={String(incident.suppressedNotificationCount)} />
						</div>
					</Panel>
				</div>
			) : isLoading ? (
				<Spinner label="Loading incident" layout="compact" size="lg" />
			) : (
				<BodyCopy>{error ? requestErrorMessage(error) : "Incident unavailable."}</BodyCopy>
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
	const initialForm = useMemo(() => (editor.mode === "edit" && editor.rule ? ruleFormFromRule(editor.rule) : defaultRuleForm()), [editor.mode, editor.rule]);
	const [form, setForm] = useState<RuleFormState>(() => initialForm);
	const metricSelectOptions = useMemo(() => metricOptionsForForm(form), [form]);
	const title = editor.mode === "edit" ? "Edit rule" : "Create rule";
	const checkTypeSupported = supportsAlertMetrics(form.checkType);
	const numberValidation = validateRuleNumbers(form);
	const numberError = ruleNumberError(numberValidation);
	const hasRuleChanges = !sameValue(form, initialForm);

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
		<EditorDrawer open title={title} ariaLabel={title} onClose={onClose}>
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
						<p className={styles.unsupportedNotice}>Traceroute alert rules are not available yet because the controller API exposes alert metrics for ping, TCP, and HTTP checks.</p>
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
				</div>
				<UnsavedChangesBar
					show={hasRuleChanges}
					saveType="submit"
					saving={isPending}
					disabled={!form.name.trim() || !checkTypeSupported || Boolean(numberError)}
					onReset={() => setForm(initialForm)}
				/>
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
	const initialForm = useMemo(
		() => (editor.mode === "edit" && editor.notification ? notificationFormFromNotification(editor.notification) : defaultNotificationForm()),
		[editor.mode, editor.notification]
	);
	const [form, setForm] = useState<NotificationFormState>(() => initialForm);
	const [step, setStep] = useState<NotificationEditorStep>(isEditing ? "detail" : "type");
	const selectedType = notificationTypeOption(form.type);
	const title = isEditing ? "Edit notification" : "Add notification";
	const hasNotificationChanges = !sameValue(form, initialForm);

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
			<EditorDrawer open title={title} ariaLabel={title} onClose={onClose}>
				<div className={styles.notificationTypeGrid}>
					{notificationTypeOptions.map(option => (
						<SelectableRow
							key={option.value}
							type="button"
							leading={<NotificationTypeIcon type={option.value} />}
							title={option.label}
							description={option.detail}
							onClick={() => chooseType(option.value)}
						/>
					))}
				</div>
			</EditorDrawer>
		);
	}

	return (
		<EditorDrawer open title={title} ariaLabel={title} onClose={onClose}>
			<form className={styles.drawerForm} onSubmit={handleSubmit}>
				<Panel tone="matte" title="Notification type">
					<SelectableRow as="div" leading={<NotificationTypeIcon type={selectedType.value} />} title={selectedType.label} description={selectedType.detail} />
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
									spellCheck={false}
									required
								/>
								<TextField
									label="Chat ID"
									value={form.chatId}
									onChange={event => updateForm({ chatId: event.currentTarget.value })}
									placeholder="-1001234567890"
									inputMode="numeric"
									autoComplete="off"
									spellCheck={false}
									required
								/>
							</>
						) : form.type === "email" ? (
							<TextAreaField
								label="Recipients"
								value={form.emailTo}
								onChange={event => updateForm({ emailTo: event.currentTarget.value })}
								placeholder="ops@example.com, sre@example.com"
								autoComplete="email"
								spellCheck={false}
								rows={4}
								required
							/>
						) : (
							<TextField
								label={notificationWebhookURLLabel(form.type)}
								value={form.url}
								onChange={event => updateForm({ url: event.currentTarget.value })}
								inputMode="url"
								type="url"
								autoComplete="url"
								spellCheck={false}
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
				</div>
				<UnsavedChangesBar show={hasNotificationChanges} saveType="submit" saving={isPending} disabled={!notificationFormReady(form)} onReset={() => setForm(initialForm)} />
			</form>
		</EditorDrawer>
	);
}
