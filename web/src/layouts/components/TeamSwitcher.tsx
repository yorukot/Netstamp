import { Select } from "@netstamp/ui";
import styles from "../AppShell.module.css";

export function TeamSwitcher() {
	return (
		<label className={styles.teamSelect}>
			<span>team</span>
			<Select variant="compact" frameClassName={styles.teamFrame} className={styles.teamControl} defaultValue="vector-ix">
				<option value="vector-ix">Vector IX / prod</option>
				<option value="helio">Helio Validators</option>
				<option value="lab">Lab Network</option>
			</Select>
		</label>
	);
}
