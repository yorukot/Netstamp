import { AdminAccessSettingsPanel } from "./AdminAccessSettingsPanel";
import { AdminAuthenticationProviderPanels } from "./AdminAuthenticationProviderPanels";
import { AdminSMTPSettingsPanel } from "./AdminSMTPSettingsPanel";
import { AdminSystemInformationPanel } from "./AdminSystemInformationPanel";
import styles from "./AdminSettingsPanels.module.css";

export const AdminSettingsPanels = () => (
	<div className={styles.settingsStack}>
		<AdminSystemInformationPanel />
		<AdminAccessSettingsPanel />
		<AdminSMTPSettingsPanel />
		<AdminAuthenticationProviderPanels />
	</div>
);
