import { formatDateTime, publicStatusPath } from "@/features/status-pages/api/statusPageAdapters";
import { useCreatePublicStatusPageMutation, useUpdatePublicStatusPageMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiPublicStatusPage, CreatePublicStatusPageInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, DataTable, Panel, Spinner, type DataColumn } from "@netstamp/ui";
import { CopyIcon } from "@phosphor-icons/react/dist/csr/Copy";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Link } from "react-router-dom";
import { StatusPageEditorDrawer } from "./StatusPageEditorDrawer";
import styles from "./StatusPagesPage.module.css";

type PageEditorState = { mode: "create" } | { mode: "edit"; page: ApiPublicStatusPage } | null;

const emptyPages: ApiPublicStatusPage[] = [];

function absolutePublicStatusURL(slug: string) {
	return new URL(publicStatusPath(slug), window.location.origin).toString();
}

export function StatusPagesPage() {
	const { projectRef } = useCurrentProject();
	const [pageEditor, setPageEditor] = useState<PageEditorState>(null);
	const pagesQuery = useQuery({
		...projectQueries.statusPages(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const pages = pagesQuery.data?.pages ?? emptyPages;
	const createPageMutation = useCreatePublicStatusPageMutation(projectRef);
	const updatePageMutation = useUpdatePublicStatusPageMutation(projectRef);
	const savingPage = createPageMutation.isPending || updatePageMutation.isPending;

	function handlePageSubmit(body: CreatePublicStatusPageInput) {
		if (pageEditor?.mode === "edit") {
			updatePageMutation.mutate(
				{ pageId: pageEditor.page.id, previousSlug: pageEditor.page.slug, body },
				{
					onSuccess: data => {
						setPageEditor(null);
						pushToast({ title: "Status page updated", message: data.page.title, tone: "success" });
					},
					onError: error => pushErrorToast(requestErrorMessage(error))
				}
			);
			return;
		}

		createPageMutation.mutate(body, {
			onSuccess: data => {
				setPageEditor(null);
				pushToast({ title: "Status page created", message: data.page.title, tone: "success" });
			},
			onError: error => pushErrorToast(requestErrorMessage(error))
		});
	}

	async function copyPageLink(page: ApiPublicStatusPage) {
		try {
			await navigator.clipboard.writeText(absolutePublicStatusURL(page.slug));
			pushToast({ title: "Link copied", message: page.title, tone: "success" });
		} catch {
			pushErrorToast("The status page link could not be copied.");
		}
	}

	const columns: DataColumn<ApiPublicStatusPage>[] = [
		{
			key: "title",
			label: "Title",
			sortable: true,
			sortValue: row => row.title,
			render: row => <strong className={styles.titleCell}>{row.title}</strong>
		},
		{
			key: "slug",
			label: "Slug",
			sortable: true,
			sortValue: row => row.slug,
			render: row => <span className={styles.slugCell}>/status/{row.slug}</span>
		},
		{
			key: "visibility",
			label: "Status",
			sortable: true,
			sortValue: row => (row.enabled ? 1 : 0),
			render: row => <Badge tone={row.enabled ? "success" : "neutral"}>{row.enabled ? "Public" : "Private"}</Badge>
		},
		{
			key: "updatedAt",
			label: "Last modified",
			sortable: true,
			sortValue: row => Date.parse(row.updatedAt),
			render: row => <time className={styles.timeCell}>{formatDateTime(row.updatedAt)}</time>
		}
	];

	return (
		<PageStack>
			<ScreenHeader
				title="Status Pages"
				actions={
					<Button type="button" onClick={() => setPageEditor({ mode: "create" })}>
						<PlusIcon aria-hidden="true" focusable="false" />
						New Page
					</Button>
				}
			/>

			<Panel title="Pages" actions={pagesQuery.isFetching ? <Badge tone="neutral">Syncing</Badge> : null} padded={false} bodySurface="transparent">
				{pagesQuery.isPending ? (
					<Spinner label="Loading status pages" layout="panel" size="lg" />
				) : (
					<DataTable
						columns={columns}
						rows={pages}
						density="compact"
						minWidth="52rem"
						ariaLabel="Project status pages"
						getRowKey={row => row.id}
						emptyLabel="No status pages yet. Create a page to share service health."
						rowActions={page => (
							<div className={styles.rowActions}>
								<Button type="button" variant="outline" size="sm" onClick={() => setPageEditor({ mode: "edit", page })}>
									Edit
								</Button>
								<Button type="button" variant="ghost" size="sm" onClick={() => void copyPageLink(page)}>
									<CopyIcon aria-hidden="true" focusable="false" />
									Copy Link
								</Button>
								<Button asChild type="button" variant="secondary" size="sm">
									<Link to={publicStatusPath(page.slug)}>View</Link>
								</Button>
							</div>
						)}
					/>
				)}
			</Panel>

			{pageEditor ? (
				<StatusPageEditorDrawer
					key={pageEditor.mode === "edit" ? pageEditor.page.id : "create"}
					open
					page={pageEditor.mode === "edit" ? pageEditor.page : null}
					saving={savingPage}
					onClose={() => setPageEditor(null)}
					onSubmit={handlePageSubmit}
				/>
			) : null}
		</PageStack>
	);
}
