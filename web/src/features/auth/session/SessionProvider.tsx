import { useCreateProjectMutation, useLoginMutation, useLogoutMutation, useRegisterMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { type AuthCredentials, type RegisterPayload, type TeamDraft, createSessionSnapshot, mapApiUser, mapProjectTeam } from "../services/authService";
import { SessionContext } from "./SessionContext";

interface SessionProviderProps {
	children: ReactNode;
}

export function SessionProvider({ children }: SessionProviderProps) {
	const meQuery = useQuery(authQueries.me());
	const session = meQuery.data ? createSessionSnapshot(meQuery.data.user) : null;
	const loginMutation = useLoginMutation();
	const registerMutation = useRegisterMutation();
	const createTeamMutation = useCreateProjectMutation();
	const logoutMutation = useLogoutMutation();
	const submitting = loginMutation.isPending || registerMutation.isPending || createTeamMutation.isPending || logoutMutation.isPending;

	async function login(payload: AuthCredentials) {
		const result = await loginMutation.mutateAsync(payload);
		return mapApiUser(result.user);
	}

	async function register(payload: RegisterPayload) {
		const result = await registerMutation.mutateAsync(payload);
		return mapApiUser(result.user, { onboardingRequired: true });
	}

	async function createTeam(payload: TeamDraft) {
		const result = await createTeamMutation.mutateAsync(payload);
		return mapProjectTeam(result.project);
	}

	function logout() {
		logoutMutation.mutate();
	}

	return (
		<SessionContext
			value={{
				session,
				loading: meQuery.isPending,
				submitting,
				isAuthenticated: Boolean(session),
				login,
				register,
				createTeam,
				logout
			}}
		>
			{children}
		</SessionContext>
	);
}
