import type { CreateProjectInput, LoginInput, RegisterInput, UserResponse } from "@/shared/api/types";
import { createGravatarUrl } from "@/shared/utils/gravatar";

export type AuthCredentials = LoginInput;
export type RegisterPayload = RegisterInput;

export type ProjectDraft = CreateProjectInput;

export interface SessionUser {
	id: string;
	name: string;
	username: string;
	email: string;
	role: string;
	isSystemAdmin: boolean;
	gravatarUrl: string;
	onboardingRequired?: boolean;
}

export interface SessionSnapshot {
	user: SessionUser;
	controller: "connected";
	project?: ProjectDraft & { role: string };
}

export async function mapApiUser(user: UserResponse, options: { onboardingRequired?: boolean } = {}): Promise<SessionUser> {
	const email = user.email || "";
	const displayName = user.displayName || email.split("@")[0] || "Netstamp user";

	return {
		id: user.id,
		name: displayName,
		username: email.split("@")[0] || displayName,
		email,
		role: user.isSystemAdmin ? "Global admin" : "User",
		isSystemAdmin: Boolean(user.isSystemAdmin),
		gravatarUrl: await createGravatarUrl(email),
		onboardingRequired: options.onboardingRequired
	};
}

export async function createSessionSnapshot(user: UserResponse, options: { onboardingRequired?: boolean } = {}): Promise<SessionSnapshot> {
	return {
		user: await mapApiUser(user, options),
		controller: "connected"
	};
}

export function mapProject({ name, slug }: ProjectDraft): ProjectDraft & { role: string } {
	return { name: name || "Vector IX", slug: slug || "vector-ix", role: "Owner" };
}
