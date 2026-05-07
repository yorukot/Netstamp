import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge, type BadgeTone } from "../index";

const tones: BadgeTone[] = ["neutral", "accent", "success", "warning", "critical", "muted"];

const meta = {
	title: "Components/Badge",
	component: Badge,
	args: {
		children: "Probe online",
		dot: true,
		tone: "accent"
	},
	argTypes: {
		tone: { control: "select", options: tones },
		dot: { control: "boolean" }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<Badge {...args} />
		</div>
	)
};

export const Tones: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<section className="storybook-specimen">
					<span>Tones</span>
					<div className="storybook-row">
						{tones.map(tone => (
							<Badge key={tone} tone={tone}>
								{tone}
							</Badge>
						))}
					</div>
				</section>
			</div>
		</div>
	)
};

export const WithoutDots: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-row">
				{tones.map(tone => (
					<Badge key={tone} tone={tone} dot={false}>
						{tone}
					</Badge>
				))}
			</div>
		</div>
	)
};
