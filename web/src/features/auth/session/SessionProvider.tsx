import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import { authQueries, createProject, loginUser, logoutUser, registerUser } from "@/shared/api/queries";
import { type AuthCredentials, type TeamDraft, createSessionSnapshot, mapApiUser, mapProjectTeam } from "../services/authService";
import { SessionContext } from "./SessionContext";

interface SessionProviderProps {
	children: ReactNode;
}

export function SessionProvider({ children }: SessionProviderProps) {
	const queryClient = useQueryClient();
	const meQuery = useQuery(authQueries.me());
	const session = meQuery.data ? createSessionSnapshot(meQuery.data.user) : null;
	const loginMutation = useMutation({
		mutationFn: loginUser,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const registerMutation = useMutation({
		mutationFn: registerUser,
		onSuccess: data => {
			queryClient.setQueryData(apiQueryKeys.auth.me(), { authenticated: true, user: data.user });
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const createTeamMutation = useMutation({
		mutationFn: createProject,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const logoutMutation = useMutation({
		mutationFn: logoutUser,
		onSettled: () => {
			queryClient.removeQueries({ queryKey: apiQueryKeys.auth.all });
			queryClient.removeQueries({ queryKey: apiQueryKeys.projects.all });
		}
	});
	const submitting = loginMutation.isPending || registerMutation.isPending || createTeamMutation.isPending || logoutMutation.isPending;

	async function login(payload: AuthCredentials) {
		const result = await loginMutation.mutateAsync(payload);
		return mapApiUser(result.user);
	}

	async function register(payload: AuthCredentials) {
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
