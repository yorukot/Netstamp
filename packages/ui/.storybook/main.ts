import type { StorybookConfig } from "@storybook/react-vite";

const config: StorybookConfig = {
	stories: ["../src/**/*.stories.@(ts|tsx)"],
	addons: [],
	features: {
		actions: true,
		backgrounds: true,
		controls: true,
		highlight: true,
		measure: true,
		outline: true,
		viewport: true
	},
	framework: {
		name: "@storybook/react-vite",
		options: {}
	},
	core: {
		disableTelemetry: true
	}
};

export default config;
