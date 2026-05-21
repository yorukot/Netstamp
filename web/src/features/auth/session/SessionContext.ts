import { createContext, useContext } from "react";
import type { AuthCredentials, ProjectDraft, RegisterPayload, SessionSnapshot } from "../services/authService";

export interface SessionContextValue {
	session: SessionSnapshot | null;
	loading: boolean;
	submitting: boolean;
	isAuthenticated: boolean;
	login: (payload: AuthCredentials) => Promise<SessionSnapshot["user"]>;
	register: (payload: RegisterPayload) => Promise<SessionSnapshot["user"]>;
	createProject: (payload: ProjectDraft) => Promise<NonNullable<SessionSnapshot["project"]>>;
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
