import { useSession } from "@/features/auth/session/SessionContext";
import { isValidPublicPageSlug, PUBLIC_PAGE_SLUG_HELPER, sanitizePublicPageSlug } from "@/features/public-pages/publicPageSlug";
import { useCreateProjectPublicPageMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiPublicPage } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { optionalTrimmedText } from "@/shared/utils/formText";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, Checkbox, DataTable, Panel, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./PublicPagesPage.module.css";

interface PageRow {
	id: string;
	title: string;
	slug: string;
	enabled: boolean;
	folders: number;
	updatedAt: string;
}

interface CreatePageDraft {
	slug: string;
	title: string;
	description: string;
	enabled: boolean;
}

const EMPTY_PUBLIC_PAGES: ApiPublicPage[] = [];

function emptyCreateDraft(): CreatePageDraft {
	return { slug: "", title: "", description: "", enabled: true };
}

function formatDateTime(value: string) {
	return new Date(value).toLocaleString();
}

export function PublicPagesPage() {
	const navigate = useNavigate();
	const { session } = useSession();
	const { projectRef } = useCurrentProject();
	const [createDraft, setCreateDraft] = useState(emptyCreateDraft);
	const pagesQuery = useQuery({
		...projectQueries.publicPages(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const currentMember = membersQuery.data?.members.find(member => member.userId === session?.user.id);
	const canManage = currentMember?.role === "owner" || currentMember?.role === "admin";
	const pages = pagesQuery.data?.publicPages ?? EMPTY_PUBLIC_PAGES;
	const createPageMutation = useCreateProjectPublicPageMutation(projectRef);
	const canWrite = Boolean(projectRef && canManage);
	const slugValid = isValidPublicPageSlug(createDraft.slug);

	const pageRows: PageRow[] = pages.map(page => ({
		id: page.id,
		title: page.title,
		slug: page.slug,
		enabled: page.enabled,
		folders: page.folders?.length ?? 0,
		updatedAt: formatDateTime(page.updatedAt)
	}));
	const pageColumns: DataColumn<PageRow>[] = [
		{
			key: "page",
			label: "Page",
			render: row => (
				<span className={styles.stackedCell}>
					<strong>{row.title}</strong>
					<span>/s/{row.slug}</span>
				</span>
			)
		},
		{ key: "enabled", label: "State", render: row => <Badge tone={row.enabled ? "success" : "warning"}>{row.enabled ? "Enabled" : "Paused"}</Badge> },
		{ key: "folders", label: "Folders" },
		{ key: "updatedAt", label: "Updated" }
	];

	function createPage() {
		if (!slugValid || !createDraft.title.trim()) {
			return;
		}

		createPageMutation.mutate(
			{
				slug: createDraft.slug,
				title: createDraft.title,
				description: optionalTrimmedText(createDraft.description),
				enabled: createDraft.enabled
			},
			{
				onSuccess: data => {
					setCreateDraft(emptyCreateDraft());
					pushToast({ title: "Public page created", message: `/s/${data.publicPage.slug} is ready for folders.`, tone: "success" });
					navigate(`/public-pages/${data.publicPage.id}`);
				}
			}
		);
	}

	return (
		<PageStack>
			<ScreenHeader title="Public Pages" />

			<Panel tone="glass" title={`${pages.length} public pages`}>
				<div className={styles.panelStack}>
					{pagesQuery.isError ? (
						<div className={styles.errorState}>
							<div>
								<strong>Public pages unavailable</strong>
								<span>{requestErrorMessage(pagesQuery.error, "List public pages failed", { prefixFallback: true })}</span>
							</div>
							<Button variant="outline" size="sm" onClick={() => void pagesQuery.refetch()}>
								Retry
							</Button>
						</div>
					) : (
						<DataTable
							columns={pageColumns}
							rows={pageRows}
							density="compact"
							minWidth="42rem"
							onRowClick={row => navigate(`/public-pages/${row.id}`)}
							getRowKey={row => row.id}
							getRowAriaLabel={row => `Edit public page ${row.title}`}
							emptyLabel={pagesQuery.isLoading ? "Loading public pages" : "No public pages"}
						/>
					)}

					{canWrite && !pagesQuery.isError ? (
						<div className={styles.formGrid}>
							<TextField
								label="New page slug"
								value={createDraft.slug}
								maxLength={64}
								pattern="[a-z0-9-]+"
								helper={PUBLIC_PAGE_SLUG_HELPER}
								error={createDraft.slug && !slugValid ? PUBLIC_PAGE_SLUG_HELPER : undefined}
								onChange={event => {
									const value = sanitizePublicPageSlug(event.currentTarget.value);
									setCreateDraft(current => ({ ...current, slug: value }));
								}}
							/>
							<TextField
								label="Title"
								value={createDraft.title}
								onChange={event => {
									const { value } = event.currentTarget;
									setCreateDraft(current => ({ ...current, title: value }));
								}}
							/>
							<TextField
								label="Description"
								value={createDraft.description}
								onChange={event => {
									const { value } = event.currentTarget;
									setCreateDraft(current => ({ ...current, description: value }));
								}}
							/>
							<label className={styles.checkboxRow}>
								<Checkbox
									checked={createDraft.enabled}
									onChange={event => {
										const { checked } = event.currentTarget;
										setCreateDraft(current => ({ ...current, enabled: checked }));
									}}
								/>
								<span>Enabled</span>
							</label>
							<Button disabled={!slugValid || !createDraft.title.trim() || createPageMutation.isPending} onClick={createPage}>
								{createPageMutation.isPending ? "Creating" : "Create page"}
							</Button>
						</div>
					) : null}
				</div>
			</Panel>
		</PageStack>
	);
}
