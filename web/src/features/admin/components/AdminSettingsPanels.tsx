import { AdminAccessSettingsPanel } from "./AdminAccessSettingsPanel";
import { AdminAuthenticationProviderPanels } from "./AdminAuthenticationProviderPanels";
import { AdminSMTPSettingsPanel } from "./AdminSMTPSettingsPanel";
import styles from "./AdminSettingsPanels.module.css";

export const AdminSettingsPanels = () => (
	<div className={styles.settingsStack}>
		<AdminAccessSettingsPanel />
		<AdminSMTPSettingsPanel />
		<AdminAuthenticationProviderPanels />
	</div>
);
