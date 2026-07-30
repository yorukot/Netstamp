import { describe, expect, it } from "vitest";
import { applySettingsIntent, googleAllowedDomainsFromText, secretValueFromForm, updateSettingsIntent } from "./adminSettingsForm";

describe("admin settings form values", () => {
	it("normalizes configured Google domains", () => {
		expect(
			googleAllowedDomainsFromText(` Example.COM, staff.example.com
EXAMPLE.com `)
		).toEqual(["example.com", "staff.example.com"]);
	});

	it("keeps a separator-only Google domain value invalid", () => {
		expect(googleAllowedDomainsFromText(",")).toEqual([""]);
	});

	it("allows a truly blank Google domain list", () => {
		expect(googleAllowedDomainsFromText("  ")).toEqual([]);
	});

	it("maps secret forms to retain, replace, and clear values", () => {
		expect(secretValueFromForm("", false)).toBeUndefined();
		expect(secretValueFromForm("replacement", false)).toBe("replacement");
		expect(secretValueFromForm("replacement", true)).toBeNull();
	});

	it("rebases only touched fields onto a newer server version", () => {
		const original = { enabled: false, displayName: "Original" };
		const intent = updateSettingsIntent(null, original, "enabled", true);
		const changedElsewhere = { enabled: false, displayName: "Changed elsewhere" };

		expect(intent).toEqual({ enabled: true });
		expect(applySettingsIntent(changedElsewhere, intent)).toEqual({
			enabled: true,
			displayName: "Changed elsewhere"
		});
	});

	it("removes intent when a field is restored to the server value", () => {
		const server = { enabled: false };
		const changed = updateSettingsIntent(null, server, "enabled", true);
		expect(updateSettingsIntent(changed, server, "enabled", false)).toBeNull();
	});
});
