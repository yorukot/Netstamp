import { assignments } from "@/features/checks/data/checks";
import { probes, type ProbeStatus } from "@/features/probes/data/probes";
import { NetworkMap } from "@/shared/components/NetworkMap";
import { classNames } from "@/shared/utils/classNames";
import { useState } from "react";
import { Outlet } from "react-router-dom";
import { ProbeDetail } from "./ProbeDetail";
import { ProbeList } from "./ProbeList";
import { ProbePageHeader } from "./ProbePageHeader";
import styles from "./ProbesPage.module.css";
import { filterProbes } from "./probeUtils";
import type { AssignedRow, ProbeSort, ProbeView } from "./types";

const providerOptions = Array.from(new Set(probes.map(probe => probe.provider)));

export function ProbesPage() {
	const [view, setView] = useState<ProbeView>("grid");
	const [selectedId, setSelectedId] = useState("ams-edge-01");
	const [search, setSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState<"all" | ProbeStatus>("all");
	const [providerFilter, setProviderFilter] = useState("all");
	const [sortKey, setSortKey] = useState<ProbeSort>("heartbeat");
	const selectedProbe = probes.find(probe => probe.id === selectedId) || probes[0];
	const visibleProbes = filterProbes(probes, search, statusFilter, providerFilter, sortKey);
	const assignedRows: AssignedRow[] = assignments.map(([probe, check, type, interval, jitter, latest]) => ({
		probe,
		check,
		type,
		interval,
		jitter,
		latest
	}));

	return (
		<section className={classNames(styles.screen, view === "map" && styles.mapScreen)}>
			{view === "grid" ? (
				<>
					<ProbePageHeader view={view} onViewChange={setView} />
					<div className={styles.gridLayout}>
						<ProbeList
							probes={visibleProbes}
							providerOptions={providerOptions}
							selectedId={selectedId}
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
							<NetworkMap probes={probes} selectedId={selectedId} onSelect={setSelectedId} mode="detail" className={styles.miniMap} />
							<ProbeDetail key={selectedProbe.id} probe={selectedProbe} assignedRows={assignedRows} />
						</div>
					</div>
				</>
			) : (
				<div className={styles.mapView}>
					<NetworkMap probes={probes} selectedId={selectedId} onSelect={setSelectedId} mode="fleet" className={styles.fullMap} />
					<ProbePageHeader view={view} onViewChange={setView} overlay />
					<ProbeDetail key={selectedProbe.id} probe={selectedProbe} assignedRows={assignedRows} floating />
				</div>
			)}

			<Outlet />
		</section>
	);
}
