import type { Meta, StoryObj } from "@storybook/react-vite";
import { FieldLabel, Input, type ControlVariant } from "../index";

const variants: ControlVariant[] = ["default", "compact", "bare"];

const meta = {
	title: "Forms/Input",
	component: Input,
	args: {
		defaultValue: "1.1.1.1",
		disabled: false,
		invalid: false,
		placeholder: "probe target",
		variant: "default"
	},
	argTypes: {
		frameClassName: { control: false },
		invalid: { control: "boolean" },
		variant: { control: "inline-radio", options: variants }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<label className="storybook-compact-field storybook-demo--narrow">
				<FieldLabel>Probe target</FieldLabel>
				<Input {...args} />
			</label>
		</div>
	)
};

export const Variants: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				{variants.map(variant => (
					<label key={variant} className="storybook-compact-field">
						<FieldLabel>{variant}</FieldLabel>
						<Input aria-label={`${variant} input`} variant={variant} defaultValue="trace-window" />
					</label>
				))}
			</div>
		</div>
	)
};

export const States: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				<label className="storybook-compact-field">
					<FieldLabel>Default</FieldLabel>
					<Input defaultValue="cloudflare.com" />
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Invalid</FieldLabel>
					<Input defaultValue="not a host" invalid />
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Disabled</FieldLabel>
					<Input defaultValue="disabled" disabled />
				</label>
			</div>
		</div>
	)
};
