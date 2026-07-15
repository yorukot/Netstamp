import type { Meta, StoryObj } from "@storybook/react-vite";
import { Spinner } from "../index";

const meta = {
	title: "Components/Spinner",
	component: Spinner,
	args: {
		label: "Loading",
		variant: "minimal"
	},
	argTypes: {
		layout: {
			control: "select",
			options: ["inline", "compact", "panel", "page"]
		},
		size: {
			control: "select",
			options: ["sm", "md", "lg"]
		},
		variant: {
			control: "select",
			options: ["minimal", "signal"]
		}
	}
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Variants: Story = {
	render: args => (
		<div className="storybook-inline-grid">
			<Spinner {...args} variant="minimal" />
			<Spinner {...args} variant="signal" />
		</div>
	)
};

export const Sizes: Story = {
	render: args => (
		<div className="storybook-inline-grid">
			<Spinner {...args} size="sm" />
			<Spinner {...args} size="md" />
			<Spinner {...args} size="lg" />
		</div>
	)
};
