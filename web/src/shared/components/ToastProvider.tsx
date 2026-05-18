import { dismissToast, subscribeToasts, type ToastMessage } from "@/shared/toast/toastStore";
import { Button } from "@netstamp/ui";
import { useEffect, useState } from "react";
import styles from "./ToastProvider.module.css";

export function ToastProvider() {
	const [messages, setMessages] = useState<ToastMessage[]>([]);

	useEffect(() => subscribeToasts(setMessages), []);

	if (!messages.length) {
		return null;
	}

	return (
		<div className={styles.viewport} role="region" aria-label="Notifications">
			{messages.map(message => (
				<div className={styles.toast} data-tone={message.tone} role="status" key={message.id}>
					<div>
						<strong>{message.title}</strong>
						<p>{message.message}</p>
					</div>
					<Button variant="ghost" size="sm" onClick={() => dismissToast(message.id)}>
						Close
					</Button>
				</div>
			))}
		</div>
	);
}
