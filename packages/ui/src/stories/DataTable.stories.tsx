import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge, Button, DataTable, type DataColumn, type DataTableProps } from "../index";

interface ProbeRow {
	id: string;
	latency: string;
	loss: string;
	probe: string;
	region: string;
	status: "online" | "degraded" | "offline";
}

const rows: ProbeRow[] = [
	{ id: "ams", latency: "42ms", loss: "0.00%", probe: "ams-edge-01", region: "Amsterdam", status: "online" },
	{ id: "tpe", latency: "118ms", loss: "0.21%", probe: "tpe-lab-02", region: "Taipei", status: "degraded" },
	{ id: "sfo", latency: "64ms", loss: "0.00%", probe: "sfo-core-03", region: "San Francisco", status: "online" },
	{ id: "gru", latency: "--", loss: "--", probe: "gru-edge-04", region: "Sao Paulo", status: "offline" }
];

const statusTone: Record<ProbeRow["status"], "critical" | "success" | "warning"> = {
	degraded: "warning",
	offline: "critical",
	online: "success"
};

const columns: DataColumn<ProbeRow>[] = [
	{ key: "probe", label: "Probe", sortable: true },
	{ key: "region", label: "Region", sortable: true },
	{
		key: "status",
		label: "Status",
		sortable: true,
		render: row => <Badge tone={statusTone[row.status]}>{row.status}</Badge>
	},
	{ key: "latency", label: "Latency", sortable: true, sortValue: row => Number.parseFloat(row.latency) },
	{ key: "loss", label: "Loss", sortable: true, sortValue: row => Number.parseFloat(row.loss) }
];

const meta = {
	title: "Components/DataTable",
	args: {
		ariaLabel: "Probe health examples",
		columns,
		density: "normal",
		getRowKey: row => row.id,
		minWidth: "44rem",
		rows
	},
	argTypes: {
		columns: { control: false },
		density: { control: "inline-radio", options: ["normal", "compact"] },
		emptyLabel: { control: "text" },
		getRowAriaLabel: { control: false },
		getRowKey: { control: false },
		onRowClick: { control: false },
		onSelectedRowKeysChange: { control: false },
		onSortChange: { control: false },
		rows: { control: false },
		selectedKey: { control: "text" },
		style: { control: false }
	},
	render: args => (
		<div className="storybook-canvas storybook-canvas--top">
			<div className="storybook-demo storybook-demo--wide">
				<DataTable<ProbeRow> {...args} />
			</div>
		</div>
	),
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta<DataTableProps<ProbeRow>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Compact: Story = {
	args: {
		density: "compact"
	}
};

export const Interactive: Story = {
	args: {
		batchActions: (
			<Button size="sm" variant="danger">
				Delete selected
			</Button>
		),
		batchLabel: (
			<>
				<strong>1 selected</strong>
				<span>tpe-lab-02</span>
			</>
		),
		getRowAriaLabel: row => `Open ${row.probe}`,
		onRowClick: () => undefined,
		onSelectedRowKeysChange: () => undefined,
		rowActions: row => (
			<Button size="sm" variant="secondary">
				Open {row.probe}
			</Button>
		),
		selectable: true,
		selectedRowKeys: ["tpe"],
		selectedKey: "tpe"
	}
};

export const Empty: Story = {
	args: {
		emptyLabel: "No probes match the current filters",
		rows: []
	}
};
