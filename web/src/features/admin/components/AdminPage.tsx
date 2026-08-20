import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import { useSession } from "@/features/auth/session/SessionContext";
import { formatDateTime } from "@/i18n/format";
import { useClearManagedUserPasswordMutation, useSetManagedUserPasswordMutation, useUpdateManagedUserMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import type { ApiManagedUser } from "@/shared/api/types";
import { useConfirm, usePromptDialog } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, BodyCopy, Button, DataTable, Panel, Spinner, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import styles from "./AdminPage.module.css";
import { AdminSettingsPanels } from "./AdminSettingsPanels";

function formatTimestamp(value: string | undefined) {
	if (!value) {
		return "";
	}
	const date = new Date(value);
	if (Number.isNaN(date.valueOf())) {
		return value;
	}
	return formatDateTime(date, {
		dateStyle: "medium",
		timeStyle: "short"
	});
}

function managedUserSearchText(user: ApiManagedUser, labels: string[]) {
	const [status, access, verification] = labels;

	return [user.displayName, user.email, status, access, verification, formatTimestamp(user.updatedAt), user.disabledAt ? formatTimestamp(user.disabledAt) : ""].join(" ").toLowerCase();
}

function filterManagedUsers(users: ApiManagedUser[], search: string, labels: (user: ApiManagedUser) => string[]) {
	const needle = search.trim().toLowerCase();

	if (!needle) {
		return users;
	}

	return users.filter(user => managedUserSearchText(user, labels(user)).includes(needle));
}

export function AdminPage() {
	const { t } = useTranslation("admin");
	const { session } = useSession();
	const confirm = useConfirm();
	const prompt = usePromptDialog();
	const requireSudo = useRequireSudo();
	const usersQuery = useQuery({ ...adminQueries.users(), enabled: Boolean(session?.user.isSystemAdmin) });
	const updateManagedUserMutation = useUpdateManagedUserMutation();
	const setManagedUserPasswordMutation = useSetManagedUserPasswordMutation();
	const clearManagedUserPasswordMutation = useClearManagedUserPasswordMutation();
	const [userSearch, setUserSearch] = useState("");
	const userRows = useMemo(() => usersQuery.data?.users ?? [], [usersQuery.data?.users]);
	const filteredUserRows = useMemo(
		() =>
			filterManagedUsers(userRows, userSearch, user => [
				user.disabledAt ? t("states.disabled") : t("states.active"),
				user.isSystemAdmin ? t("states.systemAdmin") : t("states.user"),
				user.emailVerified ? t("states.verified") : t("states.unverified")
			]),
		[userRows, userSearch, t]
	);
	const userCountLabel = userSearch.trim() ? t("filteredUsersCount", { filtered: filteredUserRows.length, total: userRows.length }) : t("usersCount", { count: userRows.length });
	const activeAdminCount = userRows.filter(user => user.isSystemAdmin && !user.disabledAt).length;

	const userColumns = useMemo<DataColumn<ApiManagedUser>[]>(
		() => [
			{
				key: "user",
				label: t("columns.user"),
				render: user => (
					<span className={styles.adminCell}>
						<strong className={styles.adminName}>{user.displayName}</strong>
						<span className={styles.adminMeta}>{user.email}</span>
					</span>
				),
				sortable: true,
				sortValue: user => user.email
			},
			{
				key: "status",
				label: t("columns.status"),
				render: user => (
					<span className={styles.adminCell}>
						<Badge tone={user.disabledAt ? "critical" : "success"}>{user.disabledAt ? t("states.disabled") : t("states.active")}</Badge>
						{user.disabledAt ? <span className={styles.adminMeta}>{formatTimestamp(user.disabledAt)}</span> : null}
					</span>
				),
				sortable: true,
				sortValue: user => (user.disabledAt ? 0 : 1)
			},
			{
				key: "access",
				label: t("columns.access"),
				render: user => <Badge tone={user.isSystemAdmin ? "accent" : "neutral"}>{user.isSystemAdmin ? t("states.systemAdmin") : t("states.user")}</Badge>,
				sortable: true,
				sortValue: user => (user.isSystemAdmin ? 1 : 0)
			},
			{
				key: "email",
				label: t("columns.email"),
				render: user => <Badge tone={user.emailVerified ? "success" : "warning"}>{user.emailVerified ? t("states.verified") : t("states.unverified")}</Badge>,
				sortable: true,
				sortValue: user => (user.emailVerified ? 1 : 0)
			},
			{
				key: "updatedAt",
				label: t("columns.updated"),
				render: user => <span className={styles.adminMeta}>{formatTimestamp(user.updatedAt)}</span>,
				sortable: true,
				sortValue: user => user.updatedAt
			}
		],
		[t]
	);

	if (!session) {
		return null;
	}

	if (!session.user.isSystemAdmin) {
		return (
			<PageStack>
				<ScreenHeader title={t("title")} />
				<Panel tone="deep" title={t("accessRequired")}>
					<BodyCopy>{t("accessRequiredDescription")}</BodyCopy>
				</Panel>
			</PageStack>
		);
	}
	const currentUserID = session.user.id;

	async function toggleDisabled(user: ApiManagedUser) {
		const nextDisabled = !user.disabledAt;
		if (nextDisabled) {
			const accepted = await confirm({
				title: t("account.disable"),
				message: t("account.disableDescription", { email: user.email }),
				confirmLabel: t("account.disableAction"),
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		await requireSudo(() =>
			updateManagedUserMutation.mutate(
				{ userId: user.id, body: { disabled: nextDisabled } },
				{
					onSuccess: data => {
						pushToast({
							title: data.user.disabledAt ? t("account.disabled") : t("account.enabled"),
							message: t("account.updated", { email: data.user.email }),
							tone: "success"
						});
					},
					onError: error => {
						pushToast({ title: t("account.updateFailed"), message: requestErrorMessage(error, t("account.updateError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function toggleSystemAdmin(user: ApiManagedUser) {
		const nextAdmin = !user.isSystemAdmin;
		if (!nextAdmin) {
			const accepted = await confirm({
				title: t("account.revokeAdmin"),
				message: t("account.revokeAdminDescription", { email: user.email }),
				confirmLabel: t("account.revoke"),
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		await requireSudo(() =>
			updateManagedUserMutation.mutate(
				{ userId: user.id, body: { systemAdmin: nextAdmin } },
				{
					onSuccess: data => {
						pushToast({
							title: data.user.isSystemAdmin ? t("account.adminGranted") : t("account.adminRevoked"),
							message: t("account.updated", { email: data.user.email }),
							tone: "success"
						});
					},
					onError: error => {
						pushToast({ title: t("account.permissionFailed"), message: requestErrorMessage(error, t("account.permissionError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function setPassword(user: ApiManagedUser) {
		const password = await prompt({
			title: t("account.setPassword"),
			message: t("account.setPasswordDescription", { email: user.email }),
			inputLabel: t("account.newPassword"),
			inputType: "password",
			confirmLabel: t("account.setPasswordAction"),
			validate: value => (value.length < 8 ? t("account.passwordTooShort") : null)
		});
		if (!password) {
			return;
		}

		await requireSudo(() =>
			setManagedUserPasswordMutation.mutate(
				{ userId: user.id, body: { password } },
				{
					onSuccess: data => {
						pushToast({ title: t("account.passwordUpdated"), message: t("account.passwordUpdatedDescription", { email: data.user.email }), tone: "success" });
					},
					onError: error => {
						pushToast({ title: t("account.passwordUpdateFailed"), message: requestErrorMessage(error, t("account.passwordUpdateError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function clearPassword(user: ApiManagedUser) {
		const accepted = await confirm({
			title: t("account.removePassword"),
			message: t("account.removePasswordDescription", { email: user.email }),
			confirmLabel: t("account.removePasswordAction"),
			tone: "danger"
		});
		if (!accepted) return;

		await requireSudo(() =>
			clearManagedUserPasswordMutation.mutate(user.id, {
				onSuccess: data => {
					pushToast({ title: t("account.passwordRemoved"), message: t("account.passwordRemovedDescription", { email: data.user.email }), tone: "success" });
				},
				onError: error => {
					pushToast({ title: t("account.passwordRemovalFailed"), message: requestErrorMessage(error, t("account.passwordRemovalError")), tone: "critical" });
				}
			})
		);
	}

	function userRowActions(user: ApiManagedUser) {
		const isSelf = user.id === currentUserID;
		const lastActiveAdmin = user.isSystemAdmin && !user.disabledAt && activeAdminCount <= 1;
		const updatingUser = updateManagedUserMutation.isPending && updateManagedUserMutation.variables?.userId === user.id;
		const settingPassword = setManagedUserPasswordMutation.isPending && setManagedUserPasswordMutation.variables?.userId === user.id;
		const clearingPassword = clearManagedUserPasswordMutation.isPending && clearManagedUserPasswordMutation.variables === user.id;

		return (
			<div className={styles.userActions}>
				<Button type="button" size="sm" variant={user.disabledAt ? "outline" : "danger"} disabled={isSelf || lastActiveAdmin || updatingUser} onClick={() => void toggleDisabled(user)}>
					{user.disabledAt ? t("account.enable") : t("account.disableAction")}
				</Button>
				<Button type="button" size="sm" variant="ghost" disabled={(isSelf && user.isSystemAdmin) || lastActiveAdmin || updatingUser} onClick={() => void toggleSystemAdmin(user)}>
					{user.isSystemAdmin ? t("account.revokeAdmin") : t("account.grantAdmin")}
				</Button>
				<Button type="button" size="sm" variant="outline" disabled={settingPassword} onClick={() => void setPassword(user)}>
					{settingPassword ? t("account.setting") : t("account.setPasswordAction")}
				</Button>
				{user.hasPassword ? (
					<Button type="button" size="sm" variant="danger" disabled={clearingPassword} onClick={() => void clearPassword(user)}>
						{clearingPassword ? t("account.removing") : t("account.removePasswordAction")}
					</Button>
				) : null}
			</div>
		);
	}

	return (
		<PageStack>
			<ScreenHeader title={t("title")} />

			<AdminSettingsPanels />
			<Panel
				tone="glass"
				title={t("users.title")}
				actions={usersQuery.isFetching ? <Badge tone="neutral">{t("users.syncing")}</Badge> : <Badge tone="neutral">{userCountLabel}</Badge>}
				bodySurface="transparent"
				padded={false}
			>
				{usersQuery.isLoading ? (
					<Spinner label={t("users.loading")} layout="panel" size="lg" />
				) : usersQuery.isError ? (
					<div className={styles.tableMessage}>
						<BodyCopy>{requestErrorMessage(usersQuery.error, t("users.loadError"))}</BodyCopy>
					</div>
				) : (
					<>
						<div className={styles.userToolbar}>
							<TextField label={t("users.search")} type="search" placeholder={t("users.searchPlaceholder")} value={userSearch} onChange={event => setUserSearch(event.currentTarget.value)} />
						</div>
						<DataTable<ApiManagedUser>
							ariaLabel={t("users.aria")}
							columns={userColumns}
							rows={filteredUserRows}
							density="compact"
							minWidth="72rem"
							emptyLabel={userSearch.trim() ? t("users.noMatch") : t("users.empty")}
							getRowKey={user => user.id}
							rowActions={userRowActions}
							rowActionsClassName={styles.userActionsCell}
							rowActionsHeaderClassName={styles.userActionsHeader}
						/>
					</>
				)}
			</Panel>
		</PageStack>
	);
}
