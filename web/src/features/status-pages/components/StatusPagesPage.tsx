import { buildElementTree, chartModeLabel, chartRangeLabel, formatDateTime, publicStatusPath } from "@/features/status-pages/api/statusPageAdapters";
import {
	useCreatePublicStatusElementMutation,
	useCreatePublicStatusPageMutation,
	useDeletePublicStatusElementMutation,
	useDeletePublicStatusPageMutation,
	useUpdatePublicStatusElementMutation,
	useUpdatePublicStatusPageMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiPublicStatusElement, ApiPublicStatusPage, CreatePublicStatusElementInput, CreatePublicStatusPageInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, Button, DataTable, LoadingState, Panel, type DataColumn } from "@netstamp/ui";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { StatusElementEditorDrawer } from "./StatusElementEditorDrawer";
import { StatusElementTree } from "./StatusElementTree";
import { StatusPageEditorDrawer } from "./StatusPageEditorDrawer";
import styles from "./StatusPagesPage.module.css";

type PageEditorState = { mode: "create" } | { mode: "edit"; page: ApiPublicStatusPage } | null;
type ElementEditorState = { mode: "create" } | { mode: "edit"; element: ApiPublicStatusElement } | null;

const pageColumns: DataColumn<ApiPublicStatusPage>[] = [
	{
		key: "title",
		label: "Page",
		sortable: true,
		sortValue: row => row.title,
		render: row => (
			<div className={styles.pageCell}>
				<strong>{row.title}</strong>
				<span>{row.slug}</span>
			</div>
		)
	},
	{
		key: "enabled",
		label: "Status",
		sortable: true,
		sortValue: row => (row.enabled ? 1 : 0),
		render: row => <Badge tone={row.enabled ? "success" : "neutral"}>{row.enabled ? "Enabled" : "Disabled"}</Badge>
	},
	{
		key: "charts",
		label: "Charts",
		render: row => (
			<span className={styles.mono}>
				{chartModeLabel(row.defaultChartMode)} / {chartRangeLabel(row.defaultChartRange)}
			</span>
		)
	},
	{
		key: "updated",
		label: "Updated",
		sortable: true,
		sortValue: row => Date.parse(row.updatedAt),
		render: row => <span className={styles.mono}>{formatDateTime(row.updatedAt)}</span>
	}
];

const emptyPages: ApiPublicStatusPage[] = [];
const emptyElements: ApiPublicStatusElement[] = [];

