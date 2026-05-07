import type { Meta, StoryObj } from "@storybook/react-vite";
import { FieldLabel, Select } from "../index";

const variants = ["default", "compact"] as const;

const options = [
	{ label: "ICMP ping", value: "icmp" },
	{ label: "DNS resolve", value: "dns" },
	{ label: "HTTP status", value: "http" }
];

const meta = {
	title: "Forms/Select",
	component: Select,
	args: {
		defaultValue: "icmp",
		disabled: false,
		invalid: false,
		variant: "default"
	},
	argTypes: {
		children: { control: false },
		frameClassName: { control: false },
		invalid: { control: "boolean" },
		variant: { control: "inline-radio", options: variants }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<label className="storybook-compact-field storybook-demo--narrow">
				<FieldLabel>Check type</FieldLabel>
				<Select {...args}>
					{options.map(option => (
						<option key={option.value} value={option.value}>
							{option.label}
						</option>
					))}
				</Select>
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
						<Select aria-label={`${variant} select`} variant={variant} defaultValue="dns">
							{options.map(option => (
								<option key={option.value} value={option.value}>
									{option.label}
								</option>
							))}
						</Select>
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
					<Select defaultValue="icmp">
						{options.map(option => (
							<option key={option.value} value={option.value}>
								{option.label}
							</option>
						))}
					</Select>
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Invalid</FieldLabel>
					<Select defaultValue="http" invalid>
						{options.map(option => (
							<option key={option.value} value={option.value}>
								{option.label}
							</option>
						))}
					</Select>
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Disabled</FieldLabel>
					<Select defaultValue="dns" disabled>
						{options.map(option => (
							<option key={option.value} value={option.value}>
								{option.label}
							</option>
						))}
					</Select>
				</label>
			</div>
		</div>
	)
};
