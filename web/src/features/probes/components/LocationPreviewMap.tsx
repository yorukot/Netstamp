import { formatCoordinate, type ProbeCoordinates } from "@/features/probes/data/probeLocation";
import { type Probe } from "@/features/probes/data/probes";
import { NetworkMap } from "@/shared/components/NetworkMap";
import { classNames } from "@/shared/utils/classNames";
import styles from "./LocationPreviewMap.module.css";

interface LocationPreviewMapProps {
	coordinates: ProbeCoordinates;
	locationName: string;
	probeName: string;
	className?: string;
}

export function LocationPreviewMap({ coordinates, locationName, probeName, className }: LocationPreviewMapProps) {
	const previewProbe: Probe = {
		id: "location-preview",
		name: probeName || locationName || "probe",
		status: "Offline",
		location: locationName || `${formatCoordinate(coordinates.latitude)}, ${formatCoordinate(coordinates.longitude)}`,
		publicIp: "-",
		asn: "-",
		provider: "-",
		region: locationName || "preview",
		ipFamily: "-",
		lastHeartbeat: "never",
		tags: [],
		version: "-",
		uptime: "-",
		cpu: "-",
		memory: "-",
		queue: "-",
		loss: "-",
		coordinates: [coordinates.longitude, coordinates.latitude],
		capabilities: []
	};

	return (
		<div className={classNames(styles.preview, className)}>
			<NetworkMap probes={[previewProbe]} selectedId={previewProbe.id} mode="detail" className={styles.map} />
			<div className={styles.meta}>
				<span>Location preview</span>
				<strong>{locationName || previewProbe.location}</strong>
				<code>
					{formatCoordinate(coordinates.latitude)}, {formatCoordinate(coordinates.longitude)}
				</code>
			</div>
		</div>
	);
}
