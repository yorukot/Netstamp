import type { AuthCredentials, RegisterPayload, RegisterResult, SessionSnapshot } from "@/features/auth/services/authService";
import { createContext, useContext } from "react";

export interface SessionContextValue {
	session: SessionSnapshot | null;
	loading: boolean;
	submitting: boolean;
	isAuthenticated: boolean;
	login: (payload: AuthCredentials) => Promise<SessionSnapshot["user"]>;
	register: (payload: RegisterPayload) => Promise<RegisterResult>;
	logout: () => void;
}

export const SessionContext = createContext<SessionContextValue | null>(null);

export function useSession() {
	const value = useContext(SessionContext);

	if (!value) {
		throw new Error("useSession must be used inside SessionProvider");
	}

	return value;
}
