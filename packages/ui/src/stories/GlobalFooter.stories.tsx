import type { Meta, StoryObj } from "@storybook/react-vite";
import { GlobalFooter } from "../index";

const meta = {
	title: "Components/GlobalFooter",
	component: GlobalFooter,
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof GlobalFooter>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-footer-frame">
				<GlobalFooter {...args} />
			</div>
		</div>
	)
};
