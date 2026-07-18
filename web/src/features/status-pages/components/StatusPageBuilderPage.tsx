import { pathForStatusPageEditor } from "@/routes/routePaths";
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
import { Badge, Button, Checkbox, IconButton, Panel, SelectField, Spinner, TextAreaField, TextField } from "@netstamp/ui";
import { ArrowDownIcon } from "@phosphor-icons/react/dist/csr/ArrowDown";
import { ArrowLeftIcon } from "@phosphor-icons/react/dist/csr/ArrowLeft";
import { ArrowRightIcon } from "@phosphor-icons/react/dist/csr/ArrowRight";
import { ArrowUpIcon } from "@phosphor-icons/react/dist/csr/ArrowUp";
import { ChartLineIcon } from "@phosphor-icons/react/dist/csr/ChartLine";
import { ClockCounterClockwiseIcon } from "@phosphor-icons/react/dist/csr/ClockCounterClockwise";
import { DotsSixVerticalIcon } from "@phosphor-icons/react/dist/csr/DotsSixVertical";
import { FolderIcon } from "@phosphor-icons/react/dist/csr/Folder";
import { GearSixIcon } from "@phosphor-icons/react/dist/csr/GearSix";
import { MapTrifoldIcon } from "@phosphor-icons/react/dist/csr/MapTrifold";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { PulseIcon } from "@phosphor-icons/react/dist/csr/Pulse";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useRef, useState, type PointerEvent, type ReactNode } from "react";
import { flushSync } from "react-dom";
import { useNavigate, useParams } from "react-router-dom";
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
	assignmentIds: string[];
	viewpoints: Array<{
		id: string;
		name: string;
		locationName?: string;
		latitude?: number;
		longitude?: number;
	}>;
}

type DragDestination = { kind: "target"; targetId: string; parentId?: string; placement: "before" | "after" } | { kind: "group"; parentId?: string };

const supportedStatusCheckTypes = new Set(["http", "ping", "tcp"]);
const emptyAssignments: ApiProjectAssignment[] = [];

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

function newCoverURL() {
	const seed = typeof crypto !== "undefined" && "randomUUID" in crypto ? crypto.randomUUID() : `${Date.now()}-${Math.random().toString(16).slice(2)}`;
	return `https://picsum.photos/2000/500?random=${seed}`;
}

function createDefaultPage(withCover = true): PageDraft {
	return {
		slug: "new-status-page",
		title: "Service Status",
		description: "Live availability and incident updates for our services.",
		enabled: false,
		footerText: "Measurements are collected by Netstamp monitoring probes.",
		bannerImageUrl: withCover ? newCoverURL() : undefined,
		theme: "auto",
		showTargets: false,
		showProbeNames: true,
		showProbeLocations: true,
		showIncidentHistory: true,
		showGeneratedAt: true,
		customCss: undefined,
		defaultChartMode: "off",
		defaultChartRange: "24h"
	};
}

function pageDraft(page: ApiPublicStatusPage): PageDraft {
	const defaults = createDefaultPage(false);
	return {
		slug: page.slug,
		title: page.title,
		description: page.description,
		enabled: page.enabled,
		footerText: page.footerText ?? defaults.footerText,
		bannerImageUrl: page.bannerImageUrl,
		theme: page.theme ?? defaults.theme,
		showTargets: page.showTargets ?? defaults.showTargets,
		showProbeNames: page.showProbeNames ?? defaults.showProbeNames,
		showProbeLocations: page.showProbeLocations ?? defaults.showProbeLocations,
		showIncidentHistory: page.showIncidentHistory ?? defaults.showIncidentHistory,
		showGeneratedAt: page.showGeneratedAt ?? defaults.showGeneratedAt,
		customCss: page.customCss,
		defaultChartMode: page.defaultChartMode ?? defaults.defaultChartMode,
		defaultChartRange: page.defaultChartRange ?? defaults.defaultChartRange
	};
}

function elementDraft(element: ApiPublicStatusElement, checks: CheckOption[]): ElementDraft {
	const resolvedCheckId =
		element.checkId ??
		(element.assignmentSelectionMode === "selected_assignments" ? undefined : (checks.find(check => check.name === element.checkName || check.name === element.title)?.id ?? checks[0]?.id));
	return {
		localId: element.id,
		persistedId: element.id,
		parentLocalId: element.parentElementId,
		kind: element.kind,
		checkId: resolvedCheckId,
		assignmentSelectionMode: element.assignmentSelectionMode,
		assignmentIds: element.assignmentIds,
		title: element.title,
		description: element.description,
		sortOrder: element.sortOrder,
		displayMode: element.displayMode ?? "status",
		chartMode: element.chartMode ?? "inherit",
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
		if (!assignment.check || !supportedStatusCheckTypes.has(assignment.check.type)) continue;
		const option = options.get(assignment.checkId) ?? {
			id: assignment.checkId,
			name: assignment.check.name,
			type: assignment.check.type,
			target: assignment.check.target,
			assignmentIds: [],
			viewpoints: []
		};
		option.assignmentIds.push(assignment.id);
		if (assignment.probe && !option.viewpoints.some(viewpoint => viewpoint.id === assignment.probeId)) {
			option.viewpoints.push({
				id: assignment.probeId,
				name: assignment.probe.name,
				locationName: assignment.probe.locationName,
				latitude: assignment.probe.latitude,
				longitude: assignment.probe.longitude
			});
		}
		options.set(assignment.checkId, option);
	}
	return [...options.values()].sort((left, right) => left.name.localeCompare(right.name));
}

