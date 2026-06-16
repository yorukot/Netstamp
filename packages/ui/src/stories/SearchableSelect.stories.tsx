import type { Meta, StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import { FieldLabel, SearchableSelect, type SearchableSelectOption } from "../index";

const options: SearchableSelectOption[] = [
	{ value: "probe-tpe-01", label: "probe-tpe-01", description: "Taipei edge / online" },
	{ value: "probe-sfo-02", label: "probe-sfo-02", description: "San Francisco backbone / online" },
	{ value: "probe-fra-03", label: "probe-fra-03", description: "Frankfurt relay / degraded" },
	{ value: "probe-nrt-04", label: "probe-nrt-04", description: "Tokyo transit / pending" }
];

const meta = {
	title: "Forms/SearchableSelect",
	component: SearchableSelect,
	args: {
		"aria-label": "Select probe",
		onValueChange: () => undefined,
		options,
		placeholder: "Select probe",
		searchPlaceholder: "Search probes",
		size: "md",
		value: "probe-tpe-01"
	},
	argTypes: {
		onValueChange: { control: false },
		options: { control: false },
		size: { control: "inline-radio", options: ["sm", "md"] }
	},
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<typeof SearchableSelect>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	render: function Render(args) {
		const [value, setValue] = useState(args.value);

		return (
			<div className="storybook-canvas">
				<label className="storybook-compact-field storybook-demo--narrow">
					<FieldLabel>Probe</FieldLabel>
					<SearchableSelect {...args} value={value} onValueChange={setValue} />
				</label>
			</div>
		);
	}
};

export const States: Story = {
	render: () => (
		<div className="storybook-canvas">
			<div className="storybook-grid">
				<label className="storybook-compact-field">
					<FieldLabel>Default</FieldLabel>
					<SearchableSelect aria-label="Default probe" options={options} value="probe-sfo-02" onValueChange={() => undefined} />
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Invalid</FieldLabel>
					<SearchableSelect aria-label="Invalid probe" options={options} value="probe-fra-03" invalid onValueChange={() => undefined} />
				</label>
				<label className="storybook-compact-field">
					<FieldLabel>Disabled</FieldLabel>
					<SearchableSelect aria-label="Disabled probe" options={options} value="probe-nrt-04" disabled onValueChange={() => undefined} />
				</label>
			</div>
		</div>
	)
};
