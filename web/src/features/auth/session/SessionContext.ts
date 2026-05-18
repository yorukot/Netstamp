import { createContext, useContext } from "react";
import type { AuthCredentials, RegisterPayload, SessionSnapshot, TeamDraft } from "../services/authService";

export interface SessionContextValue {
	session: SessionSnapshot | null;
	loading: boolean;
	submitting: boolean;
	isAuthenticated: boolean;
	login: (payload: AuthCredentials) => Promise<SessionSnapshot["user"]>;
	register: (payload: RegisterPayload) => Promise<SessionSnapshot["user"]>;
	createTeam: (payload: TeamDraft) => Promise<NonNullable<SessionSnapshot["team"]>>;
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
