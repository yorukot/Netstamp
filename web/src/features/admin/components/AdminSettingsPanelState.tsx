import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { BodyCopy, Button, Spinner } from "@netstamp/ui";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import styles from "./AdminSettingsPanels.module.css";

export type SettingsNoticeTone = "critical" | "success";

export const SettingsNotice = ({ children, tone }: { children: ReactNode; tone: SettingsNoticeTone }) => (
	<div className={styles.notice} data-tone={tone} role={tone === "critical" ? "alert" : "status"}>
		{children}
	</div>
);

export const SettingsPanelLoading = () => {
	const { t } = useTranslation("admin");
	return <Spinner label={t("loading")} layout="panel" size="lg" />;
};

export const SettingsPanelError = ({ error, onRetry }: { error: unknown; onRetry: () => void }) => {
	const { t } = useTranslation(["admin", "common"]);

	return (
		<div className={styles.sectionState}>
			<BodyCopy>{requestErrorMessage(error, t("settings.sectionLoadError"))}</BodyCopy>
			<div>
				<Button type="button" size="sm" variant="outline" onClick={onRetry}>
					{t("common:actions.retry")}
				</Button>
			</div>
		</div>
	);
};
