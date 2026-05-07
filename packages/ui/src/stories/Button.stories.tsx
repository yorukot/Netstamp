import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, type ButtonSize, type ButtonVariant } from "../index";

const variants: ButtonVariant[] = ["primary", "secondary", "outline", "ghost", "danger", "plain"];
const sizes: ButtonSize[] = ["sm", "md", "lg", "xl"];

const meta = {
	title: "Components/Button",
	component: Button,
	args: {
		children: "Run probe",
		disabled: false,
		size: "md",
		variant: "primary"
	},
	argTypes: {
		variant: { control: "select", options: variants },
		size: { control: "inline-radio", options: sizes },
		asChild: { control: "boolean" }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: args => (
		<div className="storybook-canvas">
			<Button {...args} />
		</div>
	)
};

export const Variants: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-row">
				{variants.map(variant => (
					<Button key={variant} variant={variant}>
						{variant}
					</Button>
				))}
			</div>
		</div>
	)
};

export const Sizes: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-row">
				{sizes.map(size => (
					<Button key={size} size={size}>
						{size}
					</Button>
				))}
			</div>
		</div>
	)
};

export const AsLink: Story = {
	render: () => (
		<div className="storybook-canvas">
			<Button asChild variant="outline">
				<a href="https://github.com/yorukot/netstamp" target="_blank" rel="noreferrer">
					Open repository
				</a>
			</Button>
		</div>
	)
};
