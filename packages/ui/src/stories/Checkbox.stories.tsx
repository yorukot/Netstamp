import type { Meta, StoryObj } from "@storybook/react-vite";
import { Checkbox, FieldLabel } from "../index";

const meta = {
	title: "Forms/Checkbox",
	component: Checkbox,
	args: {
		defaultChecked: true,
		disabled: false,
		invalid: false
	},
	argTypes: {
		invalid: { control: "boolean" }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<label className="storybook-checkbox-row">
				<Checkbox {...args} />
				<FieldLabel>Enable alerts</FieldLabel>
			</label>
		</div>
	)
};

export const States: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-form-stack">
				<label className="storybook-checkbox-row">
					<Checkbox defaultChecked />
					<FieldLabel>Checked</FieldLabel>
				</label>
				<label className="storybook-checkbox-row">
					<Checkbox />
					<FieldLabel>Unchecked</FieldLabel>
				</label>
				<label className="storybook-checkbox-row">
					<Checkbox invalid />
					<FieldLabel>Invalid</FieldLabel>
				</label>
				<label className="storybook-checkbox-row">
					<Checkbox disabled />
					<FieldLabel>Disabled</FieldLabel>
				</label>
			</div>
		</div>
	)
};
