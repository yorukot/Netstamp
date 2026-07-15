import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, DangerAction } from "../index";

const meta = {
	title: "Components/DangerAction",
	component: DangerAction,
	args: {
		title: "Deactivate account",
		description: "Disable sign-in and protected route access until an administrator re-enables the account.",
		descriptionId: "danger-action-description",
		action: (
			<Button variant="danger" aria-describedby="danger-action-description">
				Deactivate account
			</Button>
		)
	},
	parameters: {
		layout: "padded"
	}
} satisfies Meta<typeof DangerAction>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Disabled: Story = {
	args: {
		descriptionId: "disabled-danger-action-description",
		action: (
			<Button variant="danger" aria-describedby="disabled-danger-action-description" disabled>
				Deactivate account
			</Button>
		)
	}
};
