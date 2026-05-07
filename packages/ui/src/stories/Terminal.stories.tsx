import type { Meta, StoryObj } from "@storybook/react-vite";
import { Terminal } from "../index";

const meta = {
	title: "Components/Terminal",
	component: Terminal,
	args: {
		children: "netstamp probe run --check ping --target 1.1.1.1",
		meta: "dry run",
		title: "netstamp probe"
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Terminal>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Terminal {...args} />
			</div>
		</div>
	)
};

export const LongOutput: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Terminal title="trace" meta="live">
					{`$ netstamp trace cloudflare.com --region tpe\nresolve cloudflare.com -> 104.16.132.229\nicmp p50=31ms p95=42ms loss=0.00%\nhttp 200 edge=TPE cache=HIT`}
				</Terminal>
			</div>
		</div>
	)
};
