import type { Meta, StoryObj } from "@storybook/react-vite";
import { FieldLabel, Input } from "../index";

const meta = {
	title: "Forms/FieldLabel",
	component: FieldLabel,
	args: {
		children: "Probe target"
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof FieldLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	render: args => (
		<div className="storybook-canvas">
			<FieldLabel {...args} />
		</div>
	)
};

export const WithControl: Story = {
	render: () => (
		<div className="storybook-canvas">
			<label className="storybook-compact-field storybook-demo--narrow">
				<FieldLabel>Compact input</FieldLabel>
				<Input aria-label="Compact input preview" defaultValue="trace-window" variant="compact" />
			</label>
		</div>
	)
};
