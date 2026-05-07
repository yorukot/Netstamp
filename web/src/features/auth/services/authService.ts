import { currentUser, type CurrentUser } from "../data/mockUser";

export interface AuthCredentials {
	displayName?: string;
	email: string;
	password: string;
}

export interface TeamDraft {
	name: string;
	slug: string;
}

export interface MockUser extends CurrentUser {
	role: string;
	team: string;
	onboardingRequired?: boolean;
}

export interface SessionSnapshot {
	user: MockUser;
	controller: "waiting-for-api";
	team?: TeamDraft & { role: string };
}

const mockUser: MockUser = {
	...currentUser,
	team: "Vector IX"
};

export async function mockLogin({ email }: AuthCredentials): Promise<MockUser> {
	return { ...mockUser, email: email || mockUser.email };
}

export async function mockRegister({ displayName, email }: AuthCredentials): Promise<MockUser> {
	return {
		...mockUser,
		name: displayName || mockUser.name,
		email: email || mockUser.email,
		onboardingRequired: true
	};
}

export async function mockCreateTeam({ name, slug }: TeamDraft): Promise<TeamDraft & { role: string }> {
	return { name: name || "Vector IX", slug: slug || "vector-ix", role: "Owner" };
}

export function getSessionSnapshot(): SessionSnapshot {
	return { user: mockUser, controller: "waiting-for-api" };
}
