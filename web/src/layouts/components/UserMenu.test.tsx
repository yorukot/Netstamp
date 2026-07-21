// @vitest-environment jsdom

import type { SessionUser } from "@/features/auth/services/authService";
import { initializeI18n } from "@/i18n";
import { ThemeProvider } from "@/shared/theme/ThemeProvider";
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeAll, describe, expect, it } from "vitest";
import { UserMenuPanel } from "./UserMenu";

const user: SessionUser = {
	id: "user-1",
	name: "Yoru",
	username: "yoru",
	email: "yoru@example.com",
	role: "user",
	emailVerified: true,
	isSystemAdmin: false,
	hasPassword: true,
	gravatarUrl: ""
};

const renderUserMenu = (isSystemAdmin: boolean) =>
	render(
		<MemoryRouter>
			<ThemeProvider>
				<UserMenuPanel user={{ ...user, isSystemAdmin }} onLogout={() => undefined} />
			</ThemeProvider>
		</MemoryRouter>
	);

beforeAll(async () => {
	await initializeI18n();
});

afterEach(() => {
	cleanup();
});

describe("UserMenuPanel", () => {
	it("links system administrators to the global administration page", () => {
		renderUserMenu(true);

		expect(screen.getByRole("link", { name: "System administration" }).getAttribute("href")).toBe("/admin");
	});

	it("hides global administration from regular users", () => {
		renderUserMenu(false);

		expect(screen.queryByRole("link", { name: "System administration" })).toBeNull();
	});
});
