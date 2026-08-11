import { i18n } from "@/i18n";
import type { LoginInput, RegisterInput, UserResponse } from "@/shared/api/types";
import { createGravatarUrl } from "@/shared/utils/gravatar";

const authT = i18n.getFixedT(null, "auth") as (key: string) => string;

export type AuthCredentials = LoginInput;
export type RegisterPayload = RegisterInput;

export interface SessionUser {
	id: string;
	name: string;
	username: string;
	email: string;
	role: string;
	emailVerified: boolean;
	isSystemAdmin: boolean;
	hasPassword: boolean;
	gravatarUrl: string;
	onboardingRequired?: boolean;
}

export type RegisterResult = { user: SessionUser; emailVerificationRequired?: false } | { user: null; emailVerificationRequired: true };

export interface SessionSnapshot {
	user: SessionUser;
	controller: "connected";
}

export async function mapApiUser(user: UserResponse, options: { onboardingRequired?: boolean } = {}): Promise<SessionUser> {
	const email = user.email || "";
	const displayName = user.displayName || email.split("@")[0] || authT("fallbackUser");

	return {
		id: user.id,
		name: displayName,
		username: email.split("@")[0] || displayName,
		email,
		role: user.isSystemAdmin ? "global-admin" : "user",
		emailVerified: Boolean(user.emailVerified),
		isSystemAdmin: Boolean(user.isSystemAdmin),
		hasPassword: Boolean(user.hasPassword),
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
