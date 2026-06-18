import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiAssignments } from "@/features/checks/api/resultAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { pathForProbeDetail, pathForRoute } from "@/routes/routePaths";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { NetworkMap } from "@/shared/components/NetworkMap";
import type { AssignedRow } from "@/shared/domain/assignments";
import { classNames } from "@/shared/utils/classNames";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Outlet, useNavigate, useParams } from "react-router-dom";
import { ProbeDetail } from "./ProbeDetail";
import { ProbeList } from "./ProbeList";
import { ProbePageHeader } from "./ProbePageHeader";
import styles from "./ProbesPage.module.css";
import { filterProbes } from "./probeUtils";
import type { ProbeSort, ProbeView } from "./types";

export function ProbesPage() {
	const { projectRef } = useCurrentProject();
	const { probeId = "" } = useParams();
	const navigate = useNavigate();
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		refetchInterval: 2000,
		select: data => mapApiProbes(data.probes)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiChecks(data.checks, probesQuery.data)
	});
	const assignmentsQuery = useQuery({
		...projectQueries.assignments(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => data.assignments
	});
	const probes = probesQuery.data || [];
	const checks = checksQuery.data || [];
	const [view, setView] = useState<ProbeView>("grid");
	const [search, setSearch] = useState("");
	const [sortKey, setSortKey] = useState<ProbeSort>("heartbeat");
	const selectedProbe = probes.find(probe => probe.id === probeId) || null;
	const selectedProbeId = selectedProbe?.id || "";
	const visibleProbes = filterProbes(probes, search, sortKey);
	const assignedRows: AssignedRow[] = mapApiAssignments(assignmentsQuery.data, probes, checks);

	useEffect(() => {
		if (!projectRef || !probeId || probesQuery.isPending || probesQuery.isError || selectedProbe) {
			return;
		}

		navigate(pathForRoute("probes", { projectRef }), { replace: true });
	}, [navigate, probeId, probesQuery.isError, probesQuery.isPending, projectRef, selectedProbe]);

	function selectProbe(nextProbeId: string) {
		if (!projectRef) {
			return;
		}

		navigate(pathForProbeDetail(projectRef, nextProbeId));
	}

	function closeProbeDetail() {
		if (!projectRef) {
			return;
		}

		navigate(pathForRoute("probes", { projectRef }));
	}

	return (
		<section className={classNames(styles.screen, view === "map" && styles.mapScreen)}>
			{view === "grid" ? (
				<>
					<ProbePageHeader view={view} projectRef={projectRef} onViewChange={setView} />
					<div className={styles.gridLayout}>
						<ProbeList probes={visibleProbes} selectedId={selectedProbeId} search={search} sortKey={sortKey} onSearchChange={setSearch} onSortChange={setSortKey} onSelect={selectProbe} />
					</div>
				</>
			) : (
				<div className={styles.mapView}>
					<NetworkMap probes={probes} selectedId={selectedProbeId} onSelect={selectProbe} mode="fleet" className={styles.fullMap} />
					<ProbePageHeader view={view} projectRef={projectRef} onViewChange={setView} overlay />
				</div>
			)}

			<EditorDrawer open={Boolean(selectedProbe)} title={selectedProbe?.name || "Probe"} ariaLabel="Probe detail" backLabel="back to probes" onClose={closeProbeDetail}>
				{selectedProbe ? <ProbeDetail key={selectedProbe.id} probe={selectedProbe} assignedRows={assignedRows} projectRef={projectRef} onDeleted={closeProbeDetail} /> : null}
			</EditorDrawer>

			<Outlet />
		</section>
	);
}
