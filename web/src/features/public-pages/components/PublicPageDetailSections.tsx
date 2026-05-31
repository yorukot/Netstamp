import { PUBLIC_PAGE_SLUG_HELPER, sanitizePublicPageSlug } from "@/features/public-pages/publicPageSlug";
import type { ApiCheck, ApiPublicPage, ApiPublicPageFolder } from "@/shared/api/types";
import { publicPageFolderLabel } from "@/shared/utils/publicPageFolders";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, Checkbox, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import styles from "./PublicPageDetailPage.module.css";

export interface PageDraft {
	pageId: string;
	slug: string;
	title: string;
	description: string;
	enabled: boolean;
}

export interface FolderDraft {
	parentId: string;
	name: string;
	description: string;
	sortOrder: string;
}

export interface FolderEditDraft extends FolderDraft {
	folderId: string;
}

interface SelectOption {
	value: string;
	label: string;
}

export function PublicPageSettingsSection({
	selectedPage,
	canWrite,
	pageDraftValue,
	slugValid,
	deletePending,
	updatePending,
	onDeletePage,
	onUpdatePage,
	onPageDraftChange
}: {
	selectedPage: ApiPublicPage;
	canWrite: boolean;
	pageDraftValue: PageDraft;
	slugValid: boolean;
	deletePending: boolean;
	updatePending: boolean;
	onDeletePage: () => void;
	onUpdatePage: () => void;
	onPageDraftChange: (patch: Partial<Omit<PageDraft, "pageId">>) => void;
}) {
	return (
		<div className={styles.detailSection}>
			<div className={styles.sectionHeader}>
				<div>
					<span>Page settings</span>
					<strong>/s/{selectedPage.slug}</strong>
				</div>
				{canWrite ? (
					<Button variant="danger" size="sm" disabled={deletePending} onClick={onDeletePage}>
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
							onPageDraftChange({ slug: value });
						}}
					/>
					<TextField
						label="Title"
						value={pageDraftValue.title}
						onChange={event => {
							const { value } = event.currentTarget;
							onPageDraftChange({ title: value });
						}}
					/>
					<TextField
						label="Description"
						value={pageDraftValue.description}
						onChange={event => {
							const { value } = event.currentTarget;
							onPageDraftChange({ description: value });
						}}
					/>
					<label className={styles.checkboxRow}>
						<Checkbox
							checked={pageDraftValue.enabled}
							onChange={event => {
								const { checked } = event.currentTarget;
								onPageDraftChange({ enabled: checked });
							}}
						/>
						<span>Enabled</span>
					</label>
					<Button disabled={!slugValid || !pageDraftValue.title || updatePending} onClick={onUpdatePage}>
						{updatePending ? "Saving" : "Save page"}
					</Button>
				</div>
			) : null}
		</div>
	);
}