function defaultDashboardElements(checks: CheckOption[]): ElementDraft[] {
	if (!checks.length) return [];
	const serviceGroupId = localID();
	const networkGroupId = localID();
	const modes: Array<{ mode: DisplayMode; groupId: string; title: string; description: string }> = [
		{ mode: "status", groupId: serviceGroupId, title: "Current availability", description: "Live assignment health, uptime, and active viewpoints." },
		{ mode: "history", groupId: serviceGroupId, title: "Incident history", description: "Recent degradation and recovery context." },
		{ mode: "latency", groupId: networkGroupId, title: "Latency performance", description: "Current, median, and percentile response trends." },
		{ mode: "map", groupId: networkGroupId, title: "Monitoring locations", description: "Public probe coverage using configured coordinates." }
	];

	return [
		{
			localId: serviceGroupId,
			kind: "folder",
			assignmentIds: [],
			title: "Service health",
			description: "Availability and incident context for customer-facing services.",
			sortOrder: 0,
			displayMode: "status",
			chartMode: "inherit"
		},
		{
			localId: networkGroupId,
			kind: "folder",
			assignmentIds: [],
			title: "Performance and coverage",
			description: "Latency trends and the probes observing this network.",
			sortOrder: 1,
			displayMode: "status",
			chartMode: "inherit"
		},
		...modes.map((feature, index) => {
			const check = checks[index % checks.length] as CheckOption;
			return {
				localId: localID(),
				parentLocalId: feature.groupId,
				kind: "assignment_group" as const,
				checkId: check.id,
				assignmentSelectionMode: "all_check" as const,
				assignmentIds: [],
				title: `${check.name} · ${feature.title}`,
				description: feature.description,
				sortOrder: index % 2,
				displayMode: feature.mode,
				chartMode: feature.mode === "latency" ? ("compact" as const) : ("inherit" as const),
				chartRange: feature.mode === "latency" ? ("24h" as const) : undefined
			};
		})
	];
}

