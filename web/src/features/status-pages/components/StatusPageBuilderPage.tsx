import { publicStatusPath } from "@/features/status-pages/api/statusPageAdapters";
import statusPageBanner from "@/features/status-pages/assets/status-page-banner.svg";
import { pathForRoute, pathForStatusPageEditor } from "@/routes/routePaths";
import {
	useCreatePublicStatusElementMutation,
	useCreatePublicStatusPageMutation,
	useDeletePublicStatusElementMutation,
	useUpdatePublicStatusElementMutation,
	useUpdatePublicStatusPageMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectAssignment, ApiPublicStatusElement, ApiPublicStatusPage, CreatePublicStatusElementInput, CreatePublicStatusPageInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import netstampLogoDark from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import netstampLogoLight from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Badge, Button, Checkbox, IconButton, Panel, SelectField, Spinner, Tabs, TextAreaField, TextField } from "@netstamp/ui";
import { ArrowDownIcon } from "@phosphor-icons/react/dist/csr/ArrowDown";
import { ArrowLeftIcon } from "@phosphor-icons/react/dist/csr/ArrowLeft";
import { ArrowUpIcon } from "@phosphor-icons/react/dist/csr/ArrowUp";
import { ChartLineIcon } from "@phosphor-icons/react/dist/csr/ChartLine";
import { ClockCounterClockwiseIcon } from "@phosphor-icons/react/dist/csr/ClockCounterClockwise";
import { DotsSixVerticalIcon } from "@phosphor-icons/react/dist/csr/DotsSixVertical";
import { FolderIcon } from "@phosphor-icons/react/dist/csr/Folder";
import { GearSixIcon } from "@phosphor-icons/react/dist/csr/GearSix";
import { GlobeIcon } from "@phosphor-icons/react/dist/csr/Globe";
import { MapTrifoldIcon } from "@phosphor-icons/react/dist/csr/MapTrifold";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { PulseIcon } from "@phosphor-icons/react/dist/csr/Pulse";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState, type DragEvent, type ReactNode } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { CssCodeEditor } from "./CssCodeEditor";
import styles from "./StatusPageBuilderPage.module.css";

type PageDraft = CreatePublicStatusPageInput;
type DisplayMode = CreatePublicStatusElementInput["displayMode"];

interface ElementDraft extends Omit<CreatePublicStatusElementInput, "parentElementId"> {
	localId: string;
	persistedId?: string;
	parentLocalId?: string;
}

interface CheckOption {
	id: string;
	name: string;
	type: string;
	target: string;
}

const themeOptions = [
	{ value: "auto", label: "Auto" },
	{ value: "light", label: "Light" },
	{ value: "dark", label: "Dark" }
];

const chartRangeOptions = [
	{ value: "24h", label: "24 hours" },
	{ value: "7d", label: "7 days" },
	{ value: "30d", label: "30 days" }
];

const displayOptions = [
	{ value: "status", label: "Service status" },
	{ value: "history", label: "Incident history" },
	{ value: "latency", label: "Latency chart" },
	{ value: "map", label: "Probe map" }
];

const blockLibrary: Array<{ mode: DisplayMode; title: string; description: string; icon: ReactNode }> = [
	{ mode: "status", title: "Service status", description: "Current availability, viewpoints, and 30-day bars.", icon: <PulseIcon aria-hidden="true" /> },
	{ mode: "history", title: "Incident history", description: "Emphasize recent disruptions and recovered periods.", icon: <ClockCounterClockwiseIcon aria-hidden="true" /> },
	{ mode: "latency", title: "Latency chart", description: "Show performance trend alongside availability.", icon: <ChartLineIcon aria-hidden="true" /> },
	{ mode: "map", title: "Probe map", description: "Show public monitoring locations on a map.", icon: <MapTrifoldIcon aria-hidden="true" /> }
];

function createDefaultPage(): PageDraft {
	return {
		slug: "new-status-page",
		title: "Service Status",
		description: "Live availability and incident updates for our services.",
		enabled: false,
		footerText: "Measurements are collected by Netstamp monitoring probes.",
		bannerImageUrl: undefined,
		theme: "auto",
		showTargets: false,
		showProbeNames: false,
		showProbeLocations: false,
		showIncidentHistory: true,
		showGeneratedAt: true,
		customCss: undefined,
		defaultChartMode: "off",
		defaultChartRange: "24h"
	};
}

function pageDraft(page: ApiPublicStatusPage): PageDraft {
	return {
		slug: page.slug,
		title: page.title,
		description: page.description,
		enabled: page.enabled,
		footerText: page.footerText,
		bannerImageUrl: page.bannerImageUrl,
		theme: page.theme,
		showTargets: page.showTargets,
		showProbeNames: page.showProbeNames,
		showProbeLocations: page.showProbeLocations,
		showIncidentHistory: page.showIncidentHistory,
		showGeneratedAt: page.showGeneratedAt,
		customCss: page.customCss,
		defaultChartMode: page.defaultChartMode,
		defaultChartRange: page.defaultChartRange
	};
}

