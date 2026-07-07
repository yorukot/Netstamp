import { type AuthCredentials, type ProjectDraft, type RegisterPayload, createSessionSnapshot, mapApiUser, mapProject } from "@/features/auth/services/authService";
import { useCreateProjectMutation, useLoginMutation, useLogoutMutation, useRegisterMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { SessionContext } from "./SessionContext";

interface SessionProviderProps {
	children: ReactNode;
}

export function SessionProvider({ children }: SessionProviderProps) {
	const meQuery = useQuery(authQueries.me());
	const rawUser = meQuery.data?.user;
	const sessionQuery = useQuery({
		queryKey: [...apiQueryKeys.auth.me(), "session", rawUser?.id, rawUser?.email, rawUser?.displayName, rawUser?.emailVerified, rawUser?.isSystemAdmin],
		queryFn: () => {
			if (!rawUser) {
				throw new Error("Cannot create a session without a user.");
			}

			return createSessionSnapshot(rawUser);
		},
		enabled: Boolean(rawUser),
		staleTime: 5 * 60 * 1000
	});
	const session = rawUser ? (sessionQuery.data ?? null) : null;
	const loginMutation = useLoginMutation();
	const registerMutation = useRegisterMutation();
	const createProjectMutation = useCreateProjectMutation();
	const logoutMutation = useLogoutMutation();
	const submitting = loginMutation.isPending || registerMutation.isPending || createProjectMutation.isPending || logoutMutation.isPending;
	const loading = meQuery.isPending || Boolean(rawUser && sessionQuery.isPending);

	async function login(payload: AuthCredentials) {
		const result = await loginMutation.mutateAsync(payload);
		return mapApiUser(result.user);
	}

	async function register(payload: RegisterPayload) {
		const result = await registerMutation.mutateAsync(payload);
		if (!("user" in result)) {
			return { user: null, emailVerificationRequired: true as const };
		}

		return { user: await mapApiUser(result.user, { onboardingRequired: true }) };
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
				loading,
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
