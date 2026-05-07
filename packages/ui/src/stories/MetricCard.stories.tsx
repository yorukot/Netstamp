import type { Meta, StoryObj } from "@storybook/react-vite";
import { MetricCard, type BadgeTone } from "../index";

const tones: BadgeTone[] = ["neutral", "accent", "success", "warning", "critical", "muted"];

const meta = {
	title: "Components/MetricCard",
	component: MetricCard,
	args: {
		detail: "healthy",
		label: "p95 latency",
		tone: "success",
		value: "42ms"
	},
	argTypes: {
		tone: { control: "select", options: tones }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof MetricCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<MetricCard {...args} />
			</div>
		</div>
	)
};

export const Tones: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				{tones.map(tone => (
					<MetricCard key={tone} label={`${tone} metric`} value="99.9%" detail={tone} tone={tone} />
				))}
			</div>
		</div>
	)
};

export const OperationsGrid: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				<MetricCard label="regions" value="18" detail="active" tone="accent" />
				<MetricCard label="p95 latency" value="42ms" detail="healthy" tone="success" />
				<MetricCard label="packet loss" value="0.08%" detail="watch" tone="warning" />
			</div>
		</div>
	)
};
