import { dismissToast, subscribeToasts, type ToastMessage } from "@/shared/toast/toastStore";
import { useEffect, useState } from "react";
import { CloseButton } from "./CloseButton";
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
					<CloseButton className={styles.toastClose} ariaLabel="Close notification" onClick={() => dismissToast(message.id)} />
				</div>
			))}
		</div>
	);
}
