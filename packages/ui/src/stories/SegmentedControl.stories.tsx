import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import { SegmentedControl, type SegmentedControlOption } from "../index";

const options: SegmentedControlOption[] = [
	{ value: "map", label: "Map" },
	{ value: "grid", label: "Grid" },
	{ value: "timeline", label: "Timeline" }
];

const meta = {
	title: "Components/SegmentedControl",
	component: SegmentedControl,
	args: {
		ariaLabel: "View mode",
		onValueChange: () => undefined,
		options,
		size: "md",
		value: "map"
	},
	argTypes: {
		onValueChange: { control: false },
		options: { control: false },
		size: { control: "inline-radio", options: ["sm", "md"] }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof SegmentedControl>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: function Render(args) {
		const [value, setValue] = useState(args.value);

		return (
			<div className="storybook-canvas">
				<SegmentedControl {...args} value={value} onValueChange={setValue} />
			</div>
		);
	}
};

export const Sizes: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-row">
				<SegmentedControl ariaLabel="Small view mode" size="sm" options={options} value="grid" onValueChange={() => undefined} />
				<SegmentedControl ariaLabel="Medium view mode" size="md" options={options} value="map" onValueChange={() => undefined} />
			</div>
		</div>
	)
};
