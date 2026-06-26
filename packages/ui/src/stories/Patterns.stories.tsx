import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge, Button, CodePreview, DataTable, MetricTile, Panel, SelectableRow, SpecCard, SpecLabel, type DataColumn } from "../index";

interface CheckRow {
	id: string;
	check: string;
	target: string;
	state: "healthy" | "degraded" | "failed";
	p95: string;
	loss: string;
}

const rows: CheckRow[] = [
	{ id: "dns", check: "DNS resolver", target: "1.1.1.1", state: "healthy", p95: "33ms", loss: "0.00%" },
	{ id: "api", check: "API edge", target: "api.netstamp.dev", state: "degraded", p95: "184ms", loss: "0.18%" },
	{ id: "trace", check: "Route trace", target: "tpe -> sfo", state: "failed", p95: "--", loss: "100%" }
];

const stateTone: Record<CheckRow["state"], "success" | "warning" | "critical"> = {
	degraded: "warning",
	failed: "critical",
	healthy: "success"
};

const columns: DataColumn<CheckRow>[] = [
	{ key: "check", label: "Check", sortable: true },
	{ key: "target", label: "Target", sortable: true },
	{
		key: "state",
		label: "State",
		sortable: true,
		render: row => <Badge tone={stateTone[row.state]}>{row.state}</Badge>
	},
	{ key: "p95", label: "p95", sortable: true },
	{ key: "loss", label: "Loss", sortable: true }
];

const meta = {
	title: "Patterns/Operational workspace",
	parameters: {
		layout: "fullscreen"
	}
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const DashboardSpec: Story = {
	render: () => (
		<div className="storybook-page">
			<section className="storybook-section storybook-workspace">
				<div className="storybook-section-header">
					<SpecLabel tone="primary">Pattern</SpecLabel>
					<h2>Operational workspace</h2>
					<p>Compose shared primitives into dense app or docs surfaces without decorative containers.</p>
				</div>
				<div className="storybook-workspace-grid">
					<div className="storybook-workspace-main">
						<div className="storybook-grid">
							<MetricTile label="online probes" value="38" detail="healthy" tone="success" />
							<MetricTile label="warning checks" value="4" detail="watch" tone="warning" />
							<MetricTile label="failed routes" value="1" detail="critical" tone="critical" />
						</div>
						<Panel eyebrow="Checks" title="Network checks" summary="Primary actions stay orange; dense data stays neutral until state changes." actions={<Button size="sm">Create check</Button>}>
							<DataTable columns={columns} rows={rows} getRowKey={row => row.id} minWidth="42rem" ariaLabel="Workspace check state" density="compact" />
						</Panel>
						<CodePreview title="latest route diff" meta="trace">
							{`hop 04 changed: 203.69.35.1 -> 203.69.35.9
p95 latency +42ms
path hash b94c.22f9.changed`}
						</CodePreview>
					</div>
					<aside className="storybook-workspace-side">
						<SpecCard
							eyebrow="Next action"
							title="Acknowledge degraded API edge"
							description="Route p95 crossed alert threshold for 9 minutes."
							actions={
								<Button size="sm" variant="secondary">
									Review
								</Button>
							}
							active
						/>
						<div className="storybook-spec-list">
							<SelectableRow active tone="primary" title="API edge" description="p95 184ms" meta="warn" />
							<SelectableRow title="DNS resolver" description="p95 33ms" meta="ok" />
							<SelectableRow title="Route trace" description="packet loss 100%" meta="fail" />
						</div>
					</aside>
				</div>
			</section>
		</div>
	)
};
