import type { Meta, StoryObj } from "@storybook/react-vite";
import { TextField } from "../index";

const meta = {
	title: "Forms/TextField",
	component: TextField,
	args: {
		defaultValue: "tpe-lab-02",
		helper: "Visible in route and probe tables.",
		label: "Probe name",
		placeholder: "probe-name"
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof TextField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextField {...args} />
			</div>
		</div>
	)
};

export const Error: Story = {
	args: {
		defaultValue: "",
		error: "Probe name is required.",
		helper: undefined
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextField {...args} />
			</div>
		</div>
	)
};

export const Disabled: Story = {
	args: {
		defaultValue: "demo-probe",
		disabled: true,
		helper: "Disabled fields use the same subdued treatment as selects."
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextField {...args} />
			</div>
		</div>
	)
};
