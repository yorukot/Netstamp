import { useSession } from "@/features/auth/session/SessionContext";
import { ApiError } from "@/shared/api/client";
import {
	useCreateProjectPublicPageFolderMutation,
	useDeleteProjectPublicPageFolderMutation,
	useDeleteProjectPublicPageMutation,
	useSetProjectPublicPageFolderChecksMutation,
	useUpdateProjectPublicPageFolderMutation,
	useUpdateProjectPublicPageMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiPublicPage, ApiPublicPageFolder, UpdatePublicPageFolderInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { useConfirm } from "@/shared/components/confirmContext";
import { pushToast } from "@/shared/toast/toastStore";
import { Button, Checkbox, Panel, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { isValidPublicPageSlug, PUBLIC_PAGE_SLUG_HELPER, sanitizePublicPageSlug } from "../publicPageSlug";
import styles from "./PublicPagesPage.module.css";

interface PageDraft {
	pageId: string;
	slug: string;
	title: string;
	description: string;
	enabled: boolean;
}

interface FolderDraft {
	parentId: string;
	name: string;
	description: string;
	sortOrder: string;
}

interface FolderEditDraft extends FolderDraft {
	folderId: string;
}

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

function optionalText(value: string) {
	const trimmed = value.trim();
	return trimmed ? trimmed : undefined;
}

function pageUrl(slug: string) {
	return `${window.location.origin}/s/${slug}`;
}

function sortOrderValue(value: string) {
	const parsed = Number.parseInt(value, 10);
	return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0;
}

function folderLabel(folder: ApiPublicPageFolder, folders: ApiPublicPageFolder[]) {
	const names: string[] = [folder.name];
	let current = folder;
	const guard = new Set<string>([folder.id]);

	while (current.parentId) {
		const parent = folders.find(candidate => candidate.id === current.parentId);
		if (!parent || guard.has(parent.id)) {
			break;
		}
		names.unshift(parent.name);
		guard.add(parent.id);
		current = parent;
	}

	return names.join(" / ");
}

function isDescendantFolder(folder: ApiPublicPageFolder, ancestorID: string, folders: ApiPublicPageFolder[]) {
	let current = folder;
	const guard = new Set<string>([folder.id]);

	while (current.parentId) {
		if (current.parentId === ancestorID) {
			return true;
		}
		const parent = folders.find(candidate => candidate.id === current.parentId);
		if (!parent || guard.has(parent.id)) {
			return false;
		}
		guard.add(parent.id);
		current = parent;
	}

	return false;
}

function pingChecks(checks: ApiCheck[] | null | undefined) {
	return (checks ?? []).filter(check => check.type === "ping");
}

function requestErrorMessage(error: unknown, fallback: string) {
	if (error instanceof ApiError) {
		return `${fallback}: ${error.message}`;
	}

	if (error instanceof Error) {
		return `${fallback}: ${error.message}`;
	}

	return fallback;
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
			label: folderLabel(folder, folders)
		}))
	];
	const folderEditParentOptions = selectedFolder
		? [
				{ value: "", label: "Root folder" },
				...folders
					.filter(folder => folder.id !== selectedFolder.id && !isDescendantFolder(folder, selectedFolder.id, folders))
					.map(folder => ({
						value: folder.id,
						label: folderLabel(folder, folders)
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

		updatePageMutation.mutate(
			{
				pageId: selectedPage.id,
				body: {
					slug: pageDraftValue.slug,
					title: pageDraftValue.title,
					description: optionalText(pageDraftValue.description),
					enabled: pageDraftValue.enabled
				}
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
					description: optionalText(folderDraft.description),
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

		const body = {
			parentId: folderEditValue.parentId || null,
			name: folderEditValue.name,
			description: optionalText(folderEditValue.description),
			sortOrder: sortOrderValue(folderEditValue.sortOrder)
		} as UpdatePublicPageFolderInput;

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
				eyebrow="Public page detail"
				title={selectedPage ? selectedPage.title : "Public Page"}
				copy={selectedPage ? `/s/${selectedPage.slug}` : "Manage page metadata, folders, and published Ping checks."}
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

			<Panel tone="glass" eyebrow="Editor" title={selectedPage ? `/s/${selectedPage.slug}` : "Public page"}>
				<div className={styles.panelStack}>
					{pageDetailQuery.isError ? (
						<div className={styles.errorState}>
							<div>
								<strong>Public page unavailable</strong>
								<span>{requestErrorMessage(pageDetailQuery.error, "Load public page detail failed")}</span>
							</div>
							<Button variant="outline" size="sm" onClick={() => void pageDetailQuery.refetch()}>
								Retry
							</Button>
						</div>
					) : null}

					{!selectedPage && !pageDetailQuery.isError ? <div className={styles.emptyState}>{pageDetailQuery.isLoading ? "Loading public page" : "Select a public page"}</div> : null}

					{selectedPage ? (
						<div className={styles.detailGrid}>
							<div className={styles.detailSection}>
								<div className={styles.sectionHeader}>
									<div>
										<span>Page settings</span>
										<strong>/s/{selectedPage.slug}</strong>
									</div>
									{canWrite ? (
										<Button variant="danger" size="sm" disabled={deletePageMutation.isPending} onClick={() => void deletePage(selectedPage)}>
											Delete
										</Button>
									) : null}
								</div>
								{canWrite ? (
									<div className={styles.editGrid}>
										<TextField
											label="Slug"
											value={pageDraftValue.slug}
											maxLength={64}
											pattern="[a-z0-9-]+"
											helper={PUBLIC_PAGE_SLUG_HELPER}
											error={slugValid ? undefined : PUBLIC_PAGE_SLUG_HELPER}
											onChange={event => {
												const value = sanitizePublicPageSlug(event.currentTarget.value);
												updateSelectedPageDraft({ slug: value });
											}}
										/>
										<TextField
											label="Title"
											value={pageDraftValue.title}
											onChange={event => {
												const { value } = event.currentTarget;
												updateSelectedPageDraft({ title: value });
											}}
										/>
										<TextField
											label="Description"
											value={pageDraftValue.description}
											onChange={event => {
												const { value } = event.currentTarget;
												updateSelectedPageDraft({ description: value });
											}}
										/>
										<label className={styles.checkboxRow}>
											<Checkbox
												checked={pageDraftValue.enabled}
												onChange={event => {
													const { checked } = event.currentTarget;
													updateSelectedPageDraft({ enabled: checked });
												}}
											/>
											<span>Enabled</span>
										</label>
										<Button disabled={!slugValid || !pageDraftValue.title || updatePageMutation.isPending} onClick={updatePage}>
											{updatePageMutation.isPending ? "Saving" : "Save page"}
										</Button>
									</div>
								) : null}
							</div>

							<div className={styles.detailSection}>
								<div className={styles.sectionHeader}>
									<div>
										<span>Folders</span>
										<strong>{folders.length} nodes</strong>
									</div>
								</div>
								{canWrite ? (
									<div className={styles.folderGrid}>
										<SelectField
											label="Parent"
											value={folderDraft.parentId}
											options={folderOptions}
											onChange={event => {
												const { value } = event.currentTarget;
												setFolderDraft(current => ({ ...current, parentId: value }));
											}}
										/>
										<TextField
											label="Name"
											value={folderDraft.name}
											onChange={event => {
												const { value } = event.currentTarget;
												setFolderDraft(current => ({ ...current, name: value }));
											}}
										/>
										<TextField
											label="Sort"
											value={folderDraft.sortOrder}
											onChange={event => {
												const { value } = event.currentTarget;
												setFolderDraft(current => ({ ...current, sortOrder: value }));
											}}
										/>
										<TextAreaField
											label="Description"
											value={folderDraft.description}
											onChange={event => {
												const { value } = event.currentTarget;
												setFolderDraft(current => ({ ...current, description: value }));
											}}
											rows={2}
										/>
										<Button disabled={!folderDraft.name || createFolderMutation.isPending} onClick={createFolder}>
											{createFolderMutation.isPending ? "Creating" : "Create folder"}
										</Button>
									</div>
								) : null}
								{folders.length ? (
									<div className={styles.folderList}>
										{folders.map(folder => (
											<button
												key={folder.id}
												type="button"
												className={folder.id === selectedFolder?.id ? styles.folderButtonActive : styles.folderButton}
												onClick={() => setSelectedFolderId(folder.id)}
											>
												<span>{folderLabel(folder, folders)}</span>
												<small>{folder.checks?.length ?? 0} checks</small>
											</button>
										))}
									</div>
								) : (
									<div className={styles.emptyState}>{pageDetailQuery.isLoading ? "Loading folders" : "No folders on this public page"}</div>
								)}
								{selectedFolder && canWrite ? (
									<div className={styles.folderEditor}>
										<div className={styles.sectionHeader}>
											<div>
												<span>Selected folder</span>
												<strong>{folderLabel(selectedFolder, folders)}</strong>
											</div>
										</div>
										<div className={styles.folderGrid}>
											<SelectField
												label="Parent"
												value={folderEditValue.parentId}
												options={folderEditParentOptions}
												onChange={event => {
													const { value } = event.currentTarget;
													updateSelectedFolderDraft({ parentId: value });
												}}
											/>
											<TextField
												label="Name"
												value={folderEditValue.name}
												onChange={event => {
													const { value } = event.currentTarget;
													updateSelectedFolderDraft({ name: value });
												}}
											/>
											<TextField
												label="Sort"
												value={folderEditValue.sortOrder}
												onChange={event => {
													const { value } = event.currentTarget;
													updateSelectedFolderDraft({ sortOrder: value });
												}}
											/>
											<TextAreaField
												label="Description"
												value={folderEditValue.description}
												onChange={event => {
													const { value } = event.currentTarget;
													updateSelectedFolderDraft({ description: value });
												}}
												rows={2}
											/>
											<Button disabled={!folderEditValue.name.trim() || updateFolderMutation.isPending} onClick={updateFolder}>
												{updateFolderMutation.isPending ? "Saving" : "Save folder"}
											</Button>
										</div>
									</div>
								) : null}
							</div>
						</div>
					) : null}

					{selectedPage && selectedFolder ? (
						<div className={styles.checkSection}>
							<div className={styles.sectionHeader}>
								<div>
									<span>Published Ping checks</span>
									<strong>{selectedFolder.name}</strong>
								</div>
								{canWrite ? (
									<div className={styles.actionCluster}>
										<Button variant="outline" size="sm" disabled={setFolderChecksMutation.isPending} onClick={saveFolderChecks}>
											{setFolderChecksMutation.isPending ? "Saving" : "Save checks"}
										</Button>
										<Button variant="danger" size="sm" disabled={deleteFolderMutation.isPending} onClick={() => void deleteFolder(selectedFolder)}>
											Delete folder
										</Button>
									</div>
								) : null}
							</div>
							<div className={styles.checkGrid}>
								{checksQuery.isError ? (
									<div className={styles.errorState}>
										<div>
											<strong>Ping checks unavailable</strong>
											<span>{requestErrorMessage(checksQuery.error, "Load Ping checks failed")}</span>
										</div>
										<Button variant="outline" size="sm" onClick={() => void checksQuery.refetch()}>
											Retry
										</Button>
									</div>
								) : null}
								{publicChecks.map(check => {
									const checked = selectedCheckIds.includes(check.id);

									return (
										<label key={check.id} className={styles.checkOption}>
											<Checkbox
												checked={checked}
												disabled={!canWrite}
												onChange={event => {
													const { checked: nextChecked } = event.currentTarget;
													toggleCheck(check.id, nextChecked);
												}}
											/>
											<span>
												<strong>{check.name}</strong>
												<small>{`${check.intervalSeconds}s interval`}</small>
											</span>
										</label>
									);
								})}
								{!checksQuery.isError && !publicChecks.length ? <div className={styles.emptyState}>{checksQuery.isLoading ? "Loading Ping checks" : "No Ping checks in this project"}</div> : null}
							</div>
						</div>
					) : null}
				</div>
			</Panel>
		</PageStack>
	);
}
