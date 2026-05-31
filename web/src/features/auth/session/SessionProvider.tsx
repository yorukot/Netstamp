import { type AuthCredentials, type ProjectDraft, type RegisterPayload, createSessionSnapshot, mapApiUser, mapProject } from "@/features/auth/services/authService";
import { useCreateProjectMutation, useLoginMutation, useLogoutMutation, useRegisterMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { SessionContext } from "./SessionContext";

interface SessionProviderProps {
	children: ReactNode;
}

export function SessionProvider({ children }: SessionProviderProps) {
	const meQuery = useQuery(authQueries.me());
	const session = meQuery.data ? createSessionSnapshot(meQuery.data.user) : null;
	const loginMutation = useLoginMutation();
	const registerMutation = useRegisterMutation();
	const createProjectMutation = useCreateProjectMutation();
	const logoutMutation = useLogoutMutation();
	const submitting = loginMutation.isPending || registerMutation.isPending || createProjectMutation.isPending || logoutMutation.isPending;

	async function login(payload: AuthCredentials) {
		const result = await loginMutation.mutateAsync(payload);
		return mapApiUser(result.user);
	}

	async function register(payload: RegisterPayload) {
		const result = await registerMutation.mutateAsync(payload);
		return mapApiUser(result.user, { onboardingRequired: true });
	}

	async function createProject(payload: ProjectDraft) {
		const result = await createProjectMutation.mutateAsync(payload);
		return mapProject(result.project);
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
				createProject,
				logout
			}}
		>
			{children}
		</SessionContext>
	);
}
