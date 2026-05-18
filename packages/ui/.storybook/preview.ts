import type { Preview } from "@storybook/react-vite";
import { MINIMAL_VIEWPORTS } from "storybook/viewport";
import "../src/stories/storybook.css";
import "../src/styles/tokens.css";

const preview: Preview = {
	parameters: {
		backgrounds: {
			default: "netstamp",
			options: {
				netstamp: { name: "Netstamp", value: "#010203" },
				panel: { name: "Panel", value: "#05070a" },
				light: { name: "Light", value: "#f6f0e8" }
			}
		},
		controls: {
			expanded: true,
			matchers: {
				color: /(background|color)$/i,
				date: /Date$/i
			}
		},
		layout: "fullscreen",
		options: {
			storySort: {
				order: ["Overview", "Components", "Forms"]
			}
		},
		viewport: {
			options: {
				...MINIMAL_VIEWPORTS,
				sidebarCollapsed: {
					name: "App sidebar collapsed",
					styles: {
						width: "928px",
						height: "900px"
					},
					type: "desktop"
				},
				mobileNav: {
					name: "Mobile nav",
					styles: {
						width: "608px",
						height: "900px"
					},
					type: "tablet"
				}
			}
		}
	}
};

export default preview;
