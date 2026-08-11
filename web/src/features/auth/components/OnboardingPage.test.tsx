// @vitest-environment jsdom

import { initializeI18n } from "@/i18n";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import { OnboardingPage } from "./OnboardingPage";

const mocks = vi.hoisted(() => ({
	acceptInvite: vi.fn(),
	apiProblemCode: vi.fn((error: unknown) => (error as { code?: string })?.code),
	createInvite: vi.fn(),
	createProject: vi.fn(),
	logout: vi.fn(),
	navigate: vi.fn(),
	pushErrorToast: vi.fn(),
	pushToast: vi.fn(),
	setSelectedProjectRef: vi.fn(),
	useAcceptProjectInviteMutation: vi.fn(),
	useAuth: vi.fn(),
	useCreateProjectInviteForRefMutation: vi.fn(),
	useCreateProjectMutation: vi.fn(),
	useProjectSelection: vi.fn(),
	useQuery: vi.fn(),
	useRuntimeFeatures: vi.fn()
}));

vi.mock("@/features/auth/hooks/useAuth", () => ({ useAuth: mocks.useAuth }));
vi.mock("@/shared/api/client", () => ({ apiProblemCode: mocks.apiProblemCode }));
vi.mock("@/shared/api/mutations", () => ({
	useAcceptProjectInviteMutation: mocks.useAcceptProjectInviteMutation,
	useCreateProjectInviteForRefMutation: mocks.useCreateProjectInviteForRefMutation,
	useCreateProjectMutation: mocks.useCreateProjectMutation
}));
vi.mock("@/shared/api/useCurrentProject", () => ({ useProjectSelection: mocks.useProjectSelection }));
vi.mock("@/shared/config/features", () => ({ useRuntimeFeatures: mocks.useRuntimeFeatures }));
vi.mock("@/shared/toast/toastStore", () => ({ pushErrorToast: mocks.pushErrorToast, pushToast: mocks.pushToast }));
vi.mock("@tanstack/react-query", async importOriginal => ({
	...(await importOriginal<typeof import("@tanstack/react-query")>()),
	useQuery: mocks.useQuery
}));
vi.mock("./AuthLayout", () => ({
	AuthLayout: ({ title, description, children }: { title: string; description?: string; children: ReactNode }) => (
		<main>
			<h1>{title}</h1>
			{description ? <p>{description}</p> : null}
			{children}
		</main>
	)
}));

const user = {
	id: "user-1",
	name: "River",
	username: "river",
	email: "river@example.com",
	role: "user",
	emailVerified: true,
	isSystemAdmin: false,
	hasPassword: true,
	gravatarUrl: ""
};

const pendingInvite = {
	id: "invite-1",
	role: "viewer",
	project: {
		id: "project-invited-id",
		name: "Edge Operations",
		slug: "edge-operations"
	}
};

const renderOnboarding = () => render(<OnboardingPage navigate={mocks.navigate} />);

beforeAll(async () => {
	await initializeI18n();
});

beforeEach(() => {
	mocks.acceptInvite.mockReset();
	mocks.createInvite.mockReset();
	mocks.createProject.mockReset();
	mocks.logout.mockReset();
	mocks.navigate.mockReset();
	mocks.pushErrorToast.mockReset();
	mocks.pushToast.mockReset();
	mocks.setSelectedProjectRef.mockReset();
	mocks.apiProblemCode.mockClear();
	mocks.useAuth.mockReturnValue({
		session: { user, controller: "connected" },
		loading: false,
		submitting: false,
		logout: mocks.logout
	});
	mocks.useRuntimeFeatures.mockReturnValue({
		readOnlyMode: false,
		appFeatures: { registration: true, projectCreation: true, userCredentialChanges: true }
	});
	mocks.useQuery.mockReturnValue({ data: { invites: [] }, isPending: false });
	mocks.useProjectSelection.mockReturnValue({ setSelectedProjectRef: mocks.setSelectedProjectRef });
	mocks.useCreateProjectMutation.mockReturnValue({ mutateAsync: mocks.createProject, isPending: false });
	mocks.useCreateProjectInviteForRefMutation.mockReturnValue({ mutateAsync: mocks.createInvite, isPending: false });
	mocks.useAcceptProjectInviteMutation.mockReturnValue({ mutateAsync: mocks.acceptInvite, isPending: false });
	mocks.createProject.mockImplementation(async ({ name, slug }: { name: string; slug: string }) => ({
		project: { id: "project-created-id", name, slug }
	}));
	mocks.createInvite.mockResolvedValue({ invite: {} });
	mocks.acceptInvite.mockResolvedValue({ invite: pendingInvite });
});

afterEach(cleanup);

