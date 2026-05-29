import { useSession } from "@/features/auth/session/SessionContext";
import {
	useCreateProjectPublicPageFolderMutation,
	useDeleteProjectPublicPageFolderMutation,
	useDeleteProjectPublicPageMutation,
	useSetProjectPublicPageFolderChecksMutation,
	useUpdateProjectPublicPageFolderMutation,
	useUpdateProjectPublicPageMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiPublicPage, ApiPublicPageFolder, UpdatePublicPageFolderInput, UpdatePublicPageInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushToast } from "@/shared/toast/toastStore";
import { nullableTrimmedText, optionalTrimmedText } from "@/shared/utils/formText";
import { isPublicPageDescendantFolder, publicPageFolderLabel } from "@/shared/utils/publicPageFolders";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { isValidPublicPageSlug } from "../publicPageSlug";
import styles from "./PublicPageDetailPage.module.css";
import { PublicPageChecksSection, PublicPageFoldersSection, PublicPageSettingsSection, type FolderDraft, type FolderEditDraft, type PageDraft } from "./PublicPageDetailSections";

interface FolderChecksDraft {
	folderId: string;
	checkIds: string[];
}

const EMPTY_FOLDERS: ApiPublicPageFolder[] = [];

function emptyFolderDraft(): FolderDraft {
	return { parentId: "", name: "", description: "", sortOrder: "0" };
}

function emptyFolderEditDraft(): FolderEditDraft {
	return { folderId: "", parentId: "", name: "", description: "", sortOrder: "0" };
}

function pageDraftFromPage(page: ApiPublicPage | null): PageDraft {
	return {
		pageId: page?.id ?? "",
		slug: page?.slug ?? "",
		title: page?.title ?? "",
		description: page?.description ?? "",
		enabled: page?.enabled ?? true
	};
}

function folderEditDraftFromFolder(folder: ApiPublicPageFolder | null): FolderEditDraft {
	return {
		folderId: folder?.id ?? "",
		parentId: folder?.parentId ?? "",
		name: folder?.name ?? "",
		description: folder?.description ?? "",
		sortOrder: String(folder?.sortOrder ?? 0)
	};
}

function pageUrl(slug: string) {
	return `${window.location.origin}/s/${slug}`;
}

function sortOrderValue(value: string) {
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0;
}

function pingChecks(checks: ApiCheck[] | null | undefined) {
	return (checks ?? []).filter(check => check.type === "ping");
}