export function StatusPagesPage() {
	const { projectRef } = useCurrentProject();
	const confirm = useConfirm();
	const [selectedPageID, setSelectedPageID] = useState("");
	const [pageEditor, setPageEditor] = useState<PageEditorState>(null);
	const [elementEditor, setElementEditor] = useState<ElementEditorState>(null);
	const pagesQuery = useQuery({
		...projectQueries.statusPages(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const pages = pagesQuery.data?.pages ?? emptyPages;
	const activePageID = selectedPageID && pages.some(page => page.id === selectedPageID) ? selectedPageID : (pages[0]?.id ?? "");
	const selectedPage = pages.find(page => page.id === activePageID) ?? null;
	const detailQuery = useQuery({
		...projectQueries.statusPageDetail(projectRef || "", selectedPage?.id || ""),
		enabled: Boolean(projectRef && selectedPage?.id)
	});
	const assignmentsQuery = useQuery({
		...projectQueries.assignments(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const elements = detailQuery.data?.elements ?? emptyElements;
	const elementTree = useMemo(() => buildElementTree(elements), [elements]);
	const assignments = assignmentsQuery.data?.assignments ?? [];
	const createPageMutation = useCreatePublicStatusPageMutation(projectRef);
	const updatePageMutation = useUpdatePublicStatusPageMutation(projectRef);
	const deletePageMutation = useDeletePublicStatusPageMutation(projectRef);
	const createElementMutation = useCreatePublicStatusElementMutation(projectRef);
	const updateElementMutation = useUpdatePublicStatusElementMutation(projectRef);
	const deleteElementMutation = useDeletePublicStatusElementMutation(projectRef);
	const savingPage = createPageMutation.isPending || updatePageMutation.isPending;
	const savingElement = createElementMutation.isPending || updateElementMutation.isPending;

	function handlePageSubmit(body: CreatePublicStatusPageInput) {
		if (pageEditor?.mode === "edit") {
			updatePageMutation.mutate(
				{ pageId: pageEditor.page.id, previousSlug: pageEditor.page.slug, body },
				{
					onSuccess: data => {
						setSelectedPageID(data.page.id);
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
				setSelectedPageID(data.page.id);
				setPageEditor(null);
				pushToast({ title: "Status page created", message: data.page.title, tone: "success" });
			},
			onError: error => pushErrorToast(requestErrorMessage(error))
		});
	}

	function handleElementSubmit(body: CreatePublicStatusElementInput) {
		if (!selectedPage) {
			return;
		}
		if (elementEditor?.mode === "edit") {
			updateElementMutation.mutate(
				{ pageId: selectedPage.id, elementId: elementEditor.element.id, body },
				{
					onSuccess: () => {
						setElementEditor(null);
						pushToast({ title: "Element updated", message: selectedPage.title, tone: "success" });
					},
					onError: error => pushErrorToast(requestErrorMessage(error))
				}
			);
			return;
		}

		createElementMutation.mutate(
			{ pageId: selectedPage.id, body },
			{
				onSuccess: () => {
					setElementEditor(null);
					pushToast({ title: "Element created", message: selectedPage.title, tone: "success" });
				},
				onError: error => pushErrorToast(requestErrorMessage(error))
			}
		);
	}

	async function deleteSelectedPage(page: ApiPublicStatusPage) {
		const confirmed = await confirm({
			title: "Delete status page?",
			message: `This disables and removes "${page.title}" from public access.`,
			confirmLabel: "Delete",
			tone: "danger"
		});
		if (!confirmed) {
			return;
		}
		deletePageMutation.mutate(
			{ pageId: page.id, slug: page.slug },
			{
				onSuccess: () => {
					setSelectedPageID("");
					pushToast({ title: "Status page deleted", message: page.title, tone: "success" });
				},
				onError: error => pushErrorToast(requestErrorMessage(error))
			}
		);
	}

	async function deleteElement(element: ApiPublicStatusElement) {
		if (!selectedPage) {
			return;
		}
		const confirmed = await confirm({
			title: "Delete element?",
			message: element.kind === "folder" ? "Folder children will be removed with the folder." : "This assignment group will no longer appear on the public page.",
			confirmLabel: "Delete",
			tone: "danger"
		});
		if (!confirmed) {
			return;
		}
		deleteElementMutation.mutate(
			{ pageId: selectedPage.id, elementId: element.id },
			{
				onSuccess: () => pushToast({ title: "Element deleted", message: selectedPage.title, tone: "success" }),
				onError: error => pushErrorToast(requestErrorMessage(error))
			}
		);
	}

	return (
		<PageStack>
			<ScreenHeader title="Status Pages" actions={<Button onClick={() => setPageEditor({ mode: "create" })}>New Status Page</Button>} />

			<div className={styles.layout}>
				<Panel title="Pages" actions={pagesQuery.isFetching ? <Badge tone="neutral">Syncing</Badge> : null} padded={false}>
					{pagesQuery.isPending ? (
						<LoadingState label="Loading status pages" detail="Fetching project public status pages." />
					) : (
						<DataTable
							columns={pageColumns}
							rows={pages}
							density="compact"
							ariaLabel="Project status pages"
							getRowKey={row => row.id}
							selectedKey={selectedPage?.id}
							onRowClick={page => setSelectedPageID(page.id)}
							emptyLabel="No status pages"
							rowActions={page => (
								<div className={styles.rowActions}>
									<Button type="button" variant="ghost" size="sm" onClick={() => setPageEditor({ mode: "edit", page })}>
										Edit
									</Button>
									<Button asChild type="button" variant="ghost" size="sm">
										<a href={publicStatusPath(page.slug)} target="_blank" rel="noreferrer">
											Open
										</a>
									</Button>
								</div>
							)}
						/>
					)}
				</Panel>

				<Panel
					title={selectedPage ? selectedPage.title : "Page Elements"}
					actions={
						selectedPage ? (
							<ActionRow>
								<Button asChild variant="outline" size="sm">
									<a href={publicStatusPath(selectedPage.slug)} target="_blank" rel="noreferrer">
										Public page
									</a>
								</Button>
								<Button type="button" variant="outline" size="sm" onClick={() => setElementEditor({ mode: "create" })}>
									<PlusIcon aria-hidden="true" focusable="false" />
									Add Element
								</Button>
								<Button type="button" variant="danger" size="sm" onClick={() => deleteSelectedPage(selectedPage)}>
									<TrashIcon aria-hidden="true" focusable="false" />
									Delete Page
								</Button>
							</ActionRow>
						) : null
					}
				>
					{selectedPage ? (
						<div className={styles.detailStack}>
							<div className={styles.detailMeta}>
								<Badge tone={selectedPage.enabled ? "success" : "neutral"}>{selectedPage.enabled ? "Enabled" : "Disabled"}</Badge>
								<span>{selectedPage.slug}</span>
								<span>{chartModeLabel(selectedPage.defaultChartMode)}</span>
								<span>{chartRangeLabel(selectedPage.defaultChartRange)}</span>
							</div>
							{detailQuery.isPending ? (
								<LoadingState label="Loading elements" detail="Fetching ordered public status elements." />
							) : (
								<StatusElementTree nodes={elementTree} onDelete={deleteElement} onEdit={element => setElementEditor({ mode: "edit", element })} />
							)}
						</div>
					) : (
						<div className={styles.emptyDetail}>Create a status page to start arranging public checks.</div>
					)}
				</Panel>
			</div>

			<StatusPageEditorDrawer
				key={pageEditor?.mode === "edit" ? pageEditor.page.id : (pageEditor?.mode ?? "closed")}
				open={Boolean(pageEditor)}
				page={pageEditor?.mode === "edit" ? pageEditor.page : null}
				saving={savingPage}
				onClose={() => setPageEditor(null)}
				onSubmit={handlePageSubmit}
			/>
			<StatusElementEditorDrawer
				key={elementEditor?.mode === "edit" ? elementEditor.element.id : (elementEditor?.mode ?? "closed")}
				open={Boolean(elementEditor)}
				element={elementEditor?.mode === "edit" ? elementEditor.element : null}
				elements={elements}
				assignments={assignments}
				saving={savingElement}
				onClose={() => setElementEditor(null)}
				onSubmit={handleElementSubmit}
			/>
		</PageStack>
	);
}
