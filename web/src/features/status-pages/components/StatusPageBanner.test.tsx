// @vitest-environment jsdom

import { cleanup, fireEvent, render } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { StatusPageBanner } from "./StatusPageBanner";

afterEach(cleanup);

describe("StatusPageBanner", () => {
	it("does not reserve banner space without an image URL", () => {
		const { container, rerender } = render(<StatusPageBanner />);

		expect(container.querySelector("img")).toBeNull();

		rerender(<StatusPageBanner src="   " />);

		expect(container.querySelector("img")).toBeNull();
	});

	it("renders a trimmed image URL with the provided class", () => {
		const { container } = render(<StatusPageBanner className="status-banner" src="  https://example.com/banner.png  " />);
		const image = container.querySelector("img");

		expect(image?.getAttribute("src")).toBe("https://example.com/banner.png");
		expect(image?.className).toBe("status-banner");
		expect(image?.getAttribute("alt")).toBe("");
	});

	it("collapses a failed image and retries when its URL changes", () => {
		const { container, rerender } = render(<StatusPageBanner src="https://example.com/missing.png" />);
		const image = container.querySelector("img");

		expect(image).not.toBeNull();
		fireEvent.error(image as HTMLImageElement);
		expect(container.querySelector("img")).toBeNull();

		rerender(<StatusPageBanner src="https://example.com/replacement.png" />);

		expect(container.querySelector("img")?.getAttribute("src")).toBe("https://example.com/replacement.png");
	});
});
