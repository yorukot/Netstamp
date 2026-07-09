import { dismissToast, subscribeToasts, type ToastMessage } from "@/shared/toast/toastStore";
import { Toast, ToastClose, ToastDescription, ToastTitle, ToastViewport } from "@netstamp/ui";
import { useEffect, useState } from "react";

export function ToastProvider() {
	const [messages, setMessages] = useState<ToastMessage[]>([]);

	useEffect(() => subscribeToasts(setMessages), []);

	if (!messages.length) {
		return null;
	}

	return (
		<ToastViewport>
			{messages.map(message => (
				<Toast tone={message.tone} key={message.id}>
					<div>
						<ToastTitle>{message.title}</ToastTitle>
						<ToastDescription>{message.message}</ToastDescription>
					</div>
					<ToastClose ariaLabel="Close notification" onClick={() => dismissToast(message.id)} />
				</Toast>
			))}
		</ToastViewport>
	);
}
