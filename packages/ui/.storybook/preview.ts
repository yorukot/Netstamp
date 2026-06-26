import type { Decorator, Preview } from "@storybook/react-vite";
import { MINIMAL_VIEWPORTS } from "storybook/viewport";
import "../src/stories/storybook.css";
import "../src/styles/tokens.css";

const withTheme: Decorator = (Story, context) => {
	const theme = context.globals.colorMode === "dark" ? "dark" : "light";

	if (typeof document !== "undefined") {
		document.documentElement.dataset.theme = theme;
	}

	return Story();
};

const preview: Preview = {
	globalTypes: {
		colorMode: {
			defaultValue: "light",
			description: "Switch Netstamp design tokens between light and dark mode.",
			name: "Theme",
			toolbar: {
				dynamicTitle: true,
				icon: "contrast",
				items: [
					{ title: "Light", value: "light" },
					{ title: "Dark", value: "dark" }
				],
				title: "Theme"
			}
		}
	},
	decorators: [withTheme],
	parameters: {
		backgrounds: {
			default: "light",
			options: {
				light: { name: "Light canvas", value: "#f6f8fb" },
				lightPanel: { name: "Light panel", value: "#ffffff" },
				dark: { name: "Dark canvas", value: "#0d1520" },
				darkPanel: { name: "Dark panel", value: "#151f2d" }
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
				order: ["Overview", "Foundations", "Components", "Forms", "Patterns"]
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
