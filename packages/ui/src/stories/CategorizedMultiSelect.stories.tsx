import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import { CategorizedMultiSelect, type CategorizedMultiSelectCategory } from "../index";

const categories: CategorizedMultiSelectCategory[] = [
	{
		value: "check",
		label: "By Check",
		items: [
			{
				value: "assignment-cloudflare-tpe",
				label: "Cloudflare DNS",
				description: "Ping / 1.1.1.1 / probe-tpe-01",
				searchText: "cloudflare dns ping 1.1.1.1 probe-tpe-01 taipei",
				selectionValues: ["assignment-cloudflare-tpe"]
			},
			{
				value: "assignment-cloudflare-fra",
				label: "Cloudflare DNS",
				description: "Ping / 1.1.1.1 / probe-fra-03",
				searchText: "cloudflare dns ping 1.1.1.1 probe-fra-03 frankfurt",
				selectionValues: ["assignment-cloudflare-fra"]
			},
			{
				value: "assignment-api-tpe",
				label: "Public API",
				description: "HTTP / https://api.example.com / probe-tpe-01",
				searchText: "public api http https api example probe-tpe-01 taipei",
				selectionValues: ["assignment-api-tpe"]
			}
		]
	},
	{
		value: "probe",
		label: "By Probe",
		items: [
			{
				value: "assignment-api-tpe",
				label: "probe-tpe-01",
				description: "Taipei / Public API / HTTP",
				searchText: "probe-tpe-01 taipei public api http",
				selectionValues: ["assignment-api-tpe"]
			},
			{
				value: "assignment-cloudflare-tpe",
				label: "probe-tpe-01",
				description: "Taipei / Cloudflare DNS / Ping",
				searchText: "probe-tpe-01 taipei cloudflare dns ping",
				selectionValues: ["assignment-cloudflare-tpe"]
			},
			{
				value: "assignment-cloudflare-fra",
				label: "probe-fra-03",
				description: "Frankfurt / Cloudflare DNS / Ping",
				searchText: "probe-fra-03 frankfurt cloudflare dns ping",
				selectionValues: ["assignment-cloudflare-fra"]
			}
		]
	}
];

const meta = {
	title: "Forms/CategorizedMultiSelect",
	component: CategorizedMultiSelect,
	args: {
		label: "Assignments",
		placeholder: "Select assignments",
		valueLabel: "2 assignments selected",
		categories,
		selectedValues: ["assignment-cloudflare-tpe", "assignment-api-tpe"],
		searchPlaceholder: "Search assignments",
		selectAllLabel: "Select all visible",
		clearVisibleLabel: "Clear visible",
		emptyLabel: "No assignments match this search.",
		optionsAriaLabel: "Assignment options",
		categoriesAriaLabel: "Assignment browsing mode",
		selectItemAriaLabel: item => `Select ${String(item.label)}`,
		onValueChange: () => undefined
	},
	argTypes: {
		categories: { control: false },
		onValueChange: { control: false },
		selectItemAriaLabel: { control: false }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof CategorizedMultiSelect>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: function Render(args) {
		const [selectedValues, setSelectedValues] = useState(args.selectedValues);

		return (
			<div className="storybook-canvas">
				<div className="storybook-demo--narrow">
					<CategorizedMultiSelect {...args} selectedValues={selectedValues} valueLabel={`${selectedValues.length} assignments selected`} onValueChange={setSelectedValues} />
				</div>
			</div>
		);
	}
};

export const States: Story = {
	render: args => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				<CategorizedMultiSelect {...args} label="Empty" valueLabel="" selectedValues={[]} />
				<CategorizedMultiSelect {...args} label="Single selection" valueLabel="Cloudflare DNS / probe-tpe-01" selectedValues={["assignment-cloudflare-tpe"]} />
				<CategorizedMultiSelect {...args} label="Disabled" disabled />
			</div>
		</div>
	)
};