function elementDraft(element: ApiPublicStatusElement): ElementDraft {
	return {
		localId: element.id,
		persistedId: element.id,
		parentLocalId: element.parentElementId,
		kind: element.kind,
		checkId: element.checkId,
		assignmentSelectionMode: element.assignmentSelectionMode,
		assignmentIds: element.assignmentIds,
		title: element.title,
		description: element.description,
		sortOrder: element.sortOrder,
		displayMode: element.displayMode,
		chartMode: element.chartMode,
		chartRange: element.chartRange
	};
}

function localID() {
	return typeof crypto !== "undefined" && "randomUUID" in crypto ? crypto.randomUUID() : `draft-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function sameValue(left: unknown, right: unknown) {
	return JSON.stringify(left) === JSON.stringify(right);
}

function normalizedPage(page: PageDraft): PageDraft {
	const optional = (value: string | undefined) => value?.trim() || undefined;
	return {
		...page,
		slug: page.slug.trim(),
		title: page.title.trim(),
		description: optional(page.description),
		footerText: optional(page.footerText),
		bannerImageUrl: optional(page.bannerImageUrl),
		customCss: optional(page.customCss)
	};
}

function checkOptions(assignments: ApiProjectAssignment[]) {
	const options = new Map<string, CheckOption>();
	for (const assignment of assignments) {
		if (!assignment.check || options.has(assignment.checkId)) continue;
		options.set(assignment.checkId, {
			id: assignment.checkId,
			name: assignment.check.name,
			type: assignment.check.type,
			target: assignment.check.target
		});
	}
	return [...options.values()].sort((left, right) => left.name.localeCompare(right.name));
}

export function StatusPageBuilderPage() {
	const { pageId = "new" } = useParams();
	const { projectRef } = useCurrentProject();
	const creating = pageId === "new";
	const detailQuery = useQuery({
		...projectQueries.statusPageDetail(projectRef || "", pageId),
		enabled: Boolean(projectRef) && !creating
	});
	const assignmentsQuery = useQuery({
		...projectQueries.assignments(projectRef || ""),
		enabled: Boolean(projectRef)
	});

	if (!projectRef || (!creating && detailQuery.isPending)) {
		return <Spinner label="Loading status page builder" layout="page" size="lg" />;
	}

	if (!creating && (detailQuery.error || !detailQuery.data)) {
		return (
			<PageStack>
				<ScreenHeader title="Status Page Editor" />
				<Panel tone="deep" title="Status page unavailable">
					<p className={styles.errorCopy}>This page could not be loaded or you no longer have access to it.</p>
				</Panel>
			</PageStack>
		);
	}

	const page = detailQuery.data?.page;
	return (
		<StatusPageBuilderWorkspace
			key={page?.id ?? "new"}
			projectRef={projectRef}
			pageId={page?.id}
			initialPage={page ? pageDraft(page) : createDefaultPage()}
			initialElements={(detailQuery.data?.elements ?? []).map(elementDraft)}
			assignments={assignmentsQuery.data?.assignments ?? []}
			assignmentsLoading={assignmentsQuery.isPending}
		/>
	);
}

function StatusPageBuilderWorkspace({
	projectRef,
	pageId,
	initialPage,
	initialElements,
	assignments,
	assignmentsLoading
}: {
	projectRef: string;
	pageId?: string;
	initialPage: PageDraft;
	initialElements: ElementDraft[];
	assignments: ApiProjectAssignment[];
	assignmentsLoading: boolean;
}) {
	const navigate = useNavigate();
	const [page, setPage] = useState<PageDraft>(initialPage);
	const [elements, setElements] = useState<ElementDraft[]>(initialElements);
	const [baselinePage, setBaselinePage] = useState<PageDraft>(initialPage);
	const [baselineElements, setBaselineElements] = useState<ElementDraft[]>(initialElements);
	const [activeTab, setActiveTab] = useState("page");
	const [selectedId, setSelectedId] = useState<string>();
	const draggedId = useRef<string | undefined>(undefined);
	const options = useMemo(() => checkOptions(assignments), [assignments]);
	const selectedElement = elements.find(element => element.localId === selectedId);
	const hasChanges = !sameValue(page, baselinePage) || !sameValue(elements, baselineElements);
	const createPageMutation = useCreatePublicStatusPageMutation(projectRef, { suppressGlobalErrorToast: true });
	const updatePageMutation = useUpdatePublicStatusPageMutation(projectRef, { suppressGlobalErrorToast: true });
	const createElementMutation = useCreatePublicStatusElementMutation(projectRef, { suppressGlobalErrorToast: true });
	const updateElementMutation = useUpdatePublicStatusElementMutation(projectRef, { suppressGlobalErrorToast: true });
	const deleteElementMutation = useDeletePublicStatusElementMutation(projectRef, { suppressGlobalErrorToast: true });
	const saving = createPageMutation.isPending || updatePageMutation.isPending || createElementMutation.isPending || updateElementMutation.isPending || deleteElementMutation.isPending;

	useEffect(() => {
		if (!hasChanges) return;
		const warn = (event: BeforeUnloadEvent) => event.preventDefault();
		window.addEventListener("beforeunload", warn);
		return () => window.removeEventListener("beforeunload", warn);
	}, [hasChanges]);

	function updatePage<K extends keyof PageDraft>(key: K, value: PageDraft[K]) {
		setPage(current => ({ ...current, [key]: value }));
	}

	function updateElement(localId: string, patch: Partial<ElementDraft>) {
		setElements(current => current.map(element => (element.localId === localId ? { ...element, ...patch } : element)));
	}

	function addGroup() {
		const localId = localID();
		const rootCount = elements.filter(element => element.kind === "folder").length;
		setElements(current => [
			...current,
			{
				localId,
				kind: "folder",
				assignmentIds: [],
				title: `Group ${rootCount + 1}`,
				sortOrder: rootCount,
				displayMode: "status",
				chartMode: "inherit"
			}
		]);
		setSelectedId(localId);
		setActiveTab("blocks");
	}

	function addBlock(mode: DisplayMode) {
		const firstCheck = options[0];
		if (!firstCheck) return;
		const firstGroup = sorted(elements.filter(element => element.kind === "folder"))[0];
		const siblings = elements.filter(element => element.kind === "assignment_group" && element.parentLocalId === firstGroup?.localId);
		const localId = localID();
		setElements(current => [
			...current,
			{
				localId,
				parentLocalId: firstGroup?.localId,
				kind: "assignment_group",
				checkId: firstCheck.id,
				assignmentSelectionMode: "all_check",
				assignmentIds: [],
				title: firstCheck.name,
				sortOrder: siblings.length,
				displayMode: mode,
				chartMode: mode === "latency" ? "compact" : "inherit",
				chartRange: mode === "latency" ? page.defaultChartRange : undefined
			}
		]);
		setSelectedId(localId);
		setActiveTab("blocks");
	}

	function removeElement(localId: string) {
		setElements(current => current.filter(element => element.localId !== localId && element.parentLocalId !== localId));
		if (selectedId === localId) setSelectedId(undefined);
	}

	function moveElement(localId: string, direction: -1 | 1) {
		setElements(current => {
			const source = current.find(element => element.localId === localId);
			if (!source) return current;
			const siblings = sorted(current.filter(element => element.kind === source.kind && element.parentLocalId === source.parentLocalId));
			const index = siblings.findIndex(element => element.localId === localId);
			const other = siblings[index + direction];
			if (!other) return current;
			return current.map(element => {
				if (element.localId === source.localId) return { ...element, sortOrder: other.sortOrder };
				if (element.localId === other.localId) return { ...element, sortOrder: source.sortOrder };
				return element;
			});
		});
	}

	function handleDrop(event: DragEvent<HTMLElement>, targetId: string, parentId?: string) {
		event.preventDefault();
		const sourceId = draggedId.current;
		if (!sourceId || sourceId === targetId) return;
		setElements(current => {
			const source = current.find(element => element.localId === sourceId);
			const target = current.find(element => element.localId === targetId);
			if (!source || !target || source.kind !== target.kind) return current;
			const destinationParent = source.kind === "folder" ? undefined : parentId;
			const siblings = sorted(current.filter(element => element.kind === source.kind && element.parentLocalId === destinationParent && element.localId !== sourceId));
			const targetIndex = Math.max(
				0,
				siblings.findIndex(element => element.localId === targetId)
			);
			siblings.splice(targetIndex, 0, { ...source, parentLocalId: destinationParent });
			const order = new Map(siblings.map((element, index) => [element.localId, index]));
			return current.map(element =>
				order.has(element.localId) ? { ...element, parentLocalId: element.localId === sourceId ? destinationParent : element.parentLocalId, sortOrder: order.get(element.localId) ?? 0 } : element
			);
		});
	}

	function dropIntoGroup(event: DragEvent<HTMLElement>, groupId: string) {
		event.preventDefault();
		const sourceId = draggedId.current;
		if (!sourceId) return;
		setElements(current => {
			const source = current.find(element => element.localId === sourceId);
			if (!source || source.kind !== "assignment_group") return current;
			const sortOrder = current.filter(element => element.kind === "assignment_group" && element.parentLocalId === groupId && element.localId !== sourceId).length;
			return current.map(element => (element.localId === sourceId ? { ...element, parentLocalId: groupId, sortOrder } : element));
		});
	}

	function reset() {
		setPage(baselinePage);
		setElements(baselineElements);
		setSelectedId(undefined);
	}

	async function save() {
		const body = normalizedPage(page);
		if (!body.title || !body.slug || !/^[a-z0-9-]+$/.test(body.slug)) {
			pushErrorToast("Add a title and a lowercase slug using letters, numbers, or hyphens.");
			setActiveTab("page");
			return;
		}
		if (elements.some(element => element.kind === "assignment_group" && !element.checkId)) {
			pushErrorToast("Every status block needs a check.");
			setActiveTab("blocks");
			return;
		}

		try {
			let savedPageId = pageId;
			let previousSlug = baselinePage.slug;
			if (savedPageId) {
				await updatePageMutation.mutateAsync({ pageId: savedPageId, previousSlug, body });
			} else {
				const created = await createPageMutation.mutateAsync(body);
				savedPageId = created.page.id;
				previousSlug = created.page.slug;
			}

			const savedElements = await persistElements(savedPageId);
			setPage(body);
			setElements(savedElements);
			setBaselinePage(body);
			setBaselineElements(savedElements);
			pushToast({ title: pageId ? "Status page updated" : "Status page created", message: body.title, tone: "success" });

			if (!pageId) {
				navigate(pathForStatusPageEditor(projectRef, savedPageId), { replace: true });
			}
		} catch (error) {
			pushErrorToast(requestErrorMessage(error));
		}
	}

	async function persistElements(savedPageId: string) {
		const next = elements.map(element => ({ ...element }));
		const createdIDs = new Map<string, string>();
		const removedIDs = baselineElements
			.filter(element => element.persistedId && !elements.some(candidate => candidate.persistedId === element.persistedId))
			.sort((left, right) => (left.kind === "folder" ? 1 : 0) - (right.kind === "folder" ? 1 : 0))
			.map(element => element.persistedId as string);

		for (const element of sorted(next.filter(candidate => !candidate.persistedId && candidate.kind === "folder"))) {
			const created = await createElementMutation.mutateAsync({ pageId: savedPageId, body: elementRequest(element) });
			element.persistedId = created.element.id;
			createdIDs.set(element.localId, created.element.id);
		}

		for (const element of sorted(next.filter(candidate => !candidate.persistedId && candidate.kind === "assignment_group"))) {
			const created = await createElementMutation.mutateAsync({ pageId: savedPageId, body: elementRequest(element, next, createdIDs) });
			element.persistedId = created.element.id;
			createdIDs.set(element.localId, created.element.id);
		}

		for (const element of next.filter(candidate => candidate.persistedId)) {
			if (createdIDs.has(element.localId)) continue;
			await updateElementMutation.mutateAsync({ pageId: savedPageId, elementId: element.persistedId as string, body: elementRequest(element, next, createdIDs) });
		}

		for (const elementId of removedIDs) {
			await deleteElementMutation.mutateAsync({ pageId: savedPageId, elementId });
		}

		return next;
	}

	return (
		<PageStack className={styles.pageStack}>
			<ScreenHeader
				title={pageId ? `Edit ${page.title}` : "New Status Page"}
				actions={
					<div className={styles.headerActions}>
						<Button asChild variant="ghost">
							<Link to={pathForRoute("statusPages", { projectRef })}>
								<ArrowLeftIcon aria-hidden="true" />
								Pages
							</Link>
						</Button>
						{pageId ? (
							<Button asChild variant="secondary">
								<Link to={publicStatusPath(page.slug)} target="_blank" rel="noreferrer">
									<GlobeIcon aria-hidden="true" />
									Open
								</Link>
							</Button>
						) : null}
					</div>
				}
			/>

			<div className={styles.builder}>
				<aside className={styles.sidebar} aria-label="Status page settings">
					<div className={styles.sidebarHeader}>
						<div>
							<span>Builder</span>
							<strong>{page.enabled ? "Public page" : "Private draft"}</strong>
						</div>
						<Badge tone={page.enabled ? "success" : "neutral"}>{page.enabled ? "Live" : "Private"}</Badge>
					</div>
					<Tabs
						tabs={[
							{ value: "page", label: "Page" },
							{ value: "blocks", label: "Blocks", badge: elements.length || undefined }
						]}
						value={activeTab}
						ariaLabel="Builder sections"
						size="sm"
						onValueChange={setActiveTab}
					/>
					<div className={styles.sidebarScroll}>
						{activeTab === "page" ? <PageSettings page={page} update={updatePage} /> : null}
						{activeTab === "blocks" ? (
							selectedElement ? (
								<ElementSettings
									element={selectedElement}
									elements={elements}
									checks={options}
									update={patch => updateElement(selectedElement.localId, patch)}
									onBack={() => setSelectedId(undefined)}
								/>
							) : (
								<BlockLibrary hasChecks={Boolean(options.length)} loading={assignmentsLoading} onAddGroup={addGroup} onAddBlock={addBlock} />
							)
						) : null}
					</div>
					<UnsavedChangesBar className={styles.saveBar} show={hasChanges} saving={saving} onReset={reset} onSave={() => void save()} />
				</aside>

				<StatusPageCanvas
					page={page}
					elements={elements}
					checks={options}
					selectedId={selectedId}
					onSelect={id => {
						setSelectedId(id);
						setActiveTab("blocks");
					}}
					onAddGroup={addGroup}
					onAddBlock={() => addBlock("status")}
					onMove={moveElement}
					onRemove={removeElement}
					onDragStart={id => {
						draggedId.current = id;
					}}
					onDrop={handleDrop}
					onDropGroup={dropIntoGroup}
				/>
			</div>
		</PageStack>
	);
}

function elementRequest(element: ElementDraft, allElements: ElementDraft[] = [], createdIDs = new Map<string, string>()): CreatePublicStatusElementInput {
	const parent = element.parentLocalId ? allElements.find(candidate => candidate.localId === element.parentLocalId) : undefined;
	return {
		kind: element.kind,
		parentElementId: element.kind === "assignment_group" ? createdIDs.get(element.parentLocalId ?? "") || parent?.persistedId : undefined,
		checkId: element.kind === "assignment_group" ? element.checkId : undefined,
		assignmentSelectionMode: element.kind === "assignment_group" ? (element.assignmentSelectionMode ?? "all_check") : undefined,
		assignmentIds: element.kind === "assignment_group" && element.assignmentSelectionMode === "selected_assignments" ? element.assignmentIds : undefined,
		title: element.title?.trim() || undefined,
		description: element.description?.trim() || undefined,
		sortOrder: element.sortOrder,
		displayMode: element.kind === "folder" ? "status" : element.displayMode,
		chartMode: element.kind === "folder" ? "inherit" : element.chartMode,
		chartRange: element.kind === "folder" ? undefined : element.chartRange
	};
}

function sorted<T extends { sortOrder: number }>(values: T[]) {
	return [...values].sort((left, right) => left.sortOrder - right.sortOrder);
}

function PageSettings({ page, update }: { page: PageDraft; update: <K extends keyof PageDraft>(key: K, value: PageDraft[K]) => void }) {
	return (
		<div className={styles.settingsSection}>
			<div className={styles.sectionIntro}>
				<strong>Basic information</strong>
				<p>Identity, publishing, visual mode, and public data boundaries.</p>
			</div>
			<label className={styles.switchRow}>
				<span>
					<strong>Public visibility</strong>
					<small>Anyone with the link can view this page.</small>
				</span>
				<Checkbox checked={page.enabled} onChange={event => update("enabled", event.currentTarget.checked)} />
			</label>
			<TextField
				label="Slug"
				helper={`Public URL: /status/${page.slug || "your-page"}`}
				value={page.slug}
				maxLength={64}
				onChange={event => update("slug", event.currentTarget.value.toLowerCase().replace(/[^a-z0-9-]/g, "-"))}
			/>
			<TextField label="Title" value={page.title} maxLength={128} onChange={event => update("title", event.currentTarget.value)} />
			<TextAreaField label="Description" value={page.description ?? ""} maxLength={1024} rows={4} onChange={event => update("description", event.currentTarget.value)} />
			<TextAreaField label="Footer text" value={page.footerText ?? ""} maxLength={2048} rows={3} onChange={event => update("footerText", event.currentTarget.value)} />
			<TextField
				label="Banner image URL"
				helper="Use an absolute HTTPS image URL. Leave empty for the Netstamp banner."
				type="url"
				value={page.bannerImageUrl ?? ""}
				onChange={event => update("bannerImageUrl", event.currentTarget.value)}
			/>
			<SelectField label="Theme" value={page.theme} options={themeOptions} onChange={event => update("theme", event.currentTarget.value as PageDraft["theme"])} />

			<div className={styles.fieldGroup}>
				<div className={styles.fieldGroupHeading}>
					<strong>Public data</strong>
					<Badge tone="accent">Private by default</Badge>
				</div>
				<PrivacyToggle label="Show check targets" detail="May reveal IP addresses, hostnames, or internal URLs." checked={page.showTargets} onChange={value => update("showTargets", value)} />
				<PrivacyToggle label="Show probe names" detail="Publishes the configured probe display name." checked={page.showProbeNames} onChange={value => update("showProbeNames", value)} />
				<PrivacyToggle
					label="Show probe locations"
					detail="Required for map coordinates and location labels."
					checked={page.showProbeLocations}
					onChange={value => update("showProbeLocations", value)}
				/>
				<PrivacyToggle label="Incident history" detail="Include recently resolved incidents." checked={page.showIncidentHistory} onChange={value => update("showIncidentHistory", value)} />
				<PrivacyToggle label="Generated time" detail="Show when the page data was last calculated." checked={page.showGeneratedAt} onChange={value => update("showGeneratedAt", value)} />
			</div>

			<SelectField
				label="Default chart range"
				value={page.defaultChartRange}
				options={chartRangeOptions}
				onChange={event => update("defaultChartRange", event.currentTarget.value as PageDraft["defaultChartRange"])}
			/>
			<CssCodeEditor value={page.customCss ?? ""} onChange={value => update("customCss", value)} />
		</div>
	);
}

function PrivacyToggle({ label, detail, checked, onChange }: { label: string; detail: string; checked: boolean; onChange: (value: boolean) => void }) {
	return (
		<label className={styles.switchRow}>
			<span>
				<strong>{label}</strong>
				<small>{detail}</small>
			</span>
			<Checkbox checked={checked} onChange={event => onChange(event.currentTarget.checked)} />
		</label>
	);
}

function BlockLibrary({ hasChecks, loading, onAddGroup, onAddBlock }: { hasChecks: boolean; loading: boolean; onAddGroup: () => void; onAddBlock: (mode: DisplayMode) => void }) {
	return (
		<div className={styles.settingsSection}>
			<div className={styles.sectionIntro}>
				<strong>Add content</strong>
				<p>Choose a presentation, then select its check and group.</p>
			</div>
			<button type="button" className={styles.libraryCard} onClick={onAddGroup}>
				<FolderIcon aria-hidden="true" />
				<span>
					<strong>Group</strong>
					<small>Organize related services under a public heading.</small>
				</span>
				<PlusIcon aria-hidden="true" />
			</button>
			{blockLibrary.map(block => (
				<button key={block.mode} type="button" className={styles.libraryCard} disabled={!hasChecks || loading} onClick={() => onAddBlock(block.mode)}>
					{block.icon}
					<span>
						<strong>{block.title}</strong>
						<small>{block.description}</small>
					</span>
					<PlusIcon aria-hidden="true" />
				</button>
			))}
			{loading ? <Spinner label="Loading checks" layout="compact" size="sm" /> : null}
			{!loading && !hasChecks ? <p className={styles.inlineNotice}>Create a check and assign a probe before adding a status block.</p> : null}
		</div>
	);
}

function ElementSettings({
	element,
	elements,
	checks,
	update,
	onBack
}: {
	element: ElementDraft;
	elements: ElementDraft[];
	checks: CheckOption[];
	update: (patch: Partial<ElementDraft>) => void;
	onBack: () => void;
}) {
	const groups = sorted(elements.filter(candidate => candidate.kind === "folder"));
	return (
		<div className={styles.settingsSection}>
			<button type="button" className={styles.backButton} onClick={onBack}>
				<ArrowLeftIcon aria-hidden="true" />
				All blocks
			</button>
			<div className={styles.sectionIntro}>
				<strong>{element.kind === "folder" ? "Group settings" : "Block settings"}</strong>
				<p>Changes are reflected immediately in the preview.</p>
			</div>
			<TextField label="Title" value={element.title ?? ""} maxLength={1024} onChange={event => update({ title: event.currentTarget.value })} />
			<TextAreaField label="Description" value={element.description ?? ""} maxLength={1024} rows={3} onChange={event => update({ description: event.currentTarget.value })} />
			{element.kind === "assignment_group" ? (
				<>
					<SelectField
						label="Group"
						value={element.parentLocalId ?? ""}
						options={[{ value: "", label: "No group" }, ...groups.map(group => ({ value: group.localId, label: group.title || "Untitled group" }))]}
						onChange={event => update({ parentLocalId: event.currentTarget.value || undefined })}
					/>
					<SelectField
						label="Check"
						value={element.checkId ?? ""}
						options={[{ value: "", label: "Select a check", disabled: true }, ...checks.map(check => ({ value: check.id, label: `${check.name} / ${check.type}` }))]}
						onChange={event => {
							const check = checks.find(candidate => candidate.id === event.currentTarget.value);
							update({ checkId: check?.id, title: element.title || check?.name });
						}}
					/>
					<SelectField
						label="Display"
						value={element.displayMode}
						options={displayOptions}
						onChange={event => update({ displayMode: event.currentTarget.value as DisplayMode, chartMode: event.currentTarget.value === "latency" ? "compact" : element.chartMode })}
					/>
					{element.displayMode === "latency" ? (
						<SelectField
							label="Chart range"
							value={element.chartRange ?? "24h"}
							options={chartRangeOptions}
							onChange={event => update({ chartRange: event.currentTarget.value as ElementDraft["chartRange"] })}
						/>
					) : null}
				</>
			) : null}
		</div>
	);
}

function StatusPageCanvas({
	page,
	elements,
	checks,
	selectedId,
	onSelect,
	onAddGroup,
	onAddBlock,
	onMove,
	onRemove,
	onDragStart,
	onDrop,
	onDropGroup
}: {
	page: PageDraft;
	elements: ElementDraft[];
	checks: CheckOption[];
	selectedId?: string;
	onSelect: (id: string) => void;
	onAddGroup: () => void;
	onAddBlock: () => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onDragStart: (id: string) => void;
	onDrop: (event: DragEvent<HTMLElement>, targetId: string, parentId?: string) => void;
	onDropGroup: (event: DragEvent<HTMLElement>, groupId: string) => void;
}) {
	const groups = sorted(elements.filter(element => element.kind === "folder"));
	const ungrouped = sorted(elements.filter(element => element.kind === "assignment_group" && !element.parentLocalId));
	const previewTheme = page.theme === "auto" ? "dark" : page.theme;

	return (
		<section className={styles.canvas} aria-label="Live status page preview">
			<div className={styles.canvasToolbar}>
				<div>
					<strong>Live preview</strong>
					<span>{page.theme === "auto" ? "Auto / previewing dark" : page.theme}</span>
				</div>
				<div>
					<Button type="button" variant="outline" size="sm" onClick={onAddGroup}>
						<FolderIcon aria-hidden="true" />
						Add Group
					</Button>
					<Button type="button" size="sm" disabled={!checks.length} onClick={onAddBlock}>
						<PlusIcon aria-hidden="true" />
						Add Block
					</Button>
				</div>
			</div>

			<div className={styles.previewViewport} data-preview-theme={previewTheme}>
				<div className={styles.publicShell}>
					<header className={styles.previewHero}>
						<img className={styles.previewBanner} src={page.bannerImageUrl || statusPageBanner} alt="" />
						<div className={styles.previewHeroBody}>
							<div className={styles.previewBrand}>
								<img src={previewTheme === "dark" ? netstampLogoLight : netstampLogoDark} alt="Netstamp" />
								<span>Public status</span>
							</div>
							<div className={styles.previewTitleRow}>
								<div>
									<h1>{page.title || "Untitled status page"}</h1>
									{page.description ? <p>{page.description}</p> : null}
								</div>
								{page.showGeneratedAt ? (
									<span className={styles.previewGenerated}>
										Last checked <strong>just now</strong>
									</span>
								) : null}
							</div>
						</div>
					</header>

					<div className={styles.previewOverall}>
						<span className={styles.previewStatusMarker} aria-hidden="true" />
						<div>
							<strong>All systems operational</strong>
							<small>No active incidents reported</small>
						</div>
						<Badge className={styles.previewOverallBadge} tone="success">
							Operational
						</Badge>
					</div>

					{groups.map(group => (
						<PreviewGroup
							key={group.localId}
							group={group}
							first={group.localId === groups[0]?.localId}
							last={group.localId === groups[groups.length - 1]?.localId}
							blocks={sorted(elements.filter(element => element.kind === "assignment_group" && element.parentLocalId === group.localId))}
							checks={checks}
							page={page}
							selectedId={selectedId}
							onSelect={onSelect}
							onMove={onMove}
							onRemove={onRemove}
							onDragStart={onDragStart}
							onDrop={onDrop}
							onDropGroup={onDropGroup}
						/>
					))}

					{ungrouped.length ? (
						<section className={styles.previewGroup}>
							<div className={styles.previewGroupHeader}>
								<div>
									<h2>Other services</h2>
									<p>Blocks that are not assigned to a group.</p>
								</div>
							</div>
							<div className={styles.previewGroupBody}>
								{ungrouped.map((element, index) => (
									<PreviewBlock
										key={element.localId}
										element={element}
										check={checks.find(check => check.id === element.checkId)}
										page={page}
										selected={element.localId === selectedId}
										first={index === 0}
										last={index === ungrouped.length - 1}
										onSelect={onSelect}
										onMove={onMove}
										onRemove={onRemove}
										onDragStart={onDragStart}
										onDrop={event => onDrop(event, element.localId)}
									/>
								))}
							</div>
						</section>
					) : null}

					{!groups.length && !ungrouped.length ? (
						<div className={styles.previewEmpty}>
							<PulseIcon aria-hidden="true" />
							<strong>Build your service view</strong>
							<p>Add a group, then add status, history, latency, or map blocks.</p>
							<div>
								<Button type="button" variant="outline" size="sm" onClick={onAddGroup}>
									Add Group
								</Button>
								<Button type="button" size="sm" disabled={!checks.length} onClick={onAddBlock}>
									Add Block
								</Button>
							</div>
						</div>
					) : null}

					<footer className={styles.previewFooter}>
						<p>{page.footerText || "Measurements are collected by configured Netstamp probes."}</p>
						{page.showGeneratedAt ? <span>Updated just now</span> : null}
					</footer>
				</div>
			</div>
		</section>
	);
}

function PreviewGroup({
	group,
	first,
	last,
	blocks,
	checks,
	page,
	selectedId,
	onSelect,
	onMove,
	onRemove,
	onDragStart,
	onDrop,
	onDropGroup
}: {
	group: ElementDraft;
	first: boolean;
	last: boolean;
	blocks: ElementDraft[];
	checks: CheckOption[];
	page: PageDraft;
	selectedId?: string;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onDragStart: (id: string) => void;
	onDrop: (event: DragEvent<HTMLElement>, targetId: string, parentId?: string) => void;
	onDropGroup: (event: DragEvent<HTMLElement>, groupId: string) => void;
}) {
	return (
		<section
			className={`${styles.previewGroup} ${group.localId === selectedId ? styles.selectedElement : ""}`}
			draggable
			onDragStart={() => onDragStart(group.localId)}
			onDragOver={event => event.preventDefault()}
			onDrop={event => onDrop(event, group.localId)}
		>
			<div className={styles.previewGroupHeader} onDragOver={event => event.preventDefault()} onDrop={event => onDropGroup(event, group.localId)}>
				<div>
					<h2>{group.title || "Untitled group"}</h2>
					{group.description ? <p>{group.description}</p> : null}
				</div>
				<ElementControls element={group} first={first} last={last} onSelect={onSelect} onMove={onMove} onRemove={onRemove} />
			</div>
			<div className={styles.previewGroupBody}>
				{blocks.length ? (
					blocks.map((element, index) => (
						<PreviewBlock
							key={element.localId}
							element={element}
							check={checks.find(check => check.id === element.checkId)}
							page={page}
							selected={element.localId === selectedId}
							first={index === 0}
							last={index === blocks.length - 1}
							onSelect={onSelect}
							onMove={onMove}
							onRemove={onRemove}
							onDragStart={onDragStart}
							onDrop={event => onDrop(event, element.localId, group.localId)}
						/>
					))
				) : (
					<div className={styles.groupDropZone} onDragOver={event => event.preventDefault()} onDrop={event => onDropGroup(event, group.localId)}>
						Drag a status block into this group
					</div>
				)}
			</div>
		</section>
	);
}

function PreviewBlock({
	element,
	check,
	page,
	selected,
	first,
	last,
	onSelect,
	onMove,
	onRemove,
	onDragStart,
	onDrop
}: {
	element: ElementDraft;
	check?: CheckOption;
	page: PageDraft;
	selected: boolean;
	first: boolean;
	last: boolean;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onDragStart: (id: string) => void;
	onDrop: (event: DragEvent<HTMLElement>) => void;
}) {
	const metadata = [check?.type, page.showTargets ? check?.target : undefined, page.showProbeNames ? "Probe names" : undefined, page.showProbeLocations ? "Public locations" : undefined].filter(
		Boolean
	);

	return (
		<article
			className={`${styles.previewBlock} ${selected ? styles.selectedElement : ""}`}
			draggable
			onDragStart={() => onDragStart(element.localId)}
			onDragOver={event => event.preventDefault()}
			onDrop={onDrop}
		>
			<div className={styles.previewBlockTop}>
				<div className={styles.previewBlockIdentity}>
					<DotsSixVerticalIcon className={styles.dragIcon} aria-hidden="true" />
					<span className={styles.operationalDot} aria-hidden="true" />
					<div>
						<strong>{element.title || check?.name || "Untitled service"}</strong>
						<small>{metadata.length ? metadata.join(" / ") : displayLabel(element.displayMode)}</small>
					</div>
				</div>
				<div className={styles.previewBlockActions}>
					<span>Operational</span>
					<ElementControls element={element} first={first} last={last} onSelect={onSelect} onMove={onMove} onRemove={onRemove} />
				</div>
			</div>
			<div className={styles.uptimeBars} aria-label="Example 30-day uptime">
				{Array.from({ length: 30 }, (_, index) => (
					<span key={index} data-state={index === 11 && element.displayMode === "history" ? "degraded" : "operational"} />
				))}
			</div>
			{element.displayMode === "latency" ? (
				<svg className={styles.miniChart} viewBox="0 0 640 72" preserveAspectRatio="none" role="img" aria-label="Example latency trend">
					<path d="M0 52 C70 48 92 58 148 42 S242 35 300 44 S390 20 452 30 S552 42 640 16" />
				</svg>
			) : null}
			{element.displayMode === "map" && page.showProbeLocations ? (
				<div className={styles.miniMap} aria-label="Example public probe map">
					<span style={{ left: "22%", top: "58%" }} />
					<span style={{ left: "48%", top: "38%" }} />
					<span style={{ left: "78%", top: "62%" }} />
					<small>3 public monitoring locations</small>
				</div>
			) : null}
			{element.displayMode === "map" && !page.showProbeLocations ? <div className={styles.mapPrivacyNotice}>Enable public probe locations to render this map.</div> : null}
		</article>
	);
}

function ElementControls({
	element,
	first,
	last,
	onSelect,
	onMove,
	onRemove
}: {
	element: ElementDraft;
	first: boolean;
	last: boolean;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
}) {
	return (
		<div className={styles.elementControls}>
			<IconButton aria-label={`Move ${element.title || "element"} up`} disabled={first} onClick={() => onMove(element.localId, -1)}>
				<ArrowUpIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Move ${element.title || "element"} down`} disabled={last} onClick={() => onMove(element.localId, 1)}>
				<ArrowDownIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Configure ${element.title || "element"}`} onClick={() => onSelect(element.localId)}>
				<GearSixIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Remove ${element.title || "element"}`} danger onClick={() => onRemove(element.localId)}>
				<TrashIcon aria-hidden="true" />
			</IconButton>
		</div>
	);
}

function displayLabel(mode: DisplayMode) {
	return displayOptions.find(option => option.value === mode)?.label ?? "Service status";
}