export function PublicPageDetailPage() {
	const confirm = useConfirm();
	const navigate = useNavigate();
	const { pageId = "" } = useParams();
	const { session } = useSession();
	const { projectRef } = useCurrentProject();
	const [selectedFolderId, setSelectedFolderId] = useState("");
	const [folderDraft, setFolderDraft] = useState(emptyFolderDraft);
	const [folderEditDraft, setFolderEditDraft] = useState(emptyFolderEditDraft);
	const [pageDraft, setPageDraft] = useState<PageDraft>({ pageId: "", slug: "", title: "", description: "", enabled: true });
	const [folderChecksDraft, setFolderChecksDraft] = useState<FolderChecksDraft>({ folderId: "", checkIds: [] });
	const pageDetailQuery = useQuery({
		...projectQueries.publicPageDetail(projectRef || "", pageId),
		enabled: Boolean(projectRef && pageId)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const membersQuery = useQuery({
		...projectQueries.members(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const currentMember = membersQuery.data?.members.find(member => member.userId === session?.user.id);
	const canManage = currentMember?.role === "owner" || currentMember?.role === "admin";
	const canWrite = Boolean(projectRef && canManage);
	const selectedPage = pageDetailQuery.data?.publicPage ?? null;
	const pageDraftValue = pageDraft.pageId === selectedPage?.id ? pageDraft : pageDraftFromPage(selectedPage);
	const folders = selectedPage?.folders ?? EMPTY_FOLDERS;
	const resolvedSelectedFolderId = folders.some(folder => folder.id === selectedFolderId) ? selectedFolderId : (folders[0]?.id ?? "");
	const selectedFolder = folders.find(folder => folder.id === resolvedSelectedFolderId) ?? null;
	const folderEditValue = folderEditDraft.folderId === selectedFolder?.id ? folderEditDraft : folderEditDraftFromFolder(selectedFolder);
	const selectedCheckIds = folderChecksDraft.folderId === selectedFolder?.id ? folderChecksDraft.checkIds : (selectedFolder?.checks ?? []).map(check => check.id);
	const publicChecks = useMemo(() => pingChecks(checksQuery.data?.checks), [checksQuery.data?.checks]);
	const updatePageMutation = useUpdateProjectPublicPageMutation(projectRef);
	const deletePageMutation = useDeleteProjectPublicPageMutation(projectRef);
	const createFolderMutation = useCreateProjectPublicPageFolderMutation(projectRef);
	const updateFolderMutation = useUpdateProjectPublicPageFolderMutation(projectRef);
	const deleteFolderMutation = useDeleteProjectPublicPageFolderMutation(projectRef);
	const setFolderChecksMutation = useSetProjectPublicPageFolderChecksMutation(projectRef);
	const slugValid = isValidPublicPageSlug(pageDraftValue.slug);
	const folderOptions = [
		{ value: "", label: "Root folder" },
		...folders.map(folder => ({
			value: folder.id,
			label: publicPageFolderLabel(folder, folders)
		}))
	];
	const folderEditParentOptions = selectedFolder
		? [
				{ value: "", label: "Root folder" },
				...folders
					.filter(folder => folder.id !== selectedFolder.id && !isPublicPageDescendantFolder(folder, selectedFolder.id, folders))
					.map(folder => ({
						value: folder.id,
						label: publicPageFolderLabel(folder, folders)
					}))
			]
		: folderOptions;

	function updateSelectedPageDraft(patch: Partial<Omit<PageDraft, "pageId">>) {
		if (!selectedPage) {
			return;
		}

		setPageDraft(current => ({
			...pageDraftFromPage(selectedPage),
			...(current.pageId === selectedPage.id ? current : {}),
			...patch,
			pageId: selectedPage.id
		}));
	}

	function toggleCheck(checkId: string, checked: boolean) {
		if (!selectedFolder) {
			return;
		}

		setFolderChecksDraft({
			folderId: selectedFolder.id,
			checkIds: checked ? [...new Set([...selectedCheckIds, checkId])] : selectedCheckIds.filter(id => id !== checkId)
		});
	}

	function updateSelectedFolderDraft(patch: Partial<Omit<FolderEditDraft, "folderId">>) {
		if (!selectedFolder) {
			return;
		}

		setFolderEditDraft(current => ({
			...folderEditDraftFromFolder(selectedFolder),
			...(current.folderId === selectedFolder.id ? current : {}),
			...patch,
			folderId: selectedFolder.id
		}));
	}

	function updatePage() {
		if (!selectedPage || !slugValid) {
			return;
		}

		const body: UpdatePublicPageInput = {
			slug: pageDraftValue.slug,
			title: pageDraftValue.title,
			description: nullableTrimmedText(pageDraftValue.description),
			enabled: pageDraftValue.enabled
		};

		updatePageMutation.mutate(
			{
				pageId: selectedPage.id,
				body
			},
			{
				onSuccess: data => pushToast({ title: "Public page updated", message: `/s/${data.publicPage.slug} was saved.`, tone: "success" })
			}
		);
	}

	async function deletePage(page: ApiPublicPage) {
		const confirmed = await confirm({
			title: `Delete /s/${page.slug}?`,
			message: "This removes the public page, its folder tree, and all published check selections.",
			confirmLabel: "Delete public page",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deletePageMutation.mutate(page.id, {
			onSuccess: () => {
				pushToast({ title: "Public page deleted", message: `/s/${page.slug} was removed.`, tone: "success" });
				navigate("/public-pages");
			}
		});
	}

	function createFolder() {
		if (!selectedPage) {
			return;
		}

		createFolderMutation.mutate(
			{
				pageId: selectedPage.id,
				body: {
					parentId: folderDraft.parentId || undefined,
					name: folderDraft.name,
					description: optionalTrimmedText(folderDraft.description),
					sortOrder: sortOrderValue(folderDraft.sortOrder)
				}
			},
			{
				onSuccess: data => {
					setSelectedFolderId(data.folder.id);
					setFolderDraft(emptyFolderDraft());
					pushToast({ title: "Folder created", message: `${data.folder.name} is ready for Ping checks.`, tone: "success" });
				}
			}
		);
	}

	function updateFolder() {
		if (!selectedPage || !selectedFolder || !folderEditValue.name.trim()) {
			return;
		}

		const body: UpdatePublicPageFolderInput = {
			parentId: folderEditValue.parentId || null,
			name: folderEditValue.name,
			description: nullableTrimmedText(folderEditValue.description),
			sortOrder: sortOrderValue(folderEditValue.sortOrder)
		};

		updateFolderMutation.mutate(
			{ pageId: selectedPage.id, folderId: selectedFolder.id, body },
			{
				onSuccess: data => {
					setFolderEditDraft(folderEditDraftFromFolder(data.folder));
					pushToast({ title: "Folder updated", message: `${data.folder.name} was saved.`, tone: "success" });
				}
			}
		);
	}

	async function deleteFolder(folder: ApiPublicPageFolder) {
		if (!selectedPage) {
			return;
		}

		const confirmed = await confirm({
			title: `Delete ${folder.name}?`,
			message: "Nested folders and published check selections under this folder will be removed.",
			confirmLabel: "Delete folder",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteFolderMutation.mutate(
			{ pageId: selectedPage.id, folderId: folder.id },
			{
				onSuccess: () => pushToast({ title: "Folder deleted", message: `${folder.name} was removed.`, tone: "success" })
			}
		);
	}

	function saveFolderChecks() {
		if (!selectedPage || !selectedFolder) {
			return;
		}

		setFolderChecksMutation.mutate(
			{
				pageId: selectedPage.id,
				folderId: selectedFolder.id,
				body: { checkIds: selectedCheckIds }
			},
			{
				onSuccess: data => {
					setFolderChecksDraft({ folderId: selectedFolder.id, checkIds: data.checks.map(check => check.id) });
					pushToast({ title: "Published checks saved", message: `${data.checks.length} Ping checks are visible in ${selectedFolder.name}.`, tone: "success" });
				}
			}
		);
	}

	function copyPagePublicUrl(page: ApiPublicPage) {
		void navigator.clipboard?.writeText(pageUrl(page.slug));
		pushToast({ title: "Public URL copied", message: `/s/${page.slug}`, tone: "success" });
	}

	return (
		<PageStack>
			<ScreenHeader
				title={selectedPage ? selectedPage.title : "Public Page"}
				actions={
					<div className={styles.actionCluster}>
						<Button asChild variant="outline" size="sm">
							<Link to="/public-pages">Back</Link>
						</Button>
						{selectedPage ? (
							<>
								<Button asChild variant="outline" size="sm">
									<a href={`/s/${selectedPage.slug}`} target="_blank" rel="noreferrer">
										Open
									</a>
								</Button>
								<Button variant="ghost" size="sm" onClick={() => copyPagePublicUrl(selectedPage)}>
									Copy
								</Button>
							</>
						) : null}
					</div>
				}
			/>

			<Panel tone="glass" title={selectedPage ? `/s/${selectedPage.slug}` : "Public page"}>
				<div className={styles.panelStack}>
					{pageDetailQuery.isError ? (
						<div className={styles.errorState}>
							<div>
								<strong>Public page unavailable</strong>
								<span>{requestErrorMessage(pageDetailQuery.error, "Load public page detail failed", { prefixFallback: true })}</span>
							</div>
							<Button variant="outline" size="sm" onClick={() => void pageDetailQuery.refetch()}>
								Retry
							</Button>
						</div>
					) : null}

					{!selectedPage && !pageDetailQuery.isError ? <div className={styles.emptyState}>{pageDetailQuery.isLoading ? "Loading public page" : "Select a public page"}</div> : null}

					{selectedPage ? (
						<div className={styles.detailGrid}>
							<PublicPageSettingsSection
								selectedPage={selectedPage}
								canWrite={canWrite}
								pageDraftValue={pageDraftValue}
								slugValid={slugValid}
								deletePending={deletePageMutation.isPending}
								updatePending={updatePageMutation.isPending}
								onDeletePage={() => void deletePage(selectedPage)}
								onUpdatePage={updatePage}
								onPageDraftChange={updateSelectedPageDraft}
							/>
							<PublicPageFoldersSection
								folders={folders}
								selectedFolder={selectedFolder}
								canWrite={canWrite}
								isLoading={pageDetailQuery.isLoading}
								folderDraft={folderDraft}
								folderOptions={folderOptions}
								folderEditValue={folderEditValue}
								folderEditParentOptions={folderEditParentOptions}
								createPending={createFolderMutation.isPending}
								updatePending={updateFolderMutation.isPending}
								onCreateFolder={createFolder}
								onSelectFolder={setSelectedFolderId}
								onFolderDraftChange={patch => setFolderDraft(current => ({ ...current, ...patch }))}
								onFolderEditChange={updateSelectedFolderDraft}
								onUpdateFolder={updateFolder}
							/>
						</div>
					) : null}

					{selectedPage && selectedFolder ? (
						<PublicPageChecksSection
							selectedFolder={selectedFolder}
							canWrite={canWrite}
							publicChecks={publicChecks}
							selectedCheckIds={selectedCheckIds}
							isChecksLoading={checksQuery.isLoading}
							checksError={checksQuery.error}
							savePending={setFolderChecksMutation.isPending}
							deletePending={deleteFolderMutation.isPending}
							onSaveChecks={saveFolderChecks}
							onDeleteFolder={() => void deleteFolder(selectedFolder)}
							onRetryChecks={() => void checksQuery.refetch()}
							onToggleCheck={toggleCheck}
						/>
					) : null}
				</div>
			</Panel>
		</PageStack>
	);
}
