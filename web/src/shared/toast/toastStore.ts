import type { ToastTone } from "@netstamp/ui";
export type { ToastTone } from "@netstamp/ui";

export interface ToastMessage {
	id: string;
	message: string;
	tone: ToastTone;
	title: string;
}

type ToastListener = (messages: ToastMessage[]) => void;

const listeners = new Set<ToastListener>();
let messages: ToastMessage[] = [];
let nextToastId = 0;

function emit() {
	for (const listener of listeners) {
		listener(messages);
	}
}

export function subscribeToasts(listener: ToastListener) {
	listeners.add(listener);
	listener(messages);

	return () => {
		listeners.delete(listener);
	};
}

export function dismissToast(id: string) {
	messages = messages.filter(message => message.id !== id);
	emit();
}

export function pushToast(message: Omit<ToastMessage, "id">) {
	const id = String(nextToastId++);
	messages = [...messages, { ...message, id }].slice(-4);
	emit();
	window.setTimeout(() => dismissToast(id), 5200);
}

export function pushErrorToast(message: string) {
	pushToast({
		message,
		title: "Request failed",
		tone: "critical"
	});
}
