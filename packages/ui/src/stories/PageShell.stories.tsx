import type { Meta, StoryObj } from "@storybook/react-vite";
import { PageShell, type PageShellProps, type PageShellVariant } from "../index";

const variants: PageShellVariant[] = ["grid", "constellation"];

const meta = {
	title: "Components/PageShell",
	args: {
		as: "div",
		center: false,
		className: "storybook-shell-demo",
		variant: "grid"
	},
	argTypes: {
		as: { control: false },
		center: { control: "boolean" },
		children: { control: false },
		className: { control: false },
		variant: { control: "inline-radio", options: variants }
	},
	render: args => (
		<PageShell {...args}>
			<div className="storybook-shell-card">
				<span className="storybook-code">PageShell / {args.variant}</span>
				<strong>Full-page operational background</strong>
				<p>Use PageShell when a route needs the shared grid or constellation backdrop.</p>
			</div>
		</PageShell>
	),
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<PageShellProps<"div">>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Grid: Story = {};

export const Constellation: Story = {
	args: {
		variant: "constellation"
	}
};

export const Centered: Story = {
	args: {
		center: true,
		variant: "constellation"
	}
};
