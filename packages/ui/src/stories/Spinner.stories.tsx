import type { Meta, StoryObj } from "@storybook/react-vite";
import { Spinner } from "../index";

const meta = {
	title: "Components/Spinner",
	component: Spinner,
	args: {
		label: "Loading"
	}
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Sizes: Story = {
	render: args => (
		<div className="storybook-inline-grid">
			<Spinner {...args} size="sm" />
			<Spinner {...args} size="md" />
			<Spinner {...args} size="lg" />
		</div>
	)
};
