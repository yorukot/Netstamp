import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, Panel, type PanelTone } from "../index";

const tones: PanelTone[] = ["glass", "matte", "deep"];

const meta = {
	title: "Components/Panel",
	component: Panel,
	args: {
		children: <p>Probe telemetry, recent incidents, and action controls can live inside a Panel body.</p>,
		eyebrow: "Operations",
		padded: true,
		title: "Probe summary",
		tone: "glass"
	},
	argTypes: {
		actions: { control: false },
		as: { control: false },
		children: { control: false },
		padded: { control: "boolean" },
		tone: { control: "inline-radio", options: tones }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Panel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Panel {...args} />
			</div>
		</div>
	)
};

export const WithActions: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Panel eyebrow="Runbook" title="Incident queue" actions={<Button size="sm">Acknowledge</Button>}>
					<p>Keep primary actions visible while preserving the cut-frame panel treatment.</p>
				</Panel>
			</div>
		</div>
	)
};

export const Tones: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				{tones.map(tone => (
					<Panel key={tone} eyebrow="Tone" title={tone} tone={tone}>
						<p>Panel tone: {tone}</p>
					</Panel>
				))}
			</div>
		</div>
	)
};

export const Unpadded: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Panel eyebrow="Dense" title="Unpadded table shell" padded={false}>
					<p className="storybook-specimen">Content can opt into its own spacing when Panel padding is disabled.</p>
				</Panel>
			</div>
		</div>
	)
};