describe("OnboardingPage", () => {
	it("shows an immediately usable required project name without seeded product data", () => {
		renderOnboarding();

		const nameInput = screen.getByRole("textbox", { name: "Project name" });
		const createButton = screen.getByRole("button", { name: "Create project" });

		expect(nameInput.hasAttribute("required")).toBe(true);
		expect(nameInput.getAttribute("placeholder")).toBeNull();
		expect(createButton.hasAttribute("disabled")).toBe(true);
		expect(screen.queryByText(/Yoru Labs|first-contact|yoru:\/\//i)).toBeNull();
	});

	it("creates an ASCII-named project and opens its dashboard", async () => {
		renderOnboarding();

		fireEvent.change(screen.getByRole("textbox", { name: "Project name" }), { target: { value: "Taiwan Edge" } });
		fireEvent.click(screen.getByRole("button", { name: "Create project" }));

		await waitFor(() => expect(mocks.createProject).toHaveBeenCalledWith({ name: "Taiwan Edge", slug: "taiwan-edge" }));
		expect(mocks.setSelectedProjectRef).toHaveBeenCalledWith("taiwan-edge");
		expect(mocks.navigate).toHaveBeenCalledWith("dashboard", { projectRef: "taiwan-edge" });
		expect(mocks.pushToast).toHaveBeenCalledWith(expect.objectContaining({ tone: "success" }));
	});

	it("uses a generic random slug for a non-ASCII project name", async () => {
		renderOnboarding();

		fireEvent.change(screen.getByRole("textbox", { name: "Project name" }), { target: { value: "台灣監控" } });
		fireEvent.click(screen.getByRole("button", { name: "Create project" }));

		await waitFor(() => expect(mocks.createProject).toHaveBeenCalled());
		expect(mocks.createProject.mock.calls[0]?.[0]).toMatchObject({ name: "台灣監控", slug: expect.stringMatching(/^project-[a-z0-9]{6}$/) });
	});

	it("retries a conflicting slug with a random suffix", async () => {
		mocks.createProject.mockRejectedValueOnce({ code: "PROJECT_SLUG_ALREADY_EXISTS" }).mockResolvedValueOnce({
			project: { id: "project-created-id", name: "Edge", slug: "edge-abc123" }
		});
		renderOnboarding();

		fireEvent.change(screen.getByRole("textbox", { name: "Project name" }), { target: { value: "Edge" } });
		fireEvent.click(screen.getByRole("button", { name: "Create project" }));

		await waitFor(() => expect(mocks.createProject).toHaveBeenCalledTimes(2));
		expect(mocks.createProject.mock.calls[0]?.[0]).toEqual({ name: "Edge", slug: "edge" });
		expect(mocks.createProject.mock.calls[1]?.[0]).toMatchObject({ name: "Edge", slug: expect.stringMatching(/^edge-[a-z0-9]{6}$/) });
	});

	it("keeps member invitations optional and reports partial invite failures without blocking navigation", async () => {
		mocks.createInvite.mockResolvedValueOnce({ invite: {} }).mockRejectedValueOnce(new Error("Invite failed"));
		renderOnboarding();

		fireEvent.click(screen.getByRole("button", { name: "Show member invitations" }));
		expect(screen.queryByRole("button", { name: "Remove member 1" })).toBeNull();
		fireEvent.change(screen.getByRole("textbox", { name: "Member 1 email" }), { target: { value: "one@example.com" } });
		fireEvent.click(screen.getByRole("button", { name: "Add another member" }));
		expect(screen.getAllByRole("button", { name: /Remove member/ })).toHaveLength(2);
		fireEvent.change(screen.getByRole("textbox", { name: "Member 2 email" }), { target: { value: "two@example.com" } });
		fireEvent.change(screen.getByRole("textbox", { name: "Project name" }), { target: { value: "Operations" } });
		fireEvent.click(screen.getByRole("button", { name: "Create project" }));

		await waitFor(() => expect(mocks.createInvite).toHaveBeenCalledTimes(2));
		expect(mocks.pushErrorToast).toHaveBeenCalledWith("1 project invite could not be sent.");
		expect(mocks.navigate).toHaveBeenCalledWith("dashboard", { projectRef: "operations" });
	});

	it("opens an accepted pending project from a clear invitation choice", async () => {
		mocks.useQuery.mockReturnValue({ data: { invites: [pendingInvite] }, isPending: false });
		renderOnboarding();

		expect(screen.getByRole("heading", { name: "You have a project invitation" })).toBeTruthy();
		expect(screen.queryByRole("textbox", { name: "Project name" })).toBeNull();
		fireEvent.click(screen.getByRole("button", { name: "Open invited project" }));

		await waitFor(() => expect(mocks.acceptInvite).toHaveBeenCalledWith("invite-1"));
		expect(mocks.setSelectedProjectRef).toHaveBeenCalledWith("edge-operations");
		expect(mocks.navigate).toHaveBeenCalledWith("dashboard", { projectRef: "edge-operations" });
	});

	it("accepts pending invitations before revealing the new project form", async () => {
		mocks.useQuery.mockReturnValue({ data: { invites: [pendingInvite] }, isPending: false });
		renderOnboarding();

		fireEvent.click(screen.getByRole("button", { name: "Create a new project too" }));

		await waitFor(() => expect(screen.getByRole("textbox", { name: "Project name" })).toBeTruthy());
		expect(mocks.acceptInvite).toHaveBeenCalledWith("invite-1");
		expect(mocks.navigate).not.toHaveBeenCalled();
	});

	it("shows a formal no-access state when project creation is unavailable", () => {
		mocks.useRuntimeFeatures.mockReturnValue({
			readOnlyMode: false,
			appFeatures: { registration: false, projectCreation: false, userCredentialChanges: false }
		});
		renderOnboarding();

		expect(screen.getByRole("heading", { name: "No project access" })).toBeTruthy();
		expect(screen.getByText(/Ask an administrator to invite this account/)).toBeTruthy();
		fireEvent.click(screen.getByRole("button", { name: "Log out" }));
		expect(mocks.logout).toHaveBeenCalledOnce();
	});
});
