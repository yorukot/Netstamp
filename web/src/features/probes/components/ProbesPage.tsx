import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type ProbeStatus } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { NetworkMap } from "@/shared/components/NetworkMap";
import { classNames } from "@/shared/utils/classNames";
import { Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Outlet } from "react-router-dom";
import { ProbeDetail } from "./ProbeDetail";
import { ProbeList } from "./ProbeList";
import { ProbePageHeader } from "./ProbePageHeader";
import styles from "./ProbesPage.module.css";
import { filterProbes } from "./probeUtils";
import type { AssignedRow, ProbeSort, ProbeView } from "./types";

export function ProbesPage() {
	const { projectRef } = useCurrentProject();
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiChecks(data.checks, probesQuery.data)
	});
	const probes = probesQuery.data || [];
	const providerOptions = Array.from(new Set(probes.map(probe => probe.provider)));
	const [view, setView] = useState<ProbeView>("grid");
	const [selectedId, setSelectedId] = useState("");
	const [search, setSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState<"all" | ProbeStatus>("all");
	const [providerFilter, setProviderFilter] = useState("all");
	const [sortKey, setSortKey] = useState<ProbeSort>("heartbeat");
	const selectedProbe = probes.find(probe => probe.id === selectedId) || probes[0] || null;
	const selectedProbeId = selectedProbe?.id || "";
	const visibleProbes = filterProbes(probes, search, statusFilter, providerFilter, sortKey);
	const assignedRows: AssignedRow[] = (checksQuery.data ?? []).flatMap(check =>
		probes.map(probe => ({
			probe: probe.name,
			check: check.name,
			type: check.type,
			interval: check.interval,
			jitter: check.jitter,
			latest: check.latest
		}))
	);

	return (
		<section className={classNames(styles.screen, view === "map" && styles.mapScreen)}>
			{view === "grid" ? (
				<>
					<ProbePageHeader view={view} onViewChange={setView} />
					<div className={styles.gridLayout}>
						<ProbeList
							probes={visibleProbes}
							providerOptions={providerOptions}
							selectedId={selectedProbeId}
							search={search}
							statusFilter={statusFilter}
							providerFilter={providerFilter}
							sortKey={sortKey}
							onSearchChange={setSearch}
							onStatusChange={setStatusFilter}
							onProviderChange={setProviderFilter}
							onSortChange={setSortKey}
							onSelect={setSelectedId}
						/>
						<div className={styles.lowerGrid}>
							<NetworkMap probes={probes} selectedId={selectedProbeId} onSelect={setSelectedId} mode="detail" className={styles.miniMap} />
							{selectedProbe ? (
								<ProbeDetail key={selectedProbe.id} probe={selectedProbe} assignedRows={assignedRows} projectRef={projectRef} onDeleted={() => setSelectedId("")} />
							) : (
								<Panel tone="matte" eyebrow="Probe detail" title={projectRef ? "No probes yet" : "No project selected"} />
							)}
						</div>
					</div>
				</>
			) : (
				<div className={styles.mapView}>
					<NetworkMap probes={probes} selectedId={selectedProbeId} onSelect={setSelectedId} mode="fleet" className={styles.fullMap} />
					<ProbePageHeader view={view} onViewChange={setView} overlay />
					{selectedProbe ? <ProbeDetail key={selectedProbe.id} probe={selectedProbe} assignedRows={assignedRows} projectRef={projectRef} onDeleted={() => setSelectedId("")} floating /> : null}
				</div>
			)}

			<Outlet />
		</section>
	);
}
