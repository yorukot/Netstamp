import type { Meta, StoryObj } from "@storybook/react-vite";
import { TextAreaField } from "../index";

const meta = {
	title: "Forms/TextAreaField",
	component: TextAreaField,
	args: {
		defaultValue: "Measure DNS and ICMP from this probe before routing public traffic.",
		helper: "Shown to operators reviewing this probe.",
		label: "Notes",
		rows: 4
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof TextAreaField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextAreaField {...args} />
			</div>
		</div>
	)
};

export const Error: Story = {
	args: {
		error: "Notes must describe the check target.",
		helper: undefined
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextAreaField {...args} />
			</div>
		</div>
	)
};

export const Disabled: Story = {
	args: {
		disabled: true,
		helper: "Disabled text areas match the shared disabled field treatment."
	},
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-demo storybook-demo--narrow">
				<TextAreaField {...args} />
			</div>
		</div>
	)
};
