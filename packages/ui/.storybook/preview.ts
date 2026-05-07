import type { Preview } from "@storybook/react-vite";
import "../src/stories/storybook.css";
import "../src/styles/tokens.css";

const preview: Preview = {
	parameters: {
		controls: {
			expanded: true
		},
		layout: "fullscreen",
		options: {
			storySort: {
				order: ["Overview", "Components", "Forms"]
			}
		}
	}
};

export default preview;
