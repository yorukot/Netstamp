import type { CreateProjectInput, LoginInput, RegisterInput, UserResponse } from "@/shared/api/types";

export type AuthCredentials = LoginInput;
export type RegisterPayload = RegisterInput;

export type ProjectDraft = CreateProjectInput;

export interface SessionUser {
	id: string;
	name: string;
	username: string;
	email: string;
	role: string;
	gravatarUrl: string;
	onboardingRequired?: boolean;
}

export interface SessionSnapshot {
	user: SessionUser;
	controller: "connected";
	project?: ProjectDraft & { role: string };
}

export function mapApiUser(user: UserResponse, options: { onboardingRequired?: boolean } = {}): SessionUser {
	const email = user.email || "";
	const displayName = user.displayName || email.split("@")[0] || "Netstamp user";

	return {
		id: user.id,
		name: displayName,
		username: email.split("@")[0] || displayName,
		email,
		role: "Admin",
		gravatarUrl: `https://www.gravatar.com/avatar/?d=identicon&size=160`,
		onboardingRequired: options.onboardingRequired
	};
}

export function createSessionSnapshot(user: UserResponse, options: { onboardingRequired?: boolean } = {}): SessionSnapshot {
	return {
		user: mapApiUser(user, options),
		controller: "connected"
	};
}

export function mapProject({ name, slug }: ProjectDraft): ProjectDraft & { role: string } {
	return { name: name || "Vector IX", slug: slug || "vector-ix", role: "Owner" };
}
