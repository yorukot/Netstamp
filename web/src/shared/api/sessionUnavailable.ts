export type SessionUnavailableCode = "AUTH_MISSING_SESSION" | "AUTH_INVALID_SESSION";

type SessionUnavailableListener = (code: SessionUnavailableCode) => void;

const listeners = new Set<SessionUnavailableListener>();

export function isSessionUnavailableCode(code: string | undefined): code is SessionUnavailableCode {
	return code === "AUTH_MISSING_SESSION" || code === "AUTH_INVALID_SESSION";
}

export function reportSessionUnavailable(code: SessionUnavailableCode) {
	for (const listener of listeners) {
		listener(code);
	}
}

export function subscribeToSessionUnavailable(listener: SessionUnavailableListener) {
	listeners.add(listener);

	return () => {
		listeners.delete(listener);
	};
}