function checkForElement(element: ElementDraft, checks: CheckOption[]) {
	return checks.find(check => check.id === element.checkId) ?? checks.find(check => element.assignmentIds?.some(assignmentId => check.assignmentIds.includes(assignmentId)));
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
	const assignments = assignmentsQuery.data?.assignments ?? emptyAssignments;
	const checks = useMemo(() => checkOptions(assignments), [assignments]);
	const [defaultPage] = useState(createDefaultPage);
	const defaultElements = useMemo(() => defaultDashboardElements(checks), [checks]);

	if (!projectRef || assignmentsQuery.isPending || (!creating && detailQuery.isPending)) {
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
			initialPage={page ? pageDraft(page) : defaultPage}
			initialElements={page ? (detailQuery.data?.elements ?? []).map(element => elementDraft(element, checks)) : defaultElements}
			assignments={assignments}
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
	const [selectedId, setSelectedId] = useState<string>();
	const [addingBlock, setAddingBlock] = useState(false);
	const [draggingId, setDraggingId] = useState<string>();
	const builderRef = useRef<HTMLDivElement | null>(null);
	const draggedId = useRef<string | undefined>(undefined);
	const dragDestination = useRef<DragDestination | undefined>(undefined);
	const dragIndicator = useRef<HTMLElement | undefined>(undefined);
	const dragCleanup = useRef<(() => void) | undefined>(undefined);
	const elementsRef = useRef(elements);
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
		elementsRef.current = elements;
	}, [elements]);

	useEffect(
		() => () => {
			dragCleanup.current?.();
		},
		[]
	);

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
		setAddingBlock(false);
	}

	function addBlock(mode: DisplayMode, checkId: string, parentLocalId?: string) {
		const check = options.find(option => option.id === checkId);
		if (!check) return;
		const siblings = elements.filter(element => element.kind === "assignment_group" && element.parentLocalId === parentLocalId);
		const localId = localID();
		setElements(current => [
			...current,
			{
				localId,
				parentLocalId,
				kind: "assignment_group",
				checkId: check.id,
				assignmentSelectionMode: "all_check",
				assignmentIds: [],
				title: check.name,
				sortOrder: siblings.length,
				displayMode: mode,
				chartMode: mode === "latency" ? "compact" : "inherit",
				chartRange: mode === "latency" ? page.defaultChartRange : undefined
			}
		]);
		setSelectedId(localId);
		setAddingBlock(false);
	}

	function startAddingBlock() {
		setSelectedId(undefined);
		setAddingBlock(true);
	}

	function removeElement(localId: string) {
		setElements(current => current.filter(element => element.localId !== localId && element.parentLocalId !== localId));
		if (selectedId === localId) setSelectedId(undefined);
	}

	function sortRects() {
		const rects = new Map<string, DOMRect>();
		for (const element of builderRef.current?.querySelectorAll<HTMLElement>("[data-builder-sort-id]") ?? []) {
			const id = element.dataset.builderSortId;
			if (id) rects.set(id, element.getBoundingClientRect());
		}
		return rects;
	}

	function updateOrder(updater: (current: ElementDraft[]) => ElementDraft[]) {
		const before = sortRects();
		let changed = false;
		flushSync(() => {
			setElements(current => {
				const next = updater(current);
				changed = next !== current;
				return next;
			});
		});
		if (!changed || window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

		const after = sortRects();
		for (const [id, previousRect] of before) {
			const nextRect = after.get(id);
			const element = builderRef.current?.querySelector<HTMLElement>(`[data-builder-sort-id="${id}"]`);
			if (!nextRect || !element) continue;
			const deltaX = previousRect.left - nextRect.left;
			const deltaY = previousRect.top - nextRect.top;
			if (Math.abs(deltaX) < 1 && Math.abs(deltaY) < 1) continue;
			for (const animation of element.getAnimations()) {
				if (animation.id === "status-page-reorder") animation.cancel();
			}
			element.animate([{ transform: `translate(${deltaX}px, ${deltaY}px)` }, { transform: "translate(0, 0)" }], {
				id: "status-page-reorder",
				duration: 260,
				easing: "cubic-bezier(0.2, 0.8, 0.2, 1)"
			});
		}
	}

	function moveElement(localId: string, direction: -1 | 1) {
		updateOrder(current => {
			const source = current.find(element => element.localId === localId);
			if (!source) return current;
			const siblings = sorted(current.filter(element => element.kind === source.kind && element.parentLocalId === source.parentLocalId));
			const index = siblings.findIndex(element => element.localId === localId);
			const nextIndex = index + direction;
			if (index < 0 || nextIndex < 0 || nextIndex >= siblings.length) return current;
			const reordered = [...siblings];
			const [moved] = reordered.splice(index, 1);
			if (!moved) return current;
			reordered.splice(nextIndex, 0, moved);
			const order = new Map(reordered.map((element, orderIndex) => [element.localId, orderIndex]));
			return current.map(element => (order.has(element.localId) ? { ...element, sortOrder: order.get(element.localId) ?? element.sortOrder } : element));
		});
	}

	function moveElementToTarget(sourceId: string, targetId: string, parentId: string | undefined, placement: "before" | "after") {
		if (sourceId === targetId) return;
		updateOrder(current => {
			const source = current.find(element => element.localId === sourceId);
			const target = current.find(element => element.localId === targetId);
			if (!source || !target || source.kind !== target.kind) return current;
			const destinationParent = source.kind === "folder" ? undefined : parentId;
			const destinationSiblings = sorted(current.filter(element => element.kind === source.kind && element.parentLocalId === destinationParent && element.localId !== sourceId));
			const targetIndex = destinationSiblings.findIndex(element => element.localId === targetId);
			if (targetIndex < 0) return current;
			destinationSiblings.splice(targetIndex + (placement === "after" ? 1 : 0), 0, { ...source, parentLocalId: destinationParent });
			const destinationOrder = new Map(destinationSiblings.map((element, index) => [element.localId, index]));
			const sourceOrder = new Map(
				sorted(current.filter(element => element.kind === source.kind && element.parentLocalId === source.parentLocalId && element.localId !== sourceId)).map((element, index) => [
					element.localId,
					index
				])
			);
			const next = current.map(element => {
				if (element.localId === sourceId) {
					return { ...element, parentLocalId: destinationParent, sortOrder: destinationOrder.get(element.localId) ?? 0 };
				}
				if (destinationOrder.has(element.localId)) return { ...element, sortOrder: destinationOrder.get(element.localId) ?? element.sortOrder };
				if (source.parentLocalId !== destinationParent && sourceOrder.has(element.localId)) return { ...element, sortOrder: sourceOrder.get(element.localId) ?? element.sortOrder };
				return element;
			});
			return next.every((element, index) => element.parentLocalId === current[index]?.parentLocalId && element.sortOrder === current[index]?.sortOrder) ? current : next;
		});
	}

	function moveElementToGroupEnd(sourceId: string, groupId?: string) {
		updateOrder(current => {
			const source = current.find(element => element.localId === sourceId);
			if (!source || source.kind !== "assignment_group") return current;
			const destination = sorted(current.filter(element => element.kind === "assignment_group" && element.parentLocalId === groupId && element.localId !== sourceId));
			destination.push({ ...source, parentLocalId: groupId });
			const destinationOrder = new Map(destination.map((element, index) => [element.localId, index]));
			const sourceOrder = new Map(
				sorted(current.filter(element => element.kind === "assignment_group" && element.parentLocalId === source.parentLocalId && element.localId !== sourceId)).map((element, index) => [
					element.localId,
					index
				])
			);
			const next = current.map(element => {
				if (element.localId === sourceId) return { ...element, parentLocalId: groupId, sortOrder: destinationOrder.get(element.localId) ?? 0 };
				if (destinationOrder.has(element.localId)) return { ...element, sortOrder: destinationOrder.get(element.localId) ?? element.sortOrder };
				if (source.parentLocalId !== groupId && sourceOrder.has(element.localId)) return { ...element, sortOrder: sourceOrder.get(element.localId) ?? element.sortOrder };
				return element;
			});
			return next.every((element, index) => element.parentLocalId === current[index]?.parentLocalId && element.sortOrder === current[index]?.sortOrder) ? current : next;
		});
	}

	function startDragging(event: PointerEvent<HTMLElement>, localId: string) {
		if (event.button !== 0) return;
		event.preventDefault();
		dragCleanup.current?.();
		draggedId.current = localId;
		dragDestination.current = undefined;
		clearDragIndicator();
		setDraggingId(localId);

		const handlePointerMove = (pointerEvent: globalThis.PointerEvent) => moveDragging(pointerEvent);
		const handlePointerEnd = () => stopDragging(true);
		const handlePointerCancel = () => stopDragging(false);
		document.addEventListener("pointermove", handlePointerMove, { passive: false });
		document.addEventListener("pointerup", handlePointerEnd);
		document.addEventListener("pointercancel", handlePointerCancel);
		dragCleanup.current = () => {
			document.removeEventListener("pointermove", handlePointerMove);
			document.removeEventListener("pointerup", handlePointerEnd);
			document.removeEventListener("pointercancel", handlePointerCancel);
		};
	}

	function showDragIndicator(element: HTMLElement, placement: "before" | "after" | "inside") {
		if (dragIndicator.current !== element) {
			dragIndicator.current?.removeAttribute("data-builder-drag-over");
			dragIndicator.current = element;
		}
		element.dataset.builderDragOver = placement;
	}

	function clearDragIndicator() {
		dragIndicator.current?.removeAttribute("data-builder-drag-over");
		dragIndicator.current = undefined;
	}

	function moveDragging(event: globalThis.PointerEvent) {
		const sourceId = draggedId.current;
		if (!sourceId) return;
		event.preventDefault();
		const source = elementsRef.current.find(element => element.localId === sourceId);
		const hit = document.elementFromPoint(event.clientX, event.clientY)?.closest<HTMLElement>("[data-builder-sort-id], [data-builder-drop-parent]");
		if (!source || !hit) return;

		const targetElement = hit.closest<HTMLElement>("[data-builder-sort-id]");
		const targetId = targetElement?.dataset.builderSortId;
		const target = elementsRef.current.find(element => element.localId === targetId);
		if (targetId && targetId !== sourceId && target && target.kind === source.kind) {
			const rect = targetElement?.getBoundingClientRect();
			const placement = rect && event.clientY > rect.top + rect.height / 2 ? "after" : "before";
			const parentId = targetElement?.dataset.builderParentId || undefined;
			dragDestination.current = { kind: "target", targetId, parentId, placement };
			if (targetElement) showDragIndicator(targetElement, placement);
			return;
		}

		const dropTarget = hit.closest<HTMLElement>("[data-builder-drop-parent]");
		if (source.kind !== "assignment_group" || !dropTarget) return;
		const parentId = dropTarget.dataset.builderDropParent || undefined;
		dragDestination.current = { kind: "group", parentId };
		showDragIndicator(dropTarget, "inside");
	}

	function stopDragging(commit: boolean) {
		const sourceId = draggedId.current;
		const destination = dragDestination.current;
		dragCleanup.current?.();
		dragCleanup.current = undefined;
		draggedId.current = undefined;
		dragDestination.current = undefined;
		clearDragIndicator();
		setDraggingId(undefined);

		if (!commit || !sourceId || !destination) return;
		if (destination.kind === "target") {
			moveElementToTarget(sourceId, destination.targetId, destination.parentId, destination.placement);
		} else {
			moveElementToGroupEnd(sourceId, destination.parentId);
		}
	}

	function reset() {
		setPage(baselinePage);
		setElements(baselineElements);
		setSelectedId(undefined);
		setAddingBlock(false);
	}

	async function save() {
		const body = normalizedPage(page);
		if (!body.title || !body.slug || !/^[a-z0-9-]+$/.test(body.slug)) {
			pushErrorToast("Add a title and a lowercase slug using letters, numbers, or hyphens.");
			setSelectedId(undefined);
			setAddingBlock(false);
			return;
		}
		const invalidElement = elements.find(
			element =>
				element.kind === "assignment_group" &&
				((element.assignmentSelectionMode === "selected_assignments" && !element.assignmentIds?.length) || (element.assignmentSelectionMode !== "selected_assignments" && !element.checkId))
		);
		if (invalidElement) {
			setSelectedId(invalidElement.localId);
			setAddingBlock(false);
			pushErrorToast(invalidElement.assignmentSelectionMode === "selected_assignments" ? "Select at least one assignment for this block." : "Select a check for this block.");
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

	const editorTitle = addingBlock ? "Add Block" : selectedElement ? (selectedElement.kind === "folder" ? "Editing Group" : "Editing Block") : "Editing Page";

	return (
		<div ref={builderRef} className={styles.builder}>
			<aside className={styles.sidebar} aria-label="Status page settings">
				<div className={styles.sidebarHeader}>
					<div>
						<span>Status page builder</span>
						<strong>{editorTitle}</strong>
					</div>
					<Badge tone={page.enabled ? "success" : "neutral"}>{page.enabled ? "Live" : "Private"}</Badge>
				</div>
				<div className={styles.sidebarScroll}>
					{addingBlock ? (
						<BlockComposer checks={options} elements={elements} loading={assignmentsLoading} onAdd={addBlock} onCancel={() => setAddingBlock(false)} />
					) : selectedElement ? (
						<ElementSettings element={selectedElement} elements={elements} checks={options} update={patch => updateElement(selectedElement.localId, patch)} onBack={() => setSelectedId(undefined)} />
					) : (
						<PageSettings page={page} update={updatePage} />
					)}
				</div>
			</aside>

			<div className={styles.main}>
				<StatusPageCanvas
					page={page}
					elements={elements}
					checks={options}
					selectedId={selectedId}
					draggingId={draggingId}
					onSelect={id => {
						setSelectedId(id);
						setAddingBlock(false);
					}}
					onAddGroup={addGroup}
					onAddBlock={startAddingBlock}
					onMove={moveElement}
					onRemove={removeElement}
					onReorderStart={startDragging}
				/>
				<UnsavedChangesBar className={styles.saveBar} show={hasChanges} saving={saving} onReset={reset} onSave={() => void save()} />
			</div>
		</div>
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

function scopedPreviewCSS(css: string | undefined) {
	if (!css?.trim()) return "";
	const scope = `.${styles.previewViewport}`;

	function findClosingBrace(source: string, openingBrace: number) {
		let depth = 0;
		let quote = "";
		let comment = false;
		for (let index = openingBrace; index < source.length; index += 1) {
			const character = source[index] || "";
			const next = source[index + 1] || "";
			if (comment) {
				if (character === "*" && next === "/") {
					comment = false;
					index += 1;
				}
				continue;
			}
			if (!quote && character === "/" && next === "*") {
				comment = true;
				index += 1;
				continue;
			}
			if (quote) {
				if (character === quote && source[index - 1] !== "\\") quote = "";
				continue;
			}
			if (character === '"' || character === "'") {
				quote = character;
				continue;
			}
			if (character === "{") depth += 1;
			if (character === "}") {
				depth -= 1;
				if (depth === 0) return index;
			}
		}
		return source.length - 1;
	}

	function prefixSelector(selector: string) {
		const trimmed = selector.trim();
		if (!trimmed) return "";
		if (trimmed.startsWith(scope)) return trimmed;
		if (trimmed.startsWith(".ns-status-page")) return trimmed.replace(/^\.ns-status-page/, scope);
		if (/^(?::root|html|body)(?:$|\s|\.|#|:|\[)/.test(trimmed)) {
			return trimmed.replace(/^(?::root|html|body)/, scope);
		}
		return `${scope} ${trimmed}`;
	}

	function scopeRules(source: string): string {
		let output = "";
		let cursor = 0;
		while (cursor < source.length) {
			const openingBrace = source.indexOf("{", cursor);
			if (openingBrace < 0) return output + source.slice(cursor);
			const headerStart = cursor;
			const header = source.slice(headerStart, openingBrace).trim();
			const leadingComments = header.match(/^(?:\/\*[\s\S]*?\*\/\s*)+/)?.[0] ?? "";
			const ruleHeader = header.slice(leadingComments.length).trim();
			const closingBrace = findClosingBrace(source, openingBrace);
			const body = source.slice(openingBrace + 1, closingBrace);
			if (!ruleHeader) {
				output += source.slice(cursor, closingBrace + 1);
			} else if (/^@(media|supports|container|layer)\b/i.test(ruleHeader)) {
				output += `${leadingComments}${ruleHeader} {${scopeRules(body)}}`;
			} else if (ruleHeader.startsWith("@")) {
				output += `${leadingComments}${ruleHeader} {${body}}`;
			} else {
				const selectors = ruleHeader.split(",").map(prefixSelector).filter(Boolean).join(", ");
				output += `${leadingComments}${selectors} {${body}}`;
			}
			cursor = closingBrace + 1;
		}
		return output;
	}

	return scopeRules(css);
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
				helper="New pages start with a fixed Picsum image URL. Replace it with any absolute HTTPS image URL."
				type="url"
				value={page.bannerImageUrl ?? ""}
				onChange={event => update("bannerImageUrl", event.currentTarget.value)}
			/>
			<SelectField label="Theme" value={page.theme} options={themeOptions} onChange={event => update("theme", event.currentTarget.value as PageDraft["theme"])} />

			<div className={styles.fieldGroup}>
				<div className={styles.fieldGroupHeading}>
					<strong>Public data</strong>
					<Badge tone="accent">Review exposure</Badge>
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

function BlockComposer({
	checks,
	elements,
	loading,
	onAdd,
	onCancel
}: {
	checks: CheckOption[];
	elements: ElementDraft[];
	loading: boolean;
	onAdd: (mode: DisplayMode, checkId: string, parentLocalId?: string) => void;
	onCancel: () => void;
}) {
	const groups = sorted(elements.filter(element => element.kind === "folder"));
	const [mode, setMode] = useState<DisplayMode>();
	const [category, setCategory] = useState("all");
	const [search, setSearch] = useState("");
	const [checkId, setCheckId] = useState("");
	const [parentLocalId, setParentLocalId] = useState(groups[0]?.localId ?? "");
	const categories = ["all", ...Array.from(new Set(checks.map(check => check.type))).sort((left, right) => left.localeCompare(right))];
	const normalizedSearch = search.trim().toLowerCase();
	const visibleChecks = checks.filter(
		check => (category === "all" || check.type === category) && (!normalizedSearch || `${check.name} ${check.target} ${check.type}`.toLowerCase().includes(normalizedSearch))
	);
	const selectedMode = blockLibrary.find(block => block.mode === mode);

	if (!mode) {
		return (
			<div className={styles.settingsSection}>
				<button type="button" className={styles.backButton} onClick={onCancel}>
					<ArrowLeftIcon aria-hidden="true" />
					Page settings
				</button>
				<div className={styles.sectionIntro}>
					<strong>Choose a display</strong>
					<p>Start with the information shape visitors should see.</p>
				</div>
				{blockLibrary.map(block => (
					<button key={block.mode} type="button" className={styles.libraryCard} onClick={() => setMode(block.mode)}>
						{block.icon}
						<span>
							<strong>{block.title}</strong>
							<small>{block.description}</small>
						</span>
						<ArrowRightIcon aria-hidden="true" />
					</button>
				))}
			</div>
		);
	}

	return (
		<div className={styles.settingsSection}>
			<button type="button" className={styles.backButton} onClick={() => setMode(undefined)}>
				<ArrowLeftIcon aria-hidden="true" />
				Display types
			</button>
			<div className={styles.composerSelection}>
				<span>{selectedMode?.icon}</span>
				<div>
					<strong>{selectedMode?.title}</strong>
					<small>{selectedMode?.description}</small>
				</div>
			</div>
			<div className={styles.sectionIntro}>
				<strong>Select a check</strong>
				<p>Browse by monitor type, then choose the public service source.</p>
			</div>
			{loading ? <Spinner label="Loading checks" layout="compact" size="sm" /> : null}
			{!loading && checks.length ? (
				<div className={styles.checkPicker}>
					<div className={styles.checkCategories} role="list" aria-label="Check categories">
						{categories.map(value => {
							const count = value === "all" ? checks.length : checks.filter(check => check.type === value).length;
							return (
								<button key={value} type="button" className={styles.checkCategory} data-selected={category === value} onClick={() => setCategory(value)}>
									<span>{checkCategoryLabel(value)}</span>
									<Badge tone={category === value ? "accent" : "neutral"}>{count}</Badge>
								</button>
							);
						})}
					</div>
					<div className={styles.checkChoices}>
						<TextField label="Search checks" placeholder="name or target" value={search} onChange={event => setSearch(event.currentTarget.value)} />
						<div className={styles.checkChoiceList} role="listbox" aria-label="Checks">
							{visibleChecks.map(check => (
								<button
									key={check.id}
									type="button"
									className={styles.checkChoice}
									role="option"
									aria-selected={checkId === check.id}
									data-selected={checkId === check.id}
									onClick={() => setCheckId(check.id)}
								>
									<strong>{check.name}</strong>
									<span>{check.target}</span>
								</button>
							))}
							{!visibleChecks.length ? <p className={styles.inlineNotice}>No checks match this category and search.</p> : null}
						</div>
					</div>
				</div>
			) : null}
			{!loading && !checks.length ? <p className={styles.inlineNotice}>Create a check and assign a probe before adding a status block.</p> : null}
			<SelectField
				label="Group"
				value={parentLocalId}
				options={[{ value: "", label: "No group" }, ...groups.map(group => ({ value: group.localId, label: group.title || "Untitled group" }))]}
				onChange={event => setParentLocalId(event.currentTarget.value)}
			/>
			<div className={styles.composerActions}>
				<Button type="button" variant="ghost" onClick={onCancel}>
					Cancel
				</Button>
				<Button type="button" disabled={!checkId} onClick={() => onAdd(mode, checkId, parentLocalId || undefined)}>
					<PlusIcon aria-hidden="true" />
					Add block
				</Button>
			</div>
		</div>
	);
}

function checkCategoryLabel(value: string) {
	switch (value.toLowerCase()) {
		case "all":
			return "All";
		case "http":
			return "HTTP";
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Trace";
		default:
			return value;
	}
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
				Page settings
			</button>
			<div className={styles.sectionIntro}>
				<strong>{element.kind === "folder" ? "Group settings" : "Block settings"}</strong>
				<p>{element.kind === "folder" ? "Set the public heading and description for this service group." : "Choose the service source, presentation, and public label."}</p>
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
	draggingId,
	onSelect,
	onAddGroup,
	onAddBlock,
	onMove,
	onRemove,
	onReorderStart
}: {
	page: PageDraft;
	elements: ElementDraft[];
	checks: CheckOption[];
	selectedId?: string;
	draggingId?: string;
	onSelect: (id: string) => void;
	onAddGroup: () => void;
	onAddBlock: () => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onReorderStart: (event: PointerEvent<HTMLElement>, id: string) => void;
}) {
	const groups = sorted(elements.filter(element => element.kind === "folder"));
	const ungrouped = sorted(elements.filter(element => element.kind === "assignment_group" && !element.parentLocalId));
	const previewTheme = page.theme === "auto" ? "dark" : page.theme;
	const previewCSS = useMemo(() => scopedPreviewCSS(page.customCss), [page.customCss]);

	return (
		<section className={styles.canvas} aria-label="Status page preview">
			<div className={`${styles.previewViewport} ns-status-page`} data-preview-theme={previewTheme} data-status-preview>
				{previewCSS ? <style>{previewCSS}</style> : null}
				<div className={styles.publicShell}>
					<header className={`${styles.previewHero} ns-status-hero`}>
						{page.bannerImageUrl ? (
							<img className={`${styles.previewBanner} ns-status-banner`} src={page.bannerImageUrl} alt="" />
						) : (
							<div className={`${styles.previewBanner} ns-status-banner`} aria-hidden="true" />
						)}
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

					<div className={`${styles.previewOverall} ns-status-overall`}>
						<span className={styles.previewStatusMarker} aria-hidden="true" />
						<div>
							<strong>All systems operational</strong>
							<small>No active incidents reported</small>
						</div>
						<Badge className={styles.previewOverallBadge} tone="success">
							Operational
						</Badge>
					</div>
					<div className={styles.previewAddActions}>
						<Button type="button" variant="outline" size="sm" onClick={onAddGroup}>
							<FolderIcon aria-hidden="true" />
							Add Group
						</Button>
						<Button type="button" size="sm" disabled={!checks.length} onClick={onAddBlock}>
							<PlusIcon aria-hidden="true" />
							Add Block
						</Button>
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
							draggingId={draggingId}
							onSelect={onSelect}
							onMove={onMove}
							onRemove={onRemove}
							onReorderStart={onReorderStart}
						/>
					))}

					{ungrouped.length ? (
						<section className={styles.previewGroup} data-builder-drop-parent="">
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
										check={checkForElement(element, checks)}
										page={page}
										selected={element.localId === selectedId}
										dragging={element.localId === draggingId}
										first={index === 0}
										last={index === ungrouped.length - 1}
										onSelect={onSelect}
										onMove={onMove}
										onRemove={onRemove}
										onReorderStart={onReorderStart}
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
						</div>
					) : null}

					<footer className={`${styles.previewFooter} ns-status-footer`}>
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
	draggingId,
	onSelect,
	onMove,
	onRemove,
	onReorderStart
}: {
	group: ElementDraft;
	first: boolean;
	last: boolean;
	blocks: ElementDraft[];
	checks: CheckOption[];
	page: PageDraft;
	selectedId?: string;
	draggingId?: string;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onReorderStart: (event: PointerEvent<HTMLElement>, id: string) => void;
}) {
	return (
		<section
			className={`${styles.previewGroup} ns-status-group ${group.localId === selectedId ? styles.selectedElement : ""} ${group.localId === draggingId ? styles.draggingElement : ""}`}
			data-builder-sort-id={group.localId}
			data-builder-parent-id=""
		>
			<div className={styles.previewGroupHeader} data-builder-drop-parent={group.localId}>
				<div className={styles.previewGroupIdentity}>
					<DragHandle label={`Drag ${group.title || "group"}`} onReorderStart={event => onReorderStart(event, group.localId)} />
					<div>
						<h2>{group.title || "Untitled group"}</h2>
						{group.description ? <p>{group.description}</p> : null}
					</div>
				</div>
				<ElementControls element={group} first={first} last={last} onSelect={onSelect} onMove={onMove} onRemove={onRemove} />
			</div>
			<div className={styles.previewGroupBody}>
				{blocks.length ? (
					blocks.map((element, index) => (
						<PreviewBlock
							key={element.localId}
							element={element}
							check={checkForElement(element, checks)}
							page={page}
							selected={element.localId === selectedId}
							dragging={element.localId === draggingId}
							first={index === 0}
							last={index === blocks.length - 1}
							onSelect={onSelect}
							onMove={onMove}
							onRemove={onRemove}
							onReorderStart={onReorderStart}
						/>
					))
				) : (
					<div className={styles.groupDropZone} data-builder-drop-parent={group.localId}>
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
	dragging,
	first,
	last,
	onSelect,
	onMove,
	onRemove,
	onReorderStart
}: {
	element: ElementDraft;
	check?: CheckOption;
	page: PageDraft;
	selected: boolean;
	dragging: boolean;
	first: boolean;
	last: boolean;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
	onReorderStart: (event: PointerEvent<HTMLElement>, id: string) => void;
}) {
	const probeNames = summarizeValues(check?.viewpoints.map(viewpoint => viewpoint.name) ?? []);
	const probeLocations = summarizeValues(check?.viewpoints.flatMap(viewpoint => (viewpoint.locationName ? [viewpoint.locationName] : [])) ?? []);
	const metadata = [check?.type, page.showTargets ? check?.target : undefined, page.showProbeNames ? probeNames : undefined, page.showProbeLocations ? probeLocations : undefined].filter(Boolean);
	const mapPoints = (check?.viewpoints ?? []).flatMap(viewpoint => {
		if (typeof viewpoint.latitude !== "number" || typeof viewpoint.longitude !== "number") return [];
		return [
			{
				...viewpoint,
				left: Math.min(94, Math.max(6, ((viewpoint.longitude + 180) / 360) * 100)),
				top: Math.min(88, Math.max(12, ((90 - viewpoint.latitude) / 180) * 100))
			}
		];
	});

	return (
		<article
			className={`${styles.previewBlock} ns-status-block ${selected ? styles.selectedElement : ""} ${dragging ? styles.draggingElement : ""}`}
			data-builder-sort-id={element.localId}
			data-builder-parent-id={element.parentLocalId ?? ""}
		>
			<div className={styles.previewBlockTop}>
				<div className={styles.previewBlockIdentity}>
					<DragHandle label={`Drag ${element.title || check?.name || "block"}`} onReorderStart={event => onReorderStart(event, element.localId)} />
					<span className={styles.operationalDot} aria-hidden="true" />
					<div>
						<strong>{element.title || check?.name || "Untitled service"}</strong>
						<small>{metadata.length ? metadata.join(" / ") : displayLabel(element.displayMode)}</small>
					</div>
				</div>
				<div className={styles.previewBlockActions}>
					<span>Operational</span>
					<ElementControls element={element} label={element.title || check?.name} first={first} last={last} onSelect={onSelect} onMove={onMove} onRemove={onRemove} />
				</div>
			</div>
			<div className={styles.uptimeHeading}>
				<span>30-day uptime</span>
				<strong>{element.displayMode === "history" ? "99.94%" : "99.98%"}</strong>
			</div>
			<div className={styles.uptimeBars} aria-label="Example 30-day uptime">
				{Array.from({ length: 30 }, (_, index) => (
					<span key={index} data-state={index === 11 && element.displayMode === "history" ? "degraded" : "operational"} />
				))}
			</div>
			{element.displayMode === "status" ? (
				<div className={styles.previewMetrics}>
					<PreviewMetric label="Availability" value="99.98%" />
					<PreviewMetric label="Median latency" value="28 ms" />
					<PreviewMetric label="Viewpoints" value={`${check?.viewpoints.length ?? 0} active`} />
				</div>
			) : null}
			{element.displayMode === "history" ? (
				<div className={styles.historySummary}>
					<span className={styles.operationalDot} aria-hidden="true" />
					<div>
						<strong>Last incident resolved</strong>
						<small>Intermittent latency recovered after 18 minutes.</small>
					</div>
					<Badge tone="success">Resolved</Badge>
				</div>
			) : null}
			{element.displayMode === "latency" ? (
				<>
					<div className={styles.previewMetrics}>
						<PreviewMetric label="Current" value="31 ms" />
						<PreviewMetric label="P95" value="46 ms" />
						<PreviewMetric label="Packet loss" value="0.02%" />
					</div>
					<svg className={styles.miniChart} viewBox="0 0 640 72" preserveAspectRatio="none" role="img" aria-label="Example latency trend">
						<path d="M0 52 C70 48 92 58 148 42 S242 35 300 44 S390 20 452 30 S552 42 640 16" />
					</svg>
				</>
			) : null}
			{element.displayMode === "map" && page.showProbeLocations ? (
				<div className={styles.miniMap} aria-label="Configured public probe locations">
					<svg viewBox="0 0 100 48" preserveAspectRatio="none" aria-hidden="true">
						<path d="M0 12H100M0 24H100M0 36H100M20 0V48M40 0V48M60 0V48M80 0V48" />
					</svg>
					{mapPoints.map(point => (
						<span
							key={point.id}
							style={{ left: `${point.left}%`, top: `${point.top}%` }}
							title={[point.name, point.locationName].filter(Boolean).join(" / ")}
							role="img"
							aria-label={[point.name, point.locationName].filter(Boolean).join(" / ")}
						/>
					))}
					<small>{mapPoints.length ? `${mapPoints.length} configured monitoring ${mapPoints.length === 1 ? "location" : "locations"}` : "No probe coordinates configured"}</small>
				</div>
			) : null}
			{element.displayMode === "map" && !page.showProbeLocations ? <div className={styles.mapPrivacyNotice}>Enable public probe locations to render this map.</div> : null}
		</article>
	);
}

function summarizeValues(values: string[]) {
	const unique = [...new Set(values.filter(Boolean))];
	if (unique.length <= 2) return unique.join(", ");
	return `${unique.slice(0, 2).join(", ")} +${unique.length - 2}`;
}

function PreviewMetric({ label, value }: { label: string; value: string }) {
	return (
		<div>
			<span>{label}</span>
			<strong>{value}</strong>
		</div>
	);
}

function DragHandle({ label, onReorderStart }: { label: string; onReorderStart: (event: PointerEvent<HTMLButtonElement>) => void }) {
	return (
		<button type="button" className={styles.dragHandle} aria-label={label} onPointerDown={onReorderStart}>
			<DotsSixVerticalIcon aria-hidden="true" />
		</button>
	);
}

function ElementControls({
	element,
	label,
	first,
	last,
	onSelect,
	onMove,
	onRemove
}: {
	element: ElementDraft;
	label?: string;
	first: boolean;
	last: boolean;
	onSelect: (id: string) => void;
	onMove: (id: string, direction: -1 | 1) => void;
	onRemove: (id: string) => void;
}) {
	const accessibleLabel = label || element.title || "element";
	return (
		<div className={styles.elementControls}>
			<IconButton aria-label={`Move ${accessibleLabel} up`} disabled={first} onClick={() => onMove(element.localId, -1)}>
				<ArrowUpIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Move ${accessibleLabel} down`} disabled={last} onClick={() => onMove(element.localId, 1)}>
				<ArrowDownIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Configure ${accessibleLabel}`} onClick={() => onSelect(element.localId)}>
				<GearSixIcon aria-hidden="true" />
			</IconButton>
			<IconButton aria-label={`Remove ${accessibleLabel}`} danger onClick={() => onRemove(element.localId)}>
				<TrashIcon aria-hidden="true" />
			</IconButton>
		</div>
	);
}

function displayLabel(mode: DisplayMode) {
	return displayOptions.find(option => option.value === mode)?.label ?? "Service status";
}
