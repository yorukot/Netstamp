import { formatCoordinate, type ProbeCoordinates } from "@/features/probes/data/probeLocation";
import { classNames } from "@/shared/utils/classNames";
import { NetworkMap, type NetworkMapMarker } from "@/shared/visualizations/NetworkMap";
import { useTranslation } from "react-i18next";
import styles from "./LocationPreviewMap.module.css";

interface LocationPreviewMapProps {
	coordinates: ProbeCoordinates;
	locationName: string;
	probeName: string;
	className?: string;
}

export function LocationPreviewMap({ coordinates, locationName, probeName, className }: LocationPreviewMapProps) {
	const { t } = useTranslation("probes");
	const fallbackLocation = `${formatCoordinate(coordinates.latitude)}, ${formatCoordinate(coordinates.longitude)}`;
	const previewProbe: NetworkMapMarker = {
		id: "location-preview",
		name: probeName || locationName || t("probe"),
		coordinates: [coordinates.longitude, coordinates.latitude]
	};

	return (
		<div className={classNames(styles.preview, className)}>
			<NetworkMap probes={[previewProbe]} selectedId={previewProbe.id} mode="detail" className={styles.map} />
			<div className={styles.meta}>
				<span>{t("location.preview")}</span>
				<strong>{locationName || fallbackLocation}</strong>
				<code>
					{formatCoordinate(coordinates.latitude)}, {formatCoordinate(coordinates.longitude)}
				</code>
			</div>
		</div>
	);
}
