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
import { useTranslation } from "react-i18next";
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
	metricLabel,
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
	const { t } = useTranslation("alerts");
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
	const localizedTabs = alertTabs.map(tab => ({ ...tab, label: t(`tabs.${tab.value}`) }));
	const localizedIncidentStatusOptions = incidentStatusOptions.map(option => ({ ...option, label: t(`filters.${option.value}`) }));
	const localizedRuleStatusOptions = ruleStatusOptions.map(option => ({ ...option, label: t(option.value === "all" ? "filters.anyStatus" : `filters.${option.value}`) }));
	const localizedRuleCheckTypeOptions = ruleCheckTypeOptions.map(option => ({ ...option, label: option.value === "all" ? t("filters.anyType") : t(`rules.checkTypes.${option.value}`) }));
	const localizedNotificationStatusOptions = notificationStatusOptions.map(option => ({ ...option, label: t(option.value === "all" ? "filters.anyStatus" : `filters.${option.value}`) }));
	const localizedNotificationTypeOptions = notificationFilterTypeOptions.map(option => ({
		...option,
		label: option.value === "all" ? t("filters.anyType") : t(`notifications.types.${option.value}`)
	}));
	const ruleColumns: DataColumn<ApiAlertRule>[] = [
		{
			key: "name",
			label: t("rules.rule"),
			render: rule => (
				<div className={styles.primaryCell}>
					<strong>{rule.name}</strong>
					<span>{rule.description || t("rules.noDescription")}</span>
				</div>
			)
		},
		{
			key: "status",
			label: t("common.status"),
			render: rule => <Badge tone={rule.enabled ? "success" : "muted"}>{rule.enabled ? t("common.enabled") : t("common.disabled")}</Badge>
		},
		{
			key: "scope",
			label: t("rules.scope"),
			render: rule => <span className={styles.monoCell}>{formatRuleScope(rule)}</span>
		},
		{
			key: "condition",
			label: t("rules.condition"),
			render: rule => <span className={styles.monoCell}>{formatAlertCondition(rule.condition)}</span>
		},
		{
			key: "triggerAfter",
			label: t("rules.triggerAfter"),
			render: rule => formatDuration(rule.triggerAfterSeconds)
		},
		{
			key: "notify",
			label: t("rules.notify"),
			render: rule => <span className={styles.urlCell}>{notificationLabel(rule, notifications)}</span>
		},
		{
			key: "cooldown",
			label: t("rules.cooldown"),
			render: rule => formatDuration(rule.cooldownSeconds)
		},
		{
			key: "updatedAt",
			label: t("rules.updated"),
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
					{t("common.delete")}
				</Button>
			)
		}
	];
	const incidentColumns: DataColumn<ApiAlertIncident>[] = [
		{
			key: "severity",
			label: t("incidents.severity"),
			render: incident => <Badge tone={severityTone(incident.severity)}>{t(`rules.severityOptions.${incident.severity as AlertSeverity}`)}</Badge>
		},
		{
			key: "target",
			label: t("incidents.target"),
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
			label: t("incidents.why"),
			render: incident => (
				<div className={styles.primaryCell}>
					<strong>{formatIncidentReason(incident)}</strong>
					<span>{incident.lastEvaluationState}</span>
				</div>
			)
		},
		{
			key: "value",
			label: t("incidents.value"),
			render: incident => (typeof incident.lastValue === "number" ? formatThreshold(incident.lastSummary.metric, Number(incident.lastValue.toFixed(2))) : "-")
		},
		{
			key: "openedAt",
			label: t("incidents.opened"),
			render: incident => formatDateTime(incident.openedAt)
		},
		{
			key: "lastEvaluatedAt",
			label: t("incidents.lastChecked"),
			render: incident => formatDateTime(incident.lastEvaluatedAt)
		},
		{
			key: "notifications",
			label: t("incidents.notifications"),
			render: incident => (incident.suppressedNotificationCount > 0 ? t("incidents.suppressed", { count: incident.suppressedNotificationCount }) : formatDateTime(incident.lastNotificationSentAt))
		}
	];
	const notificationColumns: DataColumn<ApiNotification>[] = [
		{
			key: "name",
			label: t("notifications.notification"),
			render: notification => (
				<div className={styles.primaryCell}>
					<strong>{notification.name}</strong>
					<span>{notificationTypeLabel(notification.type)}</span>
				</div>
			)
		},
		{
			key: "status",
			label: t("common.status"),
			render: notification => <Badge tone={notification.enabled ? "success" : "muted"}>{notification.enabled ? t("common.enabled") : t("common.disabled")}</Badge>
		},
		{
			key: "url",
			label: t("notifications.destination"),
			render: notification => (
				<span className={styles.urlCell} title={notificationDestination(notification)}>
					{notificationDestination(notification)}
				</span>
			)
		},
		{
			key: "usedBy",
			label: t("notifications.usedBy"),
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
						{t("common.test")}
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
						{t("common.delete")}
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
			title: t("rules.deleteQuestion"),
			message: t("rules.deleteMessage", { name: rule.name }),
			confirmLabel: t("rules.delete"),
			tone: "danger"
		});

		if (!accepted) {
			return;
		}

		try {
			await deleteRuleMutation.mutateAsync(rule.id);
			pushToast({ title: t("rules.deleted"), message: rule.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function deleteNotification(notification: ApiNotification) {
		const usedBy = rulesUsingNotification(rules, notification.id).length;
		const accepted = await confirm({
			title: t("notifications.deleteQuestion"),
			message: usedBy ? t("notifications.usedMessage", { name: notification.name, count: usedBy }) : t("notifications.deleteMessage", { name: notification.name }),
			confirmLabel: t("notifications.delete"),
			tone: "danger"
		});

		if (!accepted) {
			return;
		}

		try {
			await deleteNotificationMutation.mutateAsync(notification.id);
			pushToast({ title: t("notifications.deleted"), message: notification.name, tone: "success" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function testNotification(notification: ApiNotification) {
		try {
			const response = await testNotificationMutation.mutateAsync(notification.id);
			if (response.result.delivered) {
				pushToast({ title: t("notifications.testDelivered"), message: notification.name, tone: "success" });
				return;
			}
			pushToast({ title: t("notifications.testFailed"), message: response.result.message || response.result.code || notification.name, tone: "critical" });
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	return (
		<PageStack>
			<ScreenHeader title={t("title")} />
			<section className={styles.summaryGrid} aria-label={t("summary.aria")}>
				<SummaryCard
					label={t("summary.openIncidents")}
					value={openIncidents.length}
					tone={openIncidents.length ? "critical" : "success"}
					detail={openIncidentsQuery.isLoading ? t("summary.loading") : openIncidents.length ? t("summary.needsAttention") : t("summary.noIncidents")}
				/>
				<SummaryCard label={t("summary.enabledRules")} value={enabledRules} tone={enabledRules ? "success" : "muted"} detail={t("summary.totalRules", { count: rules.length })} />
				<SummaryCard
					label={t("summary.notifications")}
					value={notifications.length}
					tone={notifications.length ? "accent" : "muted"}
					detail={notifications.length ? t("summary.ready") : t("summary.none")}
				/>
			</section>
			<Tabs tabs={localizedTabs} value={visibleTab} ariaLabel={t("sectionsAria")} onValueChange={value => changeAlertTab(value as AlertTab)} />
			{visibleTab === "incidents" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.singlePanelAction}>
							<SelectField
								label={t("filters.status")}
								value={incidentStatus}
								options={localizedIncidentStatusOptions}
								onChange={event => updateAlertSearchParam("incidentStatus", event.currentTarget.value, "open")}
							/>
						</div>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={incidentColumns}
						rows={incidents}
						density="compact"
						minWidth="58rem"
						getRowKey={incident => incident.id}
						getRowAriaLabel={incident => t("incidents.openAria", { id: shortID(incident.id) })}
						onRowClick={selectIncident}
						selectedKey={incidentId || undefined}
						emptyLabel={incidentsQuery.isLoading ? tableSpinner(t("incidents.loading")) : incidentStatus === "open" ? t("incidents.noOpen") : t("incidents.noMatch")}
					/>
				</Panel>
			) : null}
			{visibleTab === "rules" ? (
				<Panel className={styles.tablePanel} padded={false}>
					<div className={styles.tableToolbar}>
						<div className={styles.panelActions}>
							<TextField
								label={t("filters.search")}
								value={ruleSearch}
								onChange={event => updateAlertSearchParam("ruleSearch", event.currentTarget.value, "")}
								placeholder={t("filters.searchPlaceholder")}
							/>
							<SelectField
								label={t("filters.status")}
								value={ruleStatus}
								options={localizedRuleStatusOptions}
								onChange={event => updateAlertSearchParam("ruleStatus", event.currentTarget.value, "all")}
							/>
							<SelectField
								label={t("filters.type")}
								value={ruleCheckType}
								options={localizedRuleCheckTypeOptions}
								onChange={event => updateAlertSearchParam("ruleType", event.currentTarget.value, "all")}
							/>
						</div>
						<Button type="button" onClick={() => setRuleEditor({ mode: "create" })}>
							{t("rules.create")}
						</Button>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={ruleColumns}
						rows={filteredRules}
						density="compact"
						minWidth="72rem"
						getRowKey={rule => rule.id}
						getRowAriaLabel={rule => t("rules.editAria", { name: rule.name })}
						onRowClick={rule => setRuleEditor({ mode: "edit", rule })}
						selectedKey={ruleEditor?.rule?.id}
						emptyLabel={
							rulesQuery.isLoading ? (
								tableSpinner(t("rules.loading"))
							) : rules.length ? (
								t("rules.noMatch")
							) : (
								<EmptyState
									title={t("rules.empty")}
									action={
										<Button type="button" size="sm" variant="secondary" onClick={() => setRuleEditor({ mode: "create" })}>
											{t("rules.create")}
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
								label={t("filters.status")}
								value={notificationStatus}
								options={localizedNotificationStatusOptions}
								onChange={event => updateAlertSearchParam("notificationStatus", event.currentTarget.value, "all")}
							/>
							<SelectField
								label={t("filters.type")}
								value={notificationType}
								options={localizedNotificationTypeOptions}
								onChange={event => updateAlertSearchParam("notificationType", event.currentTarget.value, "all")}
							/>
						</div>
						<Button type="button" onClick={() => setNotificationEditor({ mode: "create" })}>
							{t("notifications.add")}
						</Button>
					</div>
					<DataTable
						className={styles.tableFrame}
						columns={notificationColumns}
						rows={filteredNotifications}
						density="compact"
						minWidth="52rem"
						getRowKey={notification => notification.id}
						getRowAriaLabel={notification => t("notifications.editAria", { name: notification.name })}
						onRowClick={notification => setNotificationEditor({ mode: "edit", notification })}
						selectedKey={notificationEditor?.notification?.id}
						emptyLabel={
							notificationsQuery.isLoading ? (
								tableSpinner(t("notifications.loading"))
							) : notifications.length ? (
								t("notifications.noMatch")
							) : (
								<EmptyState
									title={t("notifications.empty")}
									action={
										<Button type="button" size="sm" variant="secondary" onClick={() => setNotificationEditor({ mode: "create" })}>
											{t("notifications.add")}
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
								pushToast({ title: t("rules.updatedToast"), message: body.name, tone: "success" });
							} else {
								await createRuleMutation.mutateAsync(body);
								pushToast({ title: t("rules.createdToast"), message: body.name, tone: "success" });
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
								pushToast({ title: t("notifications.updatedToast"), message: body.name, tone: "success" });
							} else {
								const body = notificationPayload(form);
								await createNotificationMutation.mutateAsync(body);
								pushToast({ title: t("notifications.createdToast"), message: body.name, tone: "success" });
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
	const { t } = useTranslation("alerts");
	return (
		<EditorDrawer open title={t("incidents.detail")} ariaLabel={t("incidents.detail")} onClose={onClose}>
			{incident ? (
				<div className={styles.detailStack}>
					<div className={styles.detailHeader}>
						<Badge tone={incidentTone(incident.status)}>{t(`filters.${incident.status as "open" | "acknowledged" | "resolved"}`)}</Badge>
						<Badge tone={severityTone(incident.severity)}>{t(`rules.severityOptions.${incident.severity as AlertSeverity}`)}</Badge>
					</div>
					<Panel tone="deep" title={t("incidents.whatHappened")}>
						<p className={styles.detailLead}>{formatIncidentReason(incident)}</p>
						<div className={styles.keyValueGrid}>
							<KeyValueRow label={t("incidents.probe")} value={formatIncidentProbe(incident)} />
							<KeyValueRow label={t("incidents.check")} value={formatIncidentCheck(incident)} />
							<KeyValueRow label={t("incidents.target")} value={incidentCheckTarget(incident)} />
							<KeyValueRow label={t("incidents.state")} value={incident.lastEvaluationState} />
							<KeyValueRow label={t("incidents.value")} value={typeof incident.lastValue === "number" ? formatThreshold(incident.lastSummary.metric, Number(incident.lastValue.toFixed(2))) : "-"} />
							<KeyValueRow label={t("incidents.rule")} value={shortID(incident.ruleId)} />
						</div>
					</Panel>
					<Panel tone="deep" title={t("incidents.timeline")}>
						<div className={styles.keyValueGrid}>
							<KeyValueRow label={t("incidents.opened")} value={formatDateTime(incident.openedAt)} />
							<KeyValueRow label={t("incidents.resolved")} value={formatDateTime(incident.resolvedAt)} />
							<KeyValueRow label={t("incidents.lastChecked")} value={formatDateTime(incident.lastEvaluatedAt)} />
							<KeyValueRow label={t("incidents.lastTriggered")} value={formatDateTime(incident.lastTriggeredAt)} />
						</div>
					</Panel>
					<Panel tone="deep" title={t("incidents.notifications")}>
						<div className={styles.keyValueGrid}>
							<KeyValueRow label={t("incidents.lastSent")} value={formatDateTime(incident.lastNotificationSentAt)} />
							<KeyValueRow label={t("incidents.nextEligible")} value={formatDateTime(incident.nextNotificationEligibleAt)} />
							<KeyValueRow label={t("incidents.suppressedLabel")} value={String(incident.suppressedNotificationCount)} />
						</div>
					</Panel>
				</div>
			) : isLoading ? (
				<Spinner label={t("incidents.loadingDetail")} layout="compact" size="lg" />
			) : (
				<BodyCopy>{error ? requestErrorMessage(error) : t("incidents.unavailable")}</BodyCopy>
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
	const { t } = useTranslation("alerts");
	const initialForm = useMemo(() => (editor.mode === "edit" && editor.rule ? ruleFormFromRule(editor.rule) : defaultRuleForm()), [editor.mode, editor.rule]);
	const [form, setForm] = useState<RuleFormState>(() => initialForm);
	const metricSelectOptions = metricOptionsForForm(form).map(option => ({ ...option, label: metricLabel(option.value) }));
	const title = editor.mode === "edit" ? t("rules.edit") : t("rules.create");
	const localizedEnabledOptions = enabledOptions.map(option => ({ ...option, label: t(option.value === "true" ? "common.enabled" : "common.disabled") }));
	const localizedCheckTypeOptions = checkTypeOptions.map(option => ({ ...option, label: t(`rules.checkTypes.${option.value}`) }));
	const localizedOperatorOptions = operatorOptions.map(option => ({ ...option, label: t(`rules.operators.${option.value}`) }));
	const localizedSeverityOptions = severityOptions.map(option => ({ ...option, label: t(`rules.severityOptions.${option.value}`) }));
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
				<Panel tone="deep" title={t("rules.target")}>
					<div className={styles.formGrid}>
						<TextField label={t("common.name")} value={form.name} onChange={event => updateForm({ name: event.currentTarget.value })} maxLength={128} required />
						<TextAreaField label={t("common.description")} value={form.description} onChange={event => updateForm({ description: event.currentTarget.value })} rows={3} />
						<div className={styles.twoColumns}>
							<SelectField label={t("common.status")} value={form.enabled} options={localizedEnabledOptions} onChange={event => updateForm({ enabled: event.currentTarget.value })} />
							<SelectField label={t("rules.checkType")} value={form.checkType} options={localizedCheckTypeOptions} onChange={event => handleCheckTypeChange(event.currentTarget.value as CheckType)} />
						</div>
						<div className={styles.twoColumns}>
							<TextField label={t("rules.probeId")} value={form.probeId} onChange={event => updateForm({ probeId: event.currentTarget.value })} placeholder={t("rules.optionalProbe")} />
							<TextField label={t("rules.checkId")} value={form.checkId} onChange={event => updateForm({ checkId: event.currentTarget.value })} placeholder={t("rules.optionalCheck")} />
						</div>
					</div>
				</Panel>
				<Panel tone="deep" title={t("rules.conditionPanel")}>
					{checkTypeSupported ? (
						<div className={styles.formGrid}>
							<SelectField label={t("rules.metric")} value={form.metric} options={metricSelectOptions} onChange={event => updateForm({ metric: event.currentTarget.value as AlertMetric })} />
							<div className={styles.twoColumns}>
								<SelectField
									label={t("rules.operator")}
									value={form.operator}
									options={localizedOperatorOptions}
									onChange={event => updateForm({ operator: event.currentTarget.value as AlertOperator })}
								/>
								<TextField
									label={t("rules.threshold")}
									value={form.threshold}
									onChange={event => updateForm({ threshold: event.currentTarget.value })}
									inputMode="decimal"
									error={numberValidation.threshold.error || undefined}
									required
								/>
							</div>
							<TextField
								label={t("rules.triggerMinutes")}
								helper={t("rules.triggerHelper")}
								value={form.triggerAfterMinutes}
								onChange={event => updateForm({ triggerAfterMinutes: event.currentTarget.value })}
								inputMode="numeric"
								min={1}
								max={1440}
								error={numberValidation.triggerAfterMinutes.error || undefined}
								required
							/>
						</div>
					) : (
						<p className={styles.unsupportedNotice}>{t("rules.unsupported")}</p>
					)}
				</Panel>
				<Panel tone="deep" title={t("rules.notifyPanel")}>
					<div className={styles.formGrid}>
						<div className={styles.twoColumns}>
							<SelectField
								label={t("rules.severity")}
								value={form.severity}
								options={localizedSeverityOptions}
								onChange={event => updateForm({ severity: event.currentTarget.value as AlertSeverity })}
							/>
							<div className={styles.notificationPicker} aria-label={t("rules.notificationTargets")}>
								<span className={styles.notificationPickerLabel}>{t("rules.notifications")}</span>
								{notifications.length ? (
									notifications.map(notification => (
										<label className={styles.notificationPickerOption} key={notification.id}>
											<input type="checkbox" checked={form.selectedNotificationIds.includes(notification.id)} onChange={event => toggleNotification(notification.id, event.currentTarget.checked)} />
											<span>{notification.name}</span>
										</label>
									))
								) : (
									<span className={styles.notificationPickerEmpty}>{t("rules.noNotifications")}</span>
								)}
							</div>
						</div>
						<details className={styles.advancedTiming}>
							<summary>{t("rules.advancedTiming")}</summary>
							<div className={styles.threeColumns}>
								<TextField
									label={t("rules.windowSeconds")}
									value={form.windowSeconds}
									onChange={event => updateForm({ windowSeconds: event.currentTarget.value })}
									inputMode="numeric"
									min={60}
									max={86400}
									error={numberValidation.windowSeconds.error || undefined}
									required
								/>
								<TextField
									label={t("rules.minSamples")}
									value={form.minSamples}
									onChange={event => updateForm({ minSamples: event.currentTarget.value })}
									inputMode="numeric"
									min={1}
									max={10000}
									error={numberValidation.minSamples.error || undefined}
									required
								/>
								<TextField
									label={t("rules.cooldownSeconds")}
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
	const { t } = useTranslation("alerts");
	const isEditing = editor.mode === "edit";
	const initialForm = useMemo(
		() => (editor.mode === "edit" && editor.notification ? notificationFormFromNotification(editor.notification) : defaultNotificationForm()),
		[editor.mode, editor.notification]
	);
	const [form, setForm] = useState<NotificationFormState>(() => initialForm);
	const [step, setStep] = useState<NotificationEditorStep>(isEditing ? "detail" : "type");
	const selectedType = notificationTypeOption(form.type);
	const title = isEditing ? t("notifications.edit") : t("notifications.add");
	const localizedEnabledOptions = enabledOptions.map(option => ({ ...option, label: t(option.value === "true" ? "common.enabled" : "common.disabled") }));
	const localizedNotificationTypeOptions = notificationTypeOptions.map(option => notificationTypeOption(option.value));
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
					{localizedNotificationTypeOptions.map(option => (
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
				<Panel tone="deep" title={t("notifications.type")}>
					<SelectableRow as="div" leading={<NotificationTypeIcon type={selectedType.value} />} title={selectedType.label} description={selectedType.detail} />
				</Panel>
				<Panel tone="deep" title={t("notifications.settings", { type: notificationTypeLabel(form.type) })}>
					<div className={styles.formGrid}>
						<TextField label={t("common.name")} value={form.name} onChange={event => updateForm({ name: event.currentTarget.value })} maxLength={128} required />
						{form.type === "telegram" ? (
							<>
								<TextField
									label={t("notifications.botToken")}
									value={form.botToken}
									onChange={event => updateForm({ botToken: event.currentTarget.value })}
									placeholder="123456:telegram-bot-token"
									autoComplete="off"
									spellCheck={false}
									required
								/>
								<TextField
									label={t("notifications.chatId")}
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
								label={t("notifications.recipients")}
								helper={t("notifications.recipientsHelper")}
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
						<SelectField label={t("common.status")} value={form.enabled} options={localizedEnabledOptions} onChange={event => updateForm({ enabled: event.currentTarget.value })} />
					</div>
				</Panel>
				{!isEditing ? (
					<div className={styles.drawerActions}>
						<Button type="button" variant="ghost" disabled={isPending} onClick={() => setStep("type")}>
							{t("notifications.changeType")}
						</Button>
					</div>
				) : null}
				<UnsavedChangesBar show={hasNotificationChanges} saveType="submit" saving={isPending} disabled={!notificationFormReady(form)} onReset={() => setForm(initialForm)} />
			</form>
		</EditorDrawer>
	);
}
