import type { ReactNode } from "react";
import { useState } from "react";
import { type AuthCredentials, type SessionSnapshot, type TeamDraft, getSessionSnapshot, mockCreateTeam, mockLogin, mockRegister } from "../services/authService";
import { SessionContext } from "./SessionContext";

interface SessionProviderProps {
	children: ReactNode;
}

export function SessionProvider({ children }: SessionProviderProps) {
	const [session, setSession] = useState<SessionSnapshot | null>(null);
	const [submitting, setSubmitting] = useState(false);

	async function login(payload: AuthCredentials) {
		setSubmitting(true);

		try {
			const user = await mockLogin(payload);
			setSession({ user, controller: "waiting-for-api" });
			return user;
		} finally {
			setSubmitting(false);
		}
	}

	async function register(payload: AuthCredentials) {
		setSubmitting(true);

		try {
			const user = await mockRegister(payload);
			setSession({ user, controller: "waiting-for-api" });
			return user;
		} finally {
			setSubmitting(false);
		}
	}

	async function createTeam(payload: TeamDraft) {
		setSubmitting(true);

		try {
			const team = await mockCreateTeam(payload);
			setSession(current => {
				const baseSession = current ?? getSessionSnapshot();

				return {
					...baseSession,
					team,
					user: { ...baseSession.user, onboardingRequired: false }
				};
			});
			return team;
		} finally {
			setSubmitting(false);
		}
	}

	function logout() {
		setSession(null);
	}

	return <SessionContext value={{ session, submitting, isAuthenticated: Boolean(session), login, register, createTeam, logout }}>{children}</SessionContext>;
}