export function PublicPageFoldersSection({
	folders,
	selectedFolder,
	canWrite,
	isLoading,
	folderDraft,
	folderOptions,
	folderEditValue,
	folderEditParentOptions,
	createPending,
	updatePending,
	onCreateFolder,
	onSelectFolder,
	onFolderDraftChange,
	onFolderEditChange,
	onUpdateFolder
}: {
	folders: ApiPublicPageFolder[];
	selectedFolder: ApiPublicPageFolder | null;
	canWrite: boolean;
	isLoading: boolean;
	folderDraft: FolderDraft;
	folderOptions: SelectOption[];
	folderEditValue: FolderEditDraft;
	folderEditParentOptions: SelectOption[];
	createPending: boolean;
	updatePending: boolean;
	onCreateFolder: () => void;
	onSelectFolder: (folderId: string) => void;
	onFolderDraftChange: (patch: Partial<FolderDraft>) => void;
	onFolderEditChange: (patch: Partial<Omit<FolderEditDraft, "folderId">>) => void;
	onUpdateFolder: () => void;
}) {
	return (
		<div className={styles.detailSection}>
			<div className={styles.sectionHeader}>
				<div>
					<span>Folders</span>
					<strong>{folders.length} nodes</strong>
				</div>
			</div>
			{canWrite ? (
				<div className={styles.folderGrid}>
					<SelectField label="Parent" value={folderDraft.parentId} options={folderOptions} onChange={event => onFolderDraftChange({ parentId: event.currentTarget.value })} />
					<TextField label="Name" value={folderDraft.name} onChange={event => onFolderDraftChange({ name: event.currentTarget.value })} />
					<TextField label="Sort" value={folderDraft.sortOrder} onChange={event => onFolderDraftChange({ sortOrder: event.currentTarget.value })} />
					<TextAreaField label="Description" value={folderDraft.description} onChange={event => onFolderDraftChange({ description: event.currentTarget.value })} rows={2} />
					<Button disabled={!folderDraft.name || createPending} onClick={onCreateFolder}>
						{createPending ? "Creating" : "Create folder"}
					</Button>
				</div>
			) : null}
			{folders.length ? (
				<div className={styles.folderList}>
					{folders.map(folder => (
						<button key={folder.id} type="button" className={folder.id === selectedFolder?.id ? styles.folderButtonActive : styles.folderButton} onClick={() => onSelectFolder(folder.id)}>
							<span>{publicPageFolderLabel(folder, folders)}</span>
							<small>{folder.checks?.length ?? 0} checks</small>
						</button>
					))}
				</div>
			) : (
				<div className={styles.emptyState}>{isLoading ? "Loading folders" : "No folders on this public page"}</div>
			)}
			{selectedFolder && canWrite ? (
				<div className={styles.folderEditor}>
					<div className={styles.sectionHeader}>
						<div>
							<span>Selected folder</span>
							<strong>{publicPageFolderLabel(selectedFolder, folders)}</strong>
						</div>
					</div>
					<div className={styles.folderGrid}>
						<SelectField label="Parent" value={folderEditValue.parentId} options={folderEditParentOptions} onChange={event => onFolderEditChange({ parentId: event.currentTarget.value })} />
						<TextField label="Name" value={folderEditValue.name} onChange={event => onFolderEditChange({ name: event.currentTarget.value })} />
						<TextField label="Sort" value={folderEditValue.sortOrder} onChange={event => onFolderEditChange({ sortOrder: event.currentTarget.value })} />
						<TextAreaField label="Description" value={folderEditValue.description} onChange={event => onFolderEditChange({ description: event.currentTarget.value })} rows={2} />
						<Button disabled={!folderEditValue.name.trim() || updatePending} onClick={onUpdateFolder}>
							{updatePending ? "Saving" : "Save folder"}
						</Button>
					</div>
				</div>
			) : null}
		</div>
	);
}

export function PublicPageChecksSection({
	selectedFolder,
	canWrite,
	publicChecks,
	selectedCheckIds,
	isChecksLoading,
	checksError,
	savePending,
	deletePending,
	onSaveChecks,
	onDeleteFolder,
	onRetryChecks,
	onToggleCheck
}: {
	selectedFolder: ApiPublicPageFolder;
	canWrite: boolean;
	publicChecks: ApiCheck[];
	selectedCheckIds: string[];
	isChecksLoading: boolean;
	checksError: unknown;
	savePending: boolean;
	deletePending: boolean;
	onSaveChecks: () => void;
	onDeleteFolder: () => void;
	onRetryChecks: () => void;
	onToggleCheck: (checkId: string, checked: boolean) => void;
}) {
	return (
		<div className={styles.checkSection}>
			<div className={styles.sectionHeader}>
				<div>
					<span>Published Ping checks</span>
					<strong>{selectedFolder.name}</strong>
				</div>
				{canWrite ? (
					<div className={styles.actionCluster}>
						<Button variant="outline" size="sm" disabled={savePending} onClick={onSaveChecks}>
							{savePending ? "Saving" : "Save checks"}
						</Button>
						<Button variant="danger" size="sm" disabled={deletePending} onClick={onDeleteFolder}>
							Delete folder
						</Button>
					</div>
				) : null}
			</div>
			<div className={styles.checkGrid}>
				{checksError ? (
					<div className={styles.errorState}>
						<div>
							<strong>Ping checks unavailable</strong>
							<span>{requestErrorMessage(checksError, "Load Ping checks failed", { prefixFallback: true })}</span>
						</div>
						<Button variant="outline" size="sm" onClick={onRetryChecks}>
							Retry
						</Button>
					</div>
				) : null}
				{publicChecks.map(check => {
					const checked = selectedCheckIds.includes(check.id);

					return (
						<label key={check.id} className={styles.checkOption}>
							<Checkbox checked={checked} disabled={!canWrite} onChange={event => onToggleCheck(check.id, event.currentTarget.checked)} />
							<span>
								<strong>{check.name}</strong>
								<small>{`${check.intervalSeconds}s interval`}</small>
							</span>
						</label>
					);
				})}
				{!checksError && !publicChecks.length ? <div className={styles.emptyState}>{isChecksLoading ? "Loading Ping checks" : "No Ping checks in this project"}</div> : null}
			</div>
		</div>
	);
}
