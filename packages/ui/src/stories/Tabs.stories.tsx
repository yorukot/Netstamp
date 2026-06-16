import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import { Tabs, type TabItem } from "../index";

const tabs: TabItem[] = [
	{ value: "incidents", label: "Incidents", badge: "12", panelId: "tabs-incidents" },
	{ value: "rules", label: "Rules", badge: "6", panelId: "tabs-rules" },
	{ value: "notifications", label: "Notifications", panelId: "tabs-notifications" }
];

const meta = {
	title: "Components/Tabs",
	component: Tabs,
	args: {
		ariaLabel: "Alert sections",
		onValueChange: () => undefined,
		size: "md",
		tabs,
		value: "incidents"
	},
	argTypes: {
		onValueChange: { control: false },
		size: { control: "inline-radio", options: ["sm", "md"] },
		tabs: { control: false }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: function Render(args) {
		const [value, setValue] = useState(args.value);

		return (
			<div className="storybook-canvas">
				<div className="storybook-demo">
					<Tabs {...args} value={value} onValueChange={setValue} />
				</div>
			</div>
		);
	}
};

export const Sizes: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-demo">
				<Tabs ariaLabel="Small tabs" size="sm" tabs={tabs} value="rules" onValueChange={() => undefined} />
				<Tabs ariaLabel="Medium tabs" size="md" tabs={tabs} value="incidents" onValueChange={() => undefined} />
			</div>
		</div>
	)
};
