import type { Meta, StoryObj } from "@storybook/react-vite";
import { SelectField } from "../index";

const options = [
	{ label: "10 seconds", value: "10s" },
	{ label: "30 seconds", value: "30s" },
	{ label: "60 seconds", value: "60s" }
];

const meta = {
	title: "Forms/SelectField",
	component: SelectField,
	args: {
		defaultValue: "30s",
		helper: "How often this probe should run.",
		label: "Interval",
		options
	},
	argTypes: {
		options: { control: false }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof SelectField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<SelectField {...args} />
			</div>
		</div>
	)
};

export const Error: Story = {
	args: {
		error: "Choose an interval before saving.",
		helper: undefined
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<SelectField {...args} />
			</div>
		</div>
	)
};
