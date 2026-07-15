import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import { Button, Drawer, SelectField, TextAreaField, TextField } from "../index";

const meta = {
	title: "Components/Drawer",
	component: Drawer,
	args: {
		children: null,
		description: "Edit a probe without leaving the current operational view.",
		onOpenChange: () => undefined,
		open: false,
		side: "right",
		size: "md",
		title: "Probe editor"
	},
	argTypes: {
		actions: { control: false },
		children: { control: false },
		onOpenChange: { control: false },
		side: { control: "inline-radio", options: ["left", "right"] },
		size: { control: "select", options: ["sm", "md", "lg", "full"] }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Drawer>;

export default meta;
type Story = StoryObj<typeof meta>;

const drawerFields = (
	<div className="storybook-demo">
		<TextField label="Probe name" defaultValue="probe-tpe-01" />
		<SelectField
			label="Region"
			defaultValue="apac"
			options={[
				{ value: "apac", label: "APAC" },
				{ value: "emea", label: "EMEA" },
				{ value: "amer", label: "AMER" }
			]}
		/>
		<TextAreaField label="Notes" defaultValue="Taipei edge measurement node." />
	</div>
);

export const Playground: Story = {
	render: function Render(args) {
		const [open, setOpen] = useState(false);

		return (
			<div className="storybook-canvas">
				<Button type="button" onClick={() => setOpen(true)}>
					Open drawer
				</Button>
				{open ? (
					<Drawer {...args} open onOpenChange={setOpen}>
						{drawerFields}
					</Drawer>
				) : null}
			</div>
		);
	}
};

export const PersistentMount: Story = {
	render: function Render(args) {
		const [open, setOpen] = useState(false);

		return (
			<div className="storybook-canvas">
				<Button type="button" onClick={() => setOpen(true)}>
					Open drawer
				</Button>
				<Drawer {...args} open={open} onOpenChange={setOpen}>
					{drawerFields}
				</Drawer>
			</div>
		);
	}
};
